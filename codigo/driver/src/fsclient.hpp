#ifndef MIFS_FS_CLIENT_HPP
#define MIFS_FS_CLIENT_HPP

#include "expected.hpp"
#include "filemeta.hpp"
#include "http.hpp"
#include "httpc.hpp"
#include "jsend.hpp"
#include "tls.hpp"
#include "fscatalog.hpp"
#include <memory>

namespace mifs::apiclients {

class FileServerClient
{
    public:

    using http_client_ptr_t = std::shared_ptr<http::Client>;

    using list_response_t = jsend::Response<models::FileMetadata>;
    using list_response_result_t = util::Expected<list_response_t, int /* TODO */>;

    using contents_response_t = std::string;
    using contents_response_result_t = util::Expected<contents_response_t, int /* TODO */>;

    using no_response_t = util::Unexpected<int /* TODO */>;


    FileServerClient() = delete;
    FileServerClient(const FileServerClient&) = delete;
    FileServerClient(FileServerClient&&) = default;
    FileServerClient& operator=(const FileServerClient&) = delete;
    FileServerClient& operator=(FileServerClient&&) = delete;
    ~FileServerClient() = default;

    explicit FileServerClient(http_client_ptr_t http_client, util::FileServerCatalog::ptr_t fs_catalog);

    list_response_result_t get_all(std::string_view org, std::string_view server_name);
    contents_response_result_t contents(const std::string& org, const std::string& server_id, const std::string file_id);

    private:
    http_client_ptr_t client_;
    util::FileServerCatalog::ptr_t fs_catalog;
};


} // namespace mifs::apiclients

#endif // MIFS_FS_CLIENT_HPP
