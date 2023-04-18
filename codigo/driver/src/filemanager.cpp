#include "filemanager.hpp"

#include "expected.hpp"
#include "filemeta.hpp"
#include "isclient.hpp"
#include "log.hpp"
#include "mappings.hpp"

#include <algorithm>
#include <filesystem>

namespace mifs
{

namespace helpers
{
std::tuple<std::string, std::string, std::string> parse_server_ref(std::string_view path);
int write(std::string& document, const char *buffer, std::size_t size, off_t offset);
std::string_view trim_leading_slash(std::string_view path);
FileManager::Error map_fs_mirror_error(util::FSMirror::Error e);
bool is_server_path(std::filesystem::path path);
} // namespace helpers

FileManager::FileManager(apiclients::IndexServerClient is_client, apiclients::FileServerClient fs_client,
                         util::FileServerCatalog::ptr_t fs_catalog)
    : is_client_{std::move(is_client)},
      fs_client_{std::move(fs_client)},
      fs_catalog_{fs_catalog},
      logger_{log::get()}
{
}

FileManager::list_result_t FileManager::list(std::string_view path)
{
    path = helpers::trim_leading_slash(path);
    auto result{fs_mirror_.ls(path)};
    if (!result) {
        SPDLOG_LOGGER_ERROR(logger_, "failed to get folder listings: {}", result.error().message());
        return util::Unexpected<Error>{helpers::map_fs_mirror_error(result.error())};
    }
    return FileManager::list_result_t{*std::move(result)};
}

FileManager::stat_result_t FileManager::stat(std::string_view path)
{
    path = helpers::trim_leading_slash(path);
    auto result{fs_mirror_.info(path)};
    if (!result) {
        SPDLOG_LOGGER_ERROR(logger_, "failed to get item stats: {}", result.error().message());
        return util::Unexpected<Error>{helpers::map_fs_mirror_error(result.error())};
    }

    return FileManager::stat_result_t{*std::move(result)};
}

FileManager::Error FileManager::sync()
{
    SPDLOG_LOGGER_TRACE(logger_, "fetching mappings...");
    auto res_mappings{is_client_.get_mappings(true)};
    if (!res_mappings) {
        SPDLOG_LOGGER_ERROR(logger_, "failed to fetch mappings from index server: {}",
                            res_mappings.error().message());
        return Error::FailedToFetchMappings;
    }

    const auto& mappings{res_mappings->data["mappings"]};
    if (auto errors{fs_mirror_.reset_all(mappings)}; !errors.empty()) {
        for (const auto& err : errors) {
            const auto& mapping{mappings[err.first]};
            SPDLOG_LOGGER_ERROR(logger_, "error inserting mapping '{}/{}/{}' in path '{}': {}", mapping.org(),
                                mapping.server(), mapping.ref(), mapping.path(), err.second.message());
        }
    }

    SPDLOG_LOGGER_TRACE(logger_, "fetching file server information...");
    auto res_servers{is_client_.get_servers()};
    if (!res_servers) {
        SPDLOG_LOGGER_ERROR(logger_, "failed to fetch file servers information from index server: {}",
                            res_mappings.error().message());
        return Error::FailedToFetchServerInfos;
    }

    const auto& servers{res_servers->data["servers"]};
    for (const auto& server : servers) {
        fs_catalog_->update_fetch_url(server.org_name(), server.name(), server.fetch_url());
    }

    return Error::Ok;
}

FileManager::Error FileManager::touch(std::string_view path)
{
    using namespace std::chrono;
    path = helpers::trim_leading_slash(path);

    if (!helpers::is_server_path(path)) {
        SPDLOG_LOGGER_ERROR(logger_, "files can ony be created in in servers path, not '{}'", path);
        return Error::CannotWriteInNonServerPath;
    }

    auto now{time_point_cast<seconds>(system_clock::now()).time_since_epoch().count()};
    auto [org, server, ref]{helpers::parse_server_ref(path)};
    if (auto err{
            fs_client_.touch(org, server, ref, models::FileMetadata{"", ref, 0, "", "", "", "", now, false})};
        err) {

        SPDLOG_LOGGER_ERROR(logger_, "error creating new file placeholder in server: {}", err->message());
        return Error::FiledToWriteFileInServer;
    }

    if (auto err{fs_mirror_.add_file(org, server, ref, 0, now)};
        err.code() != util::FSMirror::Error::Code::Ok) {

        SPDLOG_LOGGER_CRITICAL(logger_, "error updating internal tree structure: {}", err.message());
        return helpers::map_fs_mirror_error(err);
    }

    return Error::Ok;
}

int FileManager::open(std::string_view path, int mode)
{

    path = helpers::trim_leading_slash(path);
    return open_files_.open(path, mode);
}

std::pair<int, FileManager::Error> FileManager::read(std::string_view path, char *buffer, std::size_t offset,
                                                     std::size_t count)
{
    path =  helpers::trim_leading_slash(path);
    auto gen_info{fs_mirror_.info(path)};
    if (!gen_info) {
        SPDLOG_LOGGER_ERROR(logger_, "error fetching info for path '{}': {}", path,
                            gen_info.error().message());
        return std::make_pair(0, FileManager::Error::NotFound);
    }

    const auto *file_meta{gen_info->file()};
    if (!file_meta) {
        SPDLOG_LOGGER_ERROR(logger_, "attempted to read from a non-file '{}'", path);
        return std::make_pair(0, Error::NotAFile);
    }

    SPDLOG_LOGGER_TRACE(logger_, "fetching contents for org={}, server={}, id={}",
                        file_meta->organization_name, file_meta->server_name, file_meta->ref);

    if (auto err{ensure_cached(file_meta->organization_name, file_meta->server_name, file_meta->ref)};
        err != Error::Ok) {

        SPDLOG_LOGGER_ERROR(logger_, "file '{}' could not be fetched/cached.", path);
        return std::make_pair(0, err);
    }

    auto from_cache{file_cache_.get(file_meta->organization_name, file_meta->server_name, file_meta->ref)};
    if (!from_cache) {
        SPDLOG_LOGGER_ERROR(
            logger_, "Invalid state. File '{}' was successfully fetched from server but not found in cache.",
            path);
        return std::make_pair(0, Error::FiledToReadFileFromServer);
    }

    const auto& contents{from_cache->get().contents()};
    std::size_t read_bytes{};
    while ((offset + read_bytes) < contents.size() && count > 0) {
        buffer[read_bytes] = contents[offset + read_bytes];
        read_bytes++;
        count--;
    }

    return std::make_pair(read_bytes, Error::Ok);
}

std::pair<int, FileManager::Error> FileManager::read(int fd, char *buffer, off_t offset, std::size_t count)
{
    auto of{open_files_.get(fd)};
    if (!of) {
        SPDLOG_LOGGER_ERROR(logger_, "read() requested for an FD ({}) not found in open file cache", fd);
        return std::make_pair(0, Error::NotFound);
    }

    return read(std::string{of->get().name}, buffer, offset, count);
}

std::pair<int, FileManager::Error> FileManager::write(std::string_view path, const char *buf, size_t size,
                                                      off_t offset)
{

    path = helpers::trim_leading_slash(path);
    if (!helpers::is_server_path(path)) {
        SPDLOG_LOGGER_ERROR(logger_, "cannot write in non-server-path '{}'", path);
        return std::make_pair(0, Error::CannotWriteInNonServerPath);
    }

    auto gen_info{fs_mirror_.info(path)};
    if (!gen_info) {
        SPDLOG_LOGGER_ERROR(logger_, "error fetching info for item '{}': {}", path,
                            gen_info.error().message());
        return std::make_pair(0, Error::NotFound);
    }

    const auto *file_meta{gen_info->file()};
    if (!file_meta) {
        SPDLOG_LOGGER_ERROR(logger_, "cannot write in non-file path '{}'", path);
        return std::make_pair(0, Error::NotAFile);
    }

    SPDLOG_LOGGER_TRACE(logger_, "fetching contents for org={}, server={}, id={}",
                        file_meta->organization_name, file_meta->server_name, file_meta->ref);

    if (auto err{ensure_cached(file_meta->organization_name, file_meta->server_name, file_meta->ref)};
        err != Error::Ok) {

        SPDLOG_LOGGER_ERROR(logger_, "file '{}' could not be fetched/cached.", path);
        return std::make_pair(0, err);
    }

    auto from_cache{file_cache_.get(file_meta->organization_name, file_meta->server_name, file_meta->ref)};
    if (!from_cache) {
        SPDLOG_LOGGER_ERROR(
            logger_, "Invalid state. File '{}' was successfully fetched from server but not found in cache.",
            path);
        return std::make_pair(0, Error::FiledToReadFileFromServer);
    }

    return std::make_pair(from_cache->get().write(buf, size, offset), Error::Ok);
}

FileManager::Error FileManager::ensure_cached(const std::string& org, const std::string& server,
                                              const std::string& ref)
{
    if (file_cache_.has(org, server, ref)) {
        return Error::Ok;
    }

    SPDLOG_LOGGER_TRACE(logger_, "fetching contents for file '{}' on server '{}'", ref, server);

    auto result{fs_client_.contents(org, server, ref)};
    if (!result) {
        SPDLOG_LOGGER_ERROR(logger_, "file '{}/{}/{}' could not be fetched/cached.", org, server, ref);
        return Error::FiledToReadFileFromServer;
    }

    return file_cache_.put(org, std::move(server), std::move(ref), std::move(*result))
             ? Error::Ok
             : Error::InternalCacheError;
}

FileManager::Error FileManager::link(std::string_view from, std::string_view to)
{
    from = helpers::trim_leading_slash(from);
    to = helpers::trim_leading_slash(to);

    if (!helpers::is_server_path(from)) {
        SPDLOG_LOGGER_ERROR(logger_, "link source is not a server path", from);
        return Error::InvalidLinkSource;
    }

    if (helpers::is_server_path(to)) {
        SPDLOG_LOGGER_ERROR(logger_, "link destinattion cannot be a server path", from);
        return Error::InvalidLinkDestination;
    }

    auto [org, server, ref]{helpers::parse_server_ref(from)};

    auto res{is_client_.create_mapping(models::Mapping{"", to, org, server, ref, 0, 0})};
    if (!res) {
        SPDLOG_LOGGER_ERROR(logger_, "faile to create mapping '{}' for file '{}/{}/{}' in index-server: {}",
                            to, org, server, ref, res.error().message());
        return Error::FiledToUpdateRemoteMapping;
    }

    const auto it{res->data.find("mapping")};
    assert(it != res->data.cend());
    if (auto err{fs_mirror_.link_file(it->second.id(), org, server, ref, to)};
        err.code() != util::FSMirror::Error::Code::Ok) {

        SPDLOG_LOGGER_ERROR(logger_, "failed to update internal structure when linking {}: {}", to,
                            err.message());
        return helpers::map_fs_mirror_error(err);
    }

    return Error::Ok;
}

FileManager::Error FileManager::flush(std::string_view path)
{
    path = helpers::trim_leading_slash(path);
    auto gen_info{fs_mirror_.info(path)};
    if (!gen_info) {
        SPDLOG_LOGGER_ERROR(logger_, "error fetching stats for '{}': {}", path, gen_info.error().message());
        return Error::NotFound;
    }

    const auto *file_meta{gen_info->file()};
    if (!file_meta) {
        SPDLOG_LOGGER_ERROR(logger_, "cannot flush non-file '{}'", path);
        return Error::NotAFile;
    }

    auto cache_entry_res{
        file_cache_.get(file_meta->organization_name, file_meta->server_name, file_meta->ref)};
    if (!cache_entry_res) {
        SPDLOG_LOGGER_TRACE(logger_, "file '{}/{}/{}' not present on cache. Nothing to do.",
                            file_meta->organization_name, file_meta->server_name, file_meta->ref);
        return Error::Ok;
    }

    auto& cache_entry{cache_entry_res->get()};
    if (!cache_entry.dirty()) {
        SPDLOG_LOGGER_TRACE(logger_, "file '{}/{}/{}' is not dirty. Nothing to do.",
                            file_meta->organization_name, file_meta->server_name, file_meta->ref);
        return Error::Ok;
    }

    if (auto err{fs_client_.update_contents(file_meta->organization_name, file_meta->server_name,
                                            file_meta->ref, cache_entry.contents())};
        err) {
        SPDLOG_LOGGER_ERROR(logger_, "file '{}/{}/{}' was not properly flushed: {}",
                            file_meta->organization_name, file_meta->server_name, file_meta->ref,
                            err->message());
        return Error::FiledToWriteFileInServer;
    }

    file_cache_.drop(file_meta->organization_name, file_meta->server_name, file_meta->ref);

    // TODO(mredolatti): Remove this and update FSMirror manually...
    sync();
    return Error::Ok;
}

FileManager::Error FileManager::mkdir(std::string_view path)
{
    path = helpers::trim_leading_slash(path);
    if (helpers::is_server_path(path)) {
        SPDLOG_LOGGER_ERROR(logger_, "cannot update folder structure of servers path");
        return Error::ServerTreeManipulation;
    }

    if (auto err{fs_mirror_.mkdir(path)};
        err.code() != util::FSMirror::Error::Code::Ok) {

        SPDLOG_LOGGER_CRITICAL(logger_, "failed to update internal tree representation of filesystem");
        return helpers::map_fs_mirror_error(err);
    }

    return Error::Ok;
}

FileManager::Error FileManager::rmdir(std::string_view path)
{
    path = helpers::trim_leading_slash(path);
    if (helpers::is_server_path(path)) {
        SPDLOG_LOGGER_ERROR(logger_, "cannot update folder structure of servers path");
        return Error::ServerTreeManipulation;
    }

    if (auto err{fs_mirror_.rmdir(path)};
        err.code() != util::FSMirror::Error::Code::Ok) {

        SPDLOG_LOGGER_CRITICAL(logger_, "failed to update internal tree representation of filesystem");
        return helpers::map_fs_mirror_error(err);
    }

    return Error::Ok;
}

FileManager::Error FileManager::remove(std::string_view path)
{
    path = helpers::trim_leading_slash(path);

    if (helpers::is_server_path(path)) {
        SPDLOG_LOGGER_ERROR(logger_, "cannot update folder structure of servers path");
        return Error::ServerTreeManipulation;
    }

    auto current_res{fs_mirror_.info(std::filesystem::path{path})};
    if (!current_res) {
        SPDLOG_LOGGER_ERROR(logger_, "error fetching stats for item '{}': {}", path,
                            current_res.error().message());
        return Error::NotFound;
    }

    const auto *as_link{current_res->link()};
    if (!as_link) { // it's not a link. cannot delete server files
        SPDLOG_LOGGER_ERROR(logger_, "cannot delete '{}'. Not a link to a server file reference", path);
        return Error::NotALink;
    }

    if (auto err{is_client_.delete_mapping(as_link->id)}; err) {
        SPDLOG_LOGGER_ERROR(logger_, "failed to delete mapping from index server: {}", err->message());
        return Error::FiledToUpdateRemoteMapping;
    }

    return helpers::map_fs_mirror_error(fs_mirror_.remove(std::filesystem::path{path}));
}

FileManager::Error FileManager::rename(std::string_view from, std::string_view to)
{
    from = helpers::trim_leading_slash(from);
    to = helpers::trim_leading_slash(to);

    if (helpers::is_server_path(from)) {
        SPDLOG_LOGGER_ERROR(logger_, "cannot rename server files");
        return Error::InvalidLinkSource;
    }

    if (helpers::is_server_path(to)) {
        SPDLOG_LOGGER_ERROR(logger_, "cannot rename server files");
        return Error::InvalidLinkDestination;
    }

    auto current_res{fs_mirror_.info(std::filesystem::path{from})};
    if (!current_res) {
        return Error::NotFound;
    }

    const auto *as_link{current_res->link()};
    if (!as_link) { // it's not a link. cannot rename server files
        SPDLOG_LOGGER_ERROR(logger_, "cannot rename a non-link fileystem-item '{}'");
        return Error::NotALink;
    }

    auto res{is_client_.update_mapping(models::Mapping{as_link->id, to, "", "", "", 0, 0})};
    if (!res) {
        SPDLOG_LOGGER_ERROR(logger_, "failed to update mapping: {}", res.error().message());
        return Error::FiledToUpdateRemoteMapping;
    }

    if (auto err{fs_mirror_.remove(from)}; err.code() != util::FSMirror::Error::Code::Ok) {
        SPDLOG_LOGGER_CRITICAL(logger_, "failed to update internal fileystem structure: {}", err.message());
        return helpers::map_fs_mirror_error(err.code());
    }

    const auto it{(*res).data.find("mapping")};
    assert(it != (*res).data.cend());
    return helpers::map_fs_mirror_error(fs_mirror_.link_file(it->second.id(), as_link->organization_name,
                                                             as_link->server_name, as_link->ref, to));
}

namespace helpers
{

bool is_server_path(std::filesystem::path path)
{
    auto begin{path.begin()};
    return begin != path.end() && *begin == "servers";
}

std::tuple<std::string, std::string, std::string> parse_server_ref(std::string_view path)
{
    std::filesystem::path p{path};
    auto ref{p.filename()};
    auto server{p.parent_path().filename()};
    auto org{p.parent_path().parent_path().filename()};
    return std::make_tuple(org.c_str(), server.c_str(), ref.c_str());
}

int write(std::string& document, const char *buffer, std::size_t size, off_t offset)
{
    if (auto newSize{offset + size}; newSize > document.size()) {
        document.reserve(size);
    }

    for (std::size_t i{0}; i < size; i++) {
        document[offset + i] = buffer[i];
    }

    return size;
}

std::string_view trim_leading_slash(std::string_view path)
{
    return (path.size() > 0 && path[0] == '/') ? path.substr(1) : path;
}

FileManager::Error map_fs_mirror_error(util::FSMirror::Error e)
{
    switch (e.code()) {
    case util::FSMirror::Error::Code::Ok: return FileManager::Error::Ok;
    case util::FSMirror::Error::Code::AlreadyExists: return FileManager::Error::AlreadyExists;
    case util::FSMirror::Error::Code::CannotLinkInServerFolder:
        return FileManager::Error::InvalidLinkDestination;
    case util::FSMirror::Error::Code::CannotAddInLinkedFolder:
        return FileManager::Error::CannotWriteInNonServerPath;
    case util::FSMirror::Error::Code::NotFound: return FileManager::Error::NotFound;
    }

    return FileManager::Error::Unknown;
}

} // namespace helpers

} // namespace mifs
