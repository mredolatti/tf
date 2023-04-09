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
    auto result{fs_mirror_.ls((path.size() > 0 && path[0] == '/') ? path.substr(1) : path)};
    if (!result) {
        return util::Unexpected<int>{static_cast<int>(result.error())};
    }
    return FileManager::list_result_t{*std::move(result)};
}

FileManager::stat_result_t FileManager::stat(std::string_view path)
{
    auto result{fs_mirror_.info((path.size() > 0 && path[0] == '/') ? path.substr(1) : path)};
    if (!result) {
        return util::Unexpected<int>{static_cast<int>(result.error())};
    }

    return FileManager::stat_result_t{*std::move(result)};
}

void FileManager::sync()
{
    SPDLOG_LOGGER_TRACE(logger_, "fetching mappings...");
    auto res_mappings{is_client_.get_mappings(true)};
    if (!res_mappings) {
        SPDLOG_LOGGER_ERROR(logger_, "failed to fetch mappings from index server: {}", res_mappings.error());
        return;
    }

    const auto& mappings{res_mappings->data["mappings"]};
    fs_mirror_.reset_all(mappings);

    SPDLOG_LOGGER_TRACE(logger_, "fetching file server information...");
    auto res_servers{is_client_.get_servers()};
    if (!res_servers) {
        SPDLOG_LOGGER_ERROR(logger_, "failed to fetch file servers information from index server: {}",
                            res_mappings.error());
        return;
    }

    const auto& servers{res_servers->data["servers"]};
    for (const auto& server : servers) {
        fs_catalog_->update_fetch_url(server.org_name(), server.name(), server.fetch_url());
    }
}

int FileManager::touch(std::string_view path)
{
    using namespace std::chrono;
    auto now{time_point_cast<seconds>(system_clock::now()).time_since_epoch().count()};
    auto [org, server, ref]{helpers::parse_server_ref(path)};
    fmt::print("creando {} -- {} -- {}\n", org, server, ref);
    if (!fs_client_.touch(org, server, ref, models::FileMetadata{"", ref, 0, "", "", "", "", now, false})) {
        return 1;
    }

    fs_mirror_.add_file(org, server, ref, 0, now);
    return 0;
}

int FileManager::open(std::string_view path, int mode) { return open_files_.open(path, mode); }

int FileManager::read(std::string_view path, char *buffer, std::size_t offset, std::size_t count)
{
    auto gen_info{fs_mirror_.info((path.size() > 0 && path[0] == '/') ? path.substr(1) : path)};
    if (!gen_info) {
        SPDLOG_LOGGER_ERROR(logger_, "file '{}' not found", path);
        return -1;
    }

    const auto *file_meta{gen_info->file()};
    if (!file_meta) {
        SPDLOG_LOGGER_ERROR(logger_, "invalid item returned.");
        return -1;
    }

    SPDLOG_LOGGER_TRACE(logger_, "fetching contents for org={}, server={}, id={}",
                        file_meta->organization_name, file_meta->server_name, file_meta->ref);

    if (!ensure_cached(file_meta->organization_name, file_meta->server_name, file_meta->ref)) {
        SPDLOG_LOGGER_ERROR(logger_, "file '{}' could not be fetched/cached.", path);
        return -1;
    }

    auto from_cache{file_cache_.get(file_meta->organization_name, file_meta->server_name, file_meta->ref)};
    if (!from_cache) {
        return -1;
    }

    const auto& contents{from_cache->get().contents()};
    std::size_t read_bytes{};
    while ((offset + read_bytes) < contents.size() && count > 0) {
        buffer[read_bytes] = contents[offset + read_bytes];
        read_bytes++;
        count--;
    }

    return static_cast<int>(read_bytes);
}

int FileManager::read(int fd, char *buffer, off_t offset, std::size_t count)
{
    auto of{open_files_.get(fd)};
    if (!of) {
        return -1; // TODO(mredolatti): devolver error posta
    }

    return read(std::string{of->get().name}, buffer, offset, count);
}

int FileManager::write(std::string_view path, const char *buf, size_t size, off_t offset)
{
    auto gen_info{fs_mirror_.info((path.size() > 0 && path[0] == '/') ? path.substr(1) : path)};
    if (!gen_info) {
        SPDLOG_LOGGER_ERROR(logger_, "file '{}' not found", path);
        return -1;
    }

    const auto *file_meta{gen_info->file()};
    if (!file_meta) {
        SPDLOG_LOGGER_ERROR(logger_, "invalid item returned.");
        return -1;
    }

    SPDLOG_LOGGER_TRACE(logger_, "fetching contents for org={}, server={}, id={}",
                        file_meta->organization_name, file_meta->server_name, file_meta->ref);

    if (!ensure_cached(file_meta->organization_name, file_meta->server_name, file_meta->ref)) {
        SPDLOG_LOGGER_ERROR(logger_, "file '{}' could not be fetched/cached.", path);
        return -1;
    }

    auto from_cache{file_cache_.get(file_meta->organization_name, file_meta->server_name, file_meta->ref)};
    if (!from_cache) {
        return -1;
    }

    return from_cache->get().write(buf, size, offset);
}

