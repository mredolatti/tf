#include "filemanager.hpp"
#include "expected.hpp"
#include "filemeta.hpp"
#include "fselems.hpp"
#include "isclient.hpp"
#include "log.hpp"
#include "mappings.hpp"
#include <algorithm>

namespace mifs {

namespace helpers {
std::pair<std::string_view, std::string_view> parse_server_ref(std::string_view path);
}

FileManager::FileManager(apiclients::IndexServerClient is_client, apiclients::FileServerClient fs_client) :
    is_client_{std::move(is_client)},
    fs_client_{std::move(fs_client)},
    logger_{log::get()}
{}

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
    SPDLOG_LOGGER_TRACE(logger_, "fetching from file servers...");
    std::string servers[]{"fs1"}; // TODO(mredolatti)
    std::unordered_map<std::string, std::vector<models::FileMetadata>> files_by_server;
    for (const auto& server : servers) {
        auto res{fs_client_.get_all(server)};
        if (res) {
            auto&& files{(*res).data["files"]};
            SPDLOG_LOGGER_INFO(logger_, "successfully fetched {} files from {}", files.size(), server);
            files_by_server[server] = std::move(files);
        } else {
            SPDLOG_LOGGER_ERROR(logger_, "failed to fetch files from file server server {}: {}", server, res.error());
        }
    }

    SPDLOG_LOGGER_TRACE(logger_, "fetching from index server...");
    auto res{is_client_.get_all()};
    if (!res) {
        SPDLOG_LOGGER_ERROR(logger_, "failed to fetch mappings from index server: {}", res.error());
        return;
    }

    auto&& mappings{(*res).data["mapping"]};
    fs_mirror_.reset_all(std::move(mappings), std::move(files_by_server));
}


int FileManager::open(std::string_view path, int mode)
{
    (void)mode; // TODO(mredolatti): por ahora solo read-only
    return open_files_.open(path);
}

int FileManager::read(std::string_view path, char *buffer, std::size_t offset, std::size_t count)
{
    auto gen_info{fs_mirror_.info((path.size() > 0 && path[0] == '/') ? path.substr(1) : path)};
    if (!gen_info) {
        SPDLOG_LOGGER_ERROR(logger_, "file '{}' not found", path);
        return -1;
    }

    const auto* file_meta{dynamic_cast<types::FSEFile*>((*gen_info).get())};
    if (!file_meta) {
        SPDLOG_LOGGER_ERROR(logger_, "invalid item returned.");
        return -1;
    }

    SPDLOG_LOGGER_TRACE(logger_, "fetching contents for server={}, id={}", file_meta->server_id(), file_meta->ref());

    if (!ensure_cached(file_meta->server_id(), file_meta->ref())) {
        SPDLOG_LOGGER_ERROR(logger_, "file '{}' could not be fetched/cached.", path);
        return -1;
    }

    auto from_cache{file_cache_.get(file_meta->server_id(), file_meta->ref())};
    if (!from_cache) {
        return -1;
    }

    const auto& contents{(*from_cache).get().contents()};
    std::size_t read_bytes{};
    while ((offset + read_bytes) < contents.size() && count > 0) {
        buffer[read_bytes] = contents[offset + read_bytes];
        read_bytes++;
        count--;
    }

    return static_cast<int>(read_bytes);
}

int FileManager::read(int fd, char *buffer, std::size_t offset, std::size_t count)
{
    auto of{open_files_.get(fd)};
    if (!of) {
        return -1; // TODO(mredolatti): devolver error posta
    }

    return read(std::string{of->get().name}, buffer, offset, count);
}

bool FileManager::ensure_cached(std::string server, std::string ref)
{
    if (file_cache_.has(server, ref)) {
        return true;
    }

    SPDLOG_LOGGER_TRACE(logger_, "fetching contents for file '{}' on server '{}'", ref, server);

    auto result{fs_client_.contents(server, ref)};
    if (!result) {
        return false;
    }

    return file_cache_.put(std::move(server), std::move(ref), std::move(*result));
}

namespace helpers {
std::pair<std::string_view, std::string_view> parse_server_ref(std::string_view path)
{
    if (path.size() < 10 || path.substr(0, 9) != "/servers/") {
        return {};
    }

    path = path.substr(9);
    auto first_slash{path.find_first_of('/')};
    return {path.substr(0, first_slash), path.substr(first_slash+1)};
}
}
} // namespace mifs
