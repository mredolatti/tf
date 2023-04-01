#ifndef MIFS_FILE_SERVER_CATALOG_HPP
#define MIFS_FILE_SERVER_CATALOG_HPP

#include "config.hpp"
#include "tls.hpp"
#include <memory>
#include <mutex>
#include <optional>
#include <string>
#include <unordered_map>

namespace mifs::util {

class ServerInfo
{
    public:
    ServerInfo() = delete;
    ServerInfo(const ServerInfo&) = default;
    ServerInfo(ServerInfo&&) noexcept = default;
    ServerInfo& operator=(const ServerInfo&) = default;
    ServerInfo& operator=(ServerInfo&&) noexcept = default;
    ~ServerInfo() noexcept = default;

    ServerInfo(std::string_view organization, std::string_view server_id, std::string_view server_url, tls::Config tls_config);

    const std::string& organization() const;
    const std::string& server() const;
    const std::string& files_url() const;
    const tls::Config& tls_config() const;

    void update_fetch_url(std::string_view new_url);

    private:
    std::string organization_;
    std::string server_;
    std::string files_url_;
    tls::Config tls_config_;
};


class FileServerCatalog
{
    public:

    using ptr_t = std::shared_ptr<FileServerCatalog>;
    using catalog_t = std::unordered_map<std::string, ServerInfo>;

    FileServerCatalog() = delete;
    FileServerCatalog(const FileServerCatalog&) = delete;
    FileServerCatalog(FileServerCatalog&&) noexcept = delete;
    FileServerCatalog& operator=(const FileServerCatalog&) = delete;
    FileServerCatalog& operator=(FileServerCatalog&&) noexcept = delete;
    ~FileServerCatalog() noexcept = default;

    explicit FileServerCatalog(catalog_t catalog);
    static ptr_t createFromCredentialsConfig(const mifs::Config::fs_creds_by_org_t&);

    std::optional<ServerInfo> get(std::string_view org, std::string_view server);
    bool update_fetch_url(std::string_view org, std::string_view server, std::string_view fetch_url);

    private:


    std::mutex mtx_;
    catalog_t catalog_;
};


}

#endif
