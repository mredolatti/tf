#ifndef MIFS_FILE_CACHE_HPP
#define MIFS_FILE_CACHE_HPP

#include "expected.hpp"
#include "log.hpp"
#include <chrono>
#include <functional>
#include <mutex>
#include <string>
#include <unordered_map>

namespace mifs::util {

namespace detail {

class CacheEntry
{
    public:

    using sync_time_t = std::chrono::system_clock::time_point;

    CacheEntry() = delete;
    CacheEntry(const CacheEntry&) = default;
    CacheEntry(CacheEntry&&) noexcept = default;
    CacheEntry& operator=(const CacheEntry&) = default;
    CacheEntry& operator=(CacheEntry&&) noexcept = default;
    ~CacheEntry() noexcept = default;

    const std::string& file_id() const;
    const std::string& file_name() const;
    std::string& contents();
    std::chrono::system_clock::time_point last_sync() const;
    bool dirty() const;

    CacheEntry(std::string file_id, std::string file_name, std::string contents, sync_time_t last_sync);
    void update(std::string contents, sync_time_t last_sync);
    int write(std::string_view contents, std::size_t size, off_t offset);

    private:
    std::string file_id_;
    std::string file_name_;
    std::string contents_;
    sync_time_t last_sync_;
    bool dirty_;
};


} // namespace detail

class FileCache
{
    public:

    enum class Error
    {
        Ok = 0,
        NotFound,
        Expired
    };

    using get_res_t = util::Expected<std::reference_wrapper<detail::CacheEntry>, Error>;

    FileCache(const FileCache&) = delete;
    FileCache(FileCache&&) noexcept = delete;
    FileCache& operator=(const FileCache&) = delete;
    FileCache& operator=(FileCache&&) noexcept = delete;
    ~FileCache() noexcept = default;

    FileCache();
    get_res_t get(const std::string& org_name, const std::string& server_name, const std::string& ref);
    bool has(const std::string& org_name, const std::string& server_id, const std::string& ref);
    bool put(const std::string& org_name, std::string server_id, std::string ref, std::string contents);
    bool drop(const std::string& org_name, const std::string& server_id, const std::string& ref);

    private:
    using map_t = std::unordered_map<std::string, detail::CacheEntry>;
    map_t entries_;
    log::logger_t logger_;
    mutable std::mutex mutex_;
};


} // namespace mifs::utils
#endif // MIFS_FILE_CACHE_HPP
