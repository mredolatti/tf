#ifndef MIFS_FS_CLIENT_HPP
#define MIFS_FS_CLIENT_HPP

#include "expected.hpp"
#include "filemeta.hpp"
#include "http.hpp"
#include "httpc.hpp"
#include "jsend.hpp"
#include "tls.hpp"
#include <memory>

namespace mifs::apiclients {

namespace detail {


class ServerInfo
{
    public:
    ServerInfo() = delete;
    ServerInfo(const ServerInfo&) = default;
    ServerInfo(ServerInfo&&) noexcept = default;
    ServerInfo& operator=(const ServerInfo&) = default;
    ServerInfo& operator=(ServerInfo&&) noexcept = default;
    ~ServerInfo() noexcept = default;

    ServerInfo(std::string server_id, std::string server_url, tls::Config tls_config);

    const std::string& server_id() const;
    const std::string& server_url() const;
    const tls::Config& tls_config() const;

    private:
    std::string server_id_;
    std::string server_url_;
    tls::Config tls_config_;
};

}

class FileServerClient
{
    public:

    using http_client_ptr_t = std::shared_ptr<http::Client>;

    using list_response_t = jsend::Response<models::FileMetadata>;
    using list_response_result_t = util::Expected<list_response_t, int /* TODO */>;

    using contents_response_t = std::string;
    using contents_response_result_t = util::Expected<contents_response_t, int /* TODO */>;

    using no_response_t = util::Unexpected<int /* TODO */>;

    using server_infos_t = std::unordered_map<std::string, detail::ServerInfo>;

    FileServerClient() = delete;
    FileServerClient(const FileServerClient&) = delete;
    FileServerClient(FileServerClient&&) = default;
    FileServerClient& operator=(const FileServerClient&) = delete;
    FileServerClient& operator=(FileServerClient&&) = delete;
    ~FileServerClient() = default;

    explicit FileServerClient(http_client_ptr_t http_client, server_infos_t server_infos);

    list_response_result_t get_all(const std::string& server_id);
    contents_response_result_t contents(const std::string& server_id, const std::string file_id);

    private:
    http_client_ptr_t client_;
    server_infos_t server_infos_;
};


} // namespace mifs::apiclients

#endif // MIFS_FS_CLIENT_HPP
