#ifndef MIFS_IS_CLIENT_HPP
#define MIFS_IS_CLIENT_HPP

#include "config.hpp"
#include "expected.hpp"
#include "http.hpp"
#include "httpc.hpp"
#include "istokens.hpp"
#include "jsend.hpp"
#include "mappings.hpp"
#include "servers.hpp"

#include <memory>
#include <vector>

namespace mifs::apiclients
{

class IndexServerClient
{
  public:
    using token_source_ptr_t = std::unique_ptr<IndexServerTokenSource>;
    using http_client_ptr_t = std::shared_ptr<http::Client>;
    using mappings_response_t = jsend::Response<models::Mapping>;
    using mappings_result_t = util::Expected<mappings_response_t, int /* TODO */>;
    using servers_response_t = jsend::Response<models::FileServer>;
    using servers_result_t = util::Expected<servers_response_t, int /* TODO */>;
    using no_response_t = util::Unexpected<int /* TODO */>;

    struct Config {
        std::string url;
        std::string root_cert_fn;
        token_source_ptr_t token_source;

        static Config from_parsed_conf(const mifs::Config& cfg);
    };

    IndexServerClient() = delete;
    IndexServerClient(const IndexServerClient&) = delete;
    IndexServerClient(IndexServerClient&&) = default;
    IndexServerClient& operator=(const IndexServerClient&) = delete;
    IndexServerClient& operator=(IndexServerClient&&) = delete;
    ~IndexServerClient() = default;

    explicit IndexServerClient(http_client_ptr_t http_client, Config config);
    mappings_result_t get_mappings();
    int create_mapping(const models::Mapping& m);
    int update_mapping(const models::Mapping& m);
    int delete_mapping(std::string_view mapping_id);
    servers_result_t get_servers();

  private:
    http_client_ptr_t client_;
    token_source_ptr_t token_source_;
    std::string url_;
    std::string cacert_fn_;
};

} // namespace mifs::apiclients

#endif
