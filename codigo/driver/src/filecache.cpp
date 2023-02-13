#include "filecache.hpp"
#include "expected.hpp"
#include <chrono>
#include <fmt/format.h>
#include <functional>
#include <system_error>

namespace mifs::util {

namespace detail {

CacheEntry::CacheEntry(std::string file_id, std::string file_name, std::string contents, sync_time_t last_sync) :
    file_id_{std::move(file_id)},
    file_name_{std::move(file_name)},
    contents_{std::move(contents)},
    last_sync_{last_sync}
{}

const std::string& CacheEntry::file_id() const
{
    return file_id_;
}

const std::string& CacheEntry::file_name() const
{
    return file_name_;
}

const std::string& CacheEntry::contents() const
{
    return contents_;
}

std::chrono::system_clock::time_point CacheEntry::last_sync() const
{
    return last_sync_;
}

bool CacheEntry::dirty() const
{
    return dirty_;
}

void CacheEntry::update(std::string contents, sync_time_t last_sync)
{
    contents_ = std::move(contents);
    last_sync_ = last_sync;
}

} // namespace detail

// -------------------------------------------

FileCache::FileCache() :
    logger_{log::get()}
{}


FileCache::get_res_t FileCache::get(const std::string& server_id, const std::string& ref)
{
    std::lock_guard lk{mutex_};
    auto it{entries_.find(fmt::format("{}/{}", server_id, ref))};
    if (it == entries_.end()) {
        return util::Unexpected<Error>{Error::NotFound};
    }

    return get_res_t{std::reference_wrapper<detail::CacheEntry>{it->second}};

}

bool FileCache::put(std::string server_id, std::string ref, std::string contents)
{
    std::lock_guard lk{mutex_};
    auto entry{fmt::format("{}/{}", server_id, ref)};
    SPDLOG_LOGGER_TRACE(logger_, "storing entry {} in file cache", entry);
    return entries_.insert({
            entry,
            detail::CacheEntry{std::move(server_id), entry, std::move(contents), std::chrono::system_clock::now()}
    }).second;
}

bool FileCache::has(const std::string& server_id, const std::string& ref)
{
    std::lock_guard lk{mutex_};
    auto key{fmt::format("{}/{}", server_id, ref)};
    return entries_.find(key) != entries_.end();
}


} // namespace mifs::util
