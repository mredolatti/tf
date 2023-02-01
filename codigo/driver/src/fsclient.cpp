#include "fsclient.hpp"
#include "filemeta.hpp"
#include <fmt/format.h>

namespace mifs::apiclients {

namespace detail {

ServerInfo::ServerInfo(std::string server_id, std::string server_url, tls::Config tls_config) :
    server_id_{std::move(server_id)},
    server_url_{std::move(server_url)},
    tls_config_{std::move(tls_config)}
{}

const std::string& ServerInfo::server_id() const
{
    return server_id_;
}

const std::string& ServerInfo::server_url() const
{
    return server_url_;
}

const tls::Config& ServerInfo::tls_config() const
{
    return tls_config_;
}

} // namespace detail

//------------------------------------------------

FileServerClient::FileServerClient(http_client_ptr_t http_client, server_infos_t server_infos) :
    client_{std::move(http_client)},
    server_infos_{std::move(server_infos)}
{}

FileServerClient::list_response_result_t FileServerClient::get_all(const std::string& server_id)
{
    const auto it{server_infos_.find(server_id)};
    if (it == server_infos_.end()) {
        return no_response_t{-1};
    }

    auto request{http::Request::Builder{}
        .method(http::Method::GET)
        .uri(fmt::format("{}/files", it->second.server_url()))
        .tls(it->second.tls_config())
        .build()};

    auto result{client_->execute(request)};
    if (!result) {
        return no_response_t{-1};
    }

    auto code{(*result).code()};
    if (code != 200) {
        std::cout << "code: " << code << '\n';
        return no_response_t{code};
    }

    return jsend::parse<models::FileMetadata>((*result).body(), "files");
}

    
FileServerClient::contents_response_result_t FileServerClient::contents(const std::string& server_id, const std::string file_id)
{
    const auto it{server_infos_.find(server_id)};
    if (it == server_infos_.end()) {
        return no_response_t{-1};
    }

    auto request{http::Request::Builder{}
        .method(http::Method::GET)
        .uri(fmt::format("{}/files/{}/contents", it->second.server_url(), file_id))
        .tls(it->second.tls_config())
        .build()};

    auto result{client_->execute(request)};
    if (!result) {
        return no_response_t{-1};
    }

    auto code{(*result).code()};
    if (code != 200) {
        std::cout << "code: " << code << '\n';
        return no_response_t{code};
    }

    return contents_response_result_t{(*result).body()};
}

} // namespace mifs::apiclients
