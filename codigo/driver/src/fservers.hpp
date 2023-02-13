#ifndef FILE_SERVERS_HPP
#define FILE_SERVERS_HPP

#include "expected.hpp"
#include "mappings.hpp"
#include <mutex>
#include <string>
#include <unordered_map>
#include <vector>

namespace mifs::util {


class ServerFile
{
    public:
    ServerFile() = delete;
    ServerFile(const ServerFile&) = default;
    ServerFile(ServerFile&&) noexcept = default;
    ServerFile& operator=(const ServerFile&) = default;
    ServerFile& operator=(ServerFile&&) noexcept = default;
    ~ServerFile() noexcept = default;

    ServerFile(std::string name, std::size_t size_bytes, bool is_folder);
    const std::string& name() const;
    std::size_t size_bytes() const;
    bool is_folder() const;

    private:
    std::string name_;
    std::size_t size_bytes_;
    bool is_folder_;
};


class FileServersContents
{
    public:
    using servers_t = std::unordered_map<std::string, std::vector<ServerFile>>;

    enum class Error {
        Ok = 0,
        NotFound
    };

    using list_result_t = util::Expected<std::vector<ServerFile>, Error>;

    FileServersContents() = default;
    FileServersContents(const FileServersContents&) = delete;
    FileServersContents(FileServersContents&&) = delete;
    FileServersContents& operator=(const FileServersContents&) = delete;
    FileServersContents& operator=(FileServersContents&&) = delete;
    ~FileServersContents() = default;

    list_result_t ls(std::string_view path);
    void reset(const std::vector<models::Mapping>& mappings);
    
    private:
    std::mutex mutex_;
    servers_t servers_;

};

} // namespace mifs::util

#endif // FILE_SERVERS_HPP 
