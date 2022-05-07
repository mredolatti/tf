#include "filemanager.hpp"
#include "expected.hpp"
#include "isclient.hpp"
#include "log.hpp"
#include "mappings.hpp"

namespace mifs {

DirentryStub::DirentryStub(std::string name, size_t size, bool is_directory) :
    name_{std::move(name)},
    size_{size},
    is_directory_{is_directory}
{}

const std::string& DirentryStub::name() const { return name_; }
size_t DirentryStub::size() const { return size_; }
bool DirentryStub::is_directory() const { return is_directory_; }


namespace detail {

using mappings_t = std::vector<models::Mapping>;
FileManager::list_result_t build_direntries(std::string_view path, mappings_t&& mappings);

} // namespace detail


FileManager::FileManager(http_client_ptr_t http_client) :
    is_client_{apiclients::IndexServerClient{std::move(http_client)}},
    logger_{log::get()}
{}

FileManager::list_result_t FileManager::list(std::string_view path)
{
    auto res{is_client_.get_all()};
    if (!res) {
        SPDLOG_LOGGER_ERROR(logger_, "failed to fetch mappings from index server");
        return util::Unexpected<int>{-1};
    }

    return detail::build_direntries(path, std::move((*res).data["mapping"]));
}

FileManager::stat_result_t FileManager::stat(std::string_view path)
{
    auto last_slash{path.find_last_of('/')};
    if (last_slash == std::string_view::npos) {
        SPDLOG_LOGGER_ERROR(logger_, "invalid path '{}', missing '/'");
        return util::Unexpected<int>{-1};
    }

    // we need a slash at the beginning of the path, but not one at the end.
    // when splitting the parent & child we must account for 2 cases:
    // '/': this is the root folder and the only forward slash shold be kept
    // '/path/': this is a sub-folder and the last slash should be removed before building direntries
    auto parent{path.substr(0, (last_slash == 0) ? 1 : last_slash)};
    auto item{path.substr(last_slash+1)};
    SPDLOG_LOGGER_TRACE(logger_, "fetching stats for parent='{}', item='{}'", parent, item);

    auto res{list(parent)};
    if (!res) {
        return util::Unexpected<int>{-1};
    }

    for (auto direntry : (*res)) {
        SPDLOG_LOGGER_TRACE(logger_, "comparing direntry='{}' with item='{}'", direntry.name(), item);
        if (direntry.name() == item) {
            return direntry;
        }
    }

    SPDLOG_LOGGER_ERROR(logger_, "item not found when listing: '{}'", item);
    return util::Unexpected<int>{-1};
}


namespace detail {

FileManager::list_result_t build_direntries(std::string_view path, mappings_t&& mappings)
{
    auto logger{log::get()};
    SPDLOG_LOGGER_TRACE(logger, "building direntries for path='{}'", path);

    if (path.size() == 0 || path[0] != '/') {
        SPDLOG_LOGGER_ERROR(logger, "path '{}' is empty or doesn't start with a '/'");
        return util::Unexpected<int>{-1};
    }

    path = path.substr(1); // skip leading forward slash
    std::vector<DirentryStub> direntries;
    for (auto&& mapping : mappings) {
        std::string_view sv{mapping.name()};
        if (!path.empty()) {
            if (sv.find(path, 0) != 0) {
                continue; // not in path
            }

            sv = sv.substr(path.length() + 1); // remove path and parent forward slash
        }

        auto slash_idx{sv.find('/')};
        if (slash_idx == std::string_view::npos) { // it's a file!
            direntries.emplace_back(std::string{sv}, 123, false);
            continue;
        }

        // it's a directory
        sv = sv.substr(0, slash_idx); // remove sub_path
        direntries.emplace_back(std::string{sv}, 1, true);
    }

    // remove duplicates by sorting, moving consecutive duplicates to the end, and shrinking the vector
    std::sort(direntries.begin(), direntries.end(), [](auto&& x, auto&& y) { return x.name() < y.name(); });
    auto new_end{std::unique(direntries.begin(), direntries.end(), [](const auto& x, const auto& y){ return x.name() == y.name(); })};
    direntries.erase(new_end, direntries.end());

    return FileManager::list_result_t{std::move(direntries)};
}

} // namespace detail
} // namespace mifs
