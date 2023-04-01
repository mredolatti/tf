#include "fsclient.hpp"
#include "filemeta.hpp"
#include <fmt/format.h>

namespace mifs::apiclients {

FileServerClient::FileServerClient(http_client_ptr_t http_client, util::FileServerCatalog::ptr_t fs_catalog) :
    client_{std::move(http_client)},
    fs_catalog{std::move(fs_catalog)}
{}

FileServerClient::list_response_result_t FileServerClient::get_all(std::string_view org, std::string_view server_name)
{

    auto server_data{fs_catalog->get(org, server_name)};
    if (!server_data) {
        return no_response_t{-1};
    }

    auto request{http::Request::Builder{}
        .method(http::Method::GET)
        .uri(fmt::format("{}/files", server_data->files_url()))
        .tls(server_data->tls_config())
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

    auto response_res{jsend::parse<models::FileMetadata>((*result).body(), "files")};
    if (!response_res) {
	    return no_response_t{-2};
    }

    return list_response_result_t{*response_res};
}

    
FileServerClient::contents_response_result_t FileServerClient::contents(const std::string& org, const std::string& server_name, const std::string file_id)
{
    auto server_data{fs_catalog->get(org, server_name)};
    if (!server_data) {
        return no_response_t{-1};
    }

    auto request{http::Request::Builder{}
        .method(http::Method::GET)
        .uri(fmt::format("{}/{}/contents", server_data->files_url(), file_id))
        .tls(server_data->tls_config())
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
