#include "filecache.hpp"
#include "expected.hpp"
#include <chrono>
#include <fmt/format.h>
#include <functional>
#include <system_error>

namespace mifs::util {

namespace detail {

std::string make_key(const std::string& org, const std::string& server, const std::string& ref)
{
	return fmt::format("{}/{}/{}", org, server, ref);
}

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

std::string& CacheEntry::contents()
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

int CacheEntry::write(std::string_view buffer, std::size_t size, off_t offset)
{
    contents_.resize(offset+size);
    for (std::size_t i{0}; i < size; i++) {
        contents_[offset + i] = buffer[i];
    }

    last_sync_ = std::chrono::system_clock::now();
    dirty_ = true;
    return size;
}


} // namespace detail

// -------------------------------------------

FileCache::FileCache() :
    logger_{log::get()}
{}


FileCache::get_res_t FileCache::get(const std::string& org_name, const std::string& server_id, const std::string& ref)
{
    std::lock_guard lk{mutex_};
    auto it{entries_.find(detail::make_key(org_name, server_id, ref))};
    if (it == entries_.end()) {
        return util::Unexpected<Error>{Error::NotFound};
    }

    return get_res_t{std::reference_wrapper<detail::CacheEntry>{it->second}};

}

bool FileCache::put(const std::string& org_name, std::string server_id, std::string ref, std::string contents)
{
    std::lock_guard lk{mutex_};
    auto key{detail::make_key(org_name, server_id, ref)};
    SPDLOG_LOGGER_TRACE(logger_, "storing entry {} in file cache", key);
    return entries_.insert({
            key,
            detail::CacheEntry{std::move(server_id), key, std::move(contents), std::chrono::system_clock::now()}
    }).second;
}

bool FileCache::has(const std::string& org_name, const std::string& server_id, const std::string& ref)
{
    std::lock_guard lk{mutex_};
    auto key{detail::make_key(org_name, server_id, ref)};
    return entries_.find(key) != entries_.end();
}

bool FileCache::drop(const std::string& org_name, const std::string& server_id, const std::string& ref)
{
    std::lock_guard lk{mutex_};
    auto key{detail::make_key(org_name, server_id, ref)};
    if (auto it{entries_.find(key)}; it != entries_.end()) {
        entries_.erase(it);
        return true;
    }
    return false;
}


} // namespace mifs::util