bool FileManager::ensure_cached(const std::string& org, const std::string& server, const std::string& ref)
{
    if (file_cache_.has(org, server, ref)) {
        return true;
    }

    SPDLOG_LOGGER_TRACE(logger_, "fetching contents for file '{}' on server '{}'", ref, server);

    auto result{fs_client_.contents(org, server, ref)};
    if (!result) {
        return false;
    }

    return file_cache_.put(org, std::move(server), std::move(ref), std::move(*result));
}

bool FileManager::link(std::string_view from, std::string_view to)
{
    auto [org, server, ref]{helpers::parse_server_ref(from)};
    fmt::print("Link: org={}, server={}, ref={}", org, server, ref);

    auto res{is_client_.create_mapping(models::Mapping{"", to, org, server, ref, 0, 0})};
    if (!res) {
        return false;
    }

    const auto it{res->data.find("mapping")};
    assert(it != res->data.cend());
    return fs_mirror_.link_file(it->second.id(), org, server, ref, to) == util::FSMirror::Error::Ok;
}

bool FileManager::flush(std::string_view path)
{
    auto gen_info{fs_mirror_.info((path.size() > 0 && path[0] == '/') ? path.substr(1) : path)};
    if (!gen_info) {
        SPDLOG_LOGGER_ERROR(logger_, "file '{}' not found", path);
        return -1;
    }

    const auto *file_meta{gen_info->file()};
    if (!file_meta) {
        SPDLOG_LOGGER_ERROR(logger_, "invalid item returned.");
        return -1;
    }

    auto cache_entry_res{
        file_cache_.get(file_meta->organization_name, file_meta->server_name, file_meta->ref)};
    if (!cache_entry_res) {
        SPDLOG_LOGGER_TRACE(logger_, "file '{}/{}/{}' not present on cache. Nothing to do.",
                            file_meta->organization_name, file_meta->server_name, file_meta->ref);
        return true;
    }

    auto& cache_entry{cache_entry_res->get()};
    if (!cache_entry.dirty()) {
        SPDLOG_LOGGER_TRACE(logger_, "file '{}/{}/{}' is not dirty. Nothing to do.",
                            file_meta->organization_name, file_meta->server_name, file_meta->ref);
        return true;
    }

    auto res{fs_client_.update_contents(file_meta->organization_name, file_meta->server_name, file_meta->ref,
                                        cache_entry.contents())};
    if (!res) {
        SPDLOG_LOGGER_ERROR(logger_, "file '{}/{}/{}' was not properly flushed", file_meta->organization_name,
                            file_meta->server_name, file_meta->ref);
    }

    file_cache_.drop(file_meta->organization_name, file_meta->server_name, file_meta->ref);
    sync();
    return res;
}

bool FileManager::mkdir(std::string_view path)
{
    return fs_mirror_.mkdir(std::filesystem::path{path}) == util::FSMirror::Error::Ok;
}

bool FileManager::rmdir(std::string_view path)
{
    return fs_mirror_.rmdir(std::filesystem::path{path}) == util::FSMirror::Error::Ok;
}

bool FileManager::remove(std::string_view path)
{
    auto current_res{fs_mirror_.info(std::filesystem::path{path})};
    if (!current_res) {
        return false;
    }

    const auto *as_link{current_res->link()};
    if (!as_link) { // it's not a link. cannot delete server files
        return false;
    }

    // TODO: validate `to` is not in servers folder

    if (!is_client_.delete_mapping(as_link->id)) {
        return false;
    }

    return fs_mirror_.remove(std::filesystem::path{path}) == util::FSMirror::Error::Ok;
}

bool FileManager::rename(std::string_view from, std::string_view to)
{
    auto current_res{fs_mirror_.info(std::filesystem::path{from})};
    if (!current_res) {
        return false;
    }

    const auto *as_link{current_res->link()};
    if (!as_link) { // it's not a link. cannot rename server files
        return false;
    }

    // TODO: validate `to` is not in servers folder

    auto res{is_client_.update_mapping(models::Mapping{as_link->id, to, "", "", "", 0, 0})};
    if (!res) {
        return false;
    }

    if (fs_mirror_.remove(from) != util::FSMirror::Error::Ok) {
        // TODO(mredolatti): ???
    }

    const auto it{(*res).data.find("mapping")};
    assert(it != (*res).data.cend());
    return fs_mirror_.link_file(it->second.id(), as_link->organization_name, as_link->server_name,
                                as_link->ref, to) == util::FSMirror::Error::Ok;
}

namespace helpers
{

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

} // namespace helpers

} // namespace mifs
