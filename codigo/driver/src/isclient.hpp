#ifndef MIFS_IS_CLIENT_HPP
#define MIFS_IS_CLIENT_HPP

#include "apierror.hpp"
#include "config.hpp"
#include "expected.hpp"
#include "http.hpp"
#include "httpc.hpp"
#include "istokens.hpp"
#include "jsend.hpp"
#include "mappings.hpp"
#include "nsresp.hpp"
#include "servers.hpp"

#include <memory>
#include <vector>

namespace mifs::apiclients
{

class IndexServerClient
{
  public:
    // dependencies
    using token_source_ptr_t = std::unique_ptr<IndexServerTokenSource>;
    using http_client_ptr_t = std::shared_ptr<http::Client>;

    // responses
    using mapping_response_t = jsend::SingleItemResponse<models::Mapping>;
    using mappings_list_response_t = jsend::MultipleItemResponse<models::Mapping>;
    using servers_response_t = jsend::MultipleItemResponse<models::FileServer>;

    // results
    using auth_result_t = util::Expected<nsresp::TokenResponse, Error>;
    using setup2fa_result_t = util::Expected<std::string, Error>;
    using mapping_result_t = util::Expected<mapping_response_t, Error>;
    using mappings_result_t = util::Expected<mappings_list_response_t, Error>;
    using servers_result_t = util::Expected<servers_response_t, Error>;
    using error_t = util::Unexpected<Error>;

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

    std::optional<Error> signin(std::string_view user, std::string_view email, std::string_view password);
    auth_result_t auth(std::string_view email, std::string_view password, std::string_view otp);
    setup2fa_result_t setup2fa();
    std::optional<Error> link_fs(std::string organization, std::string server, std::string_view cert_fn,
                                 std::string_view key_fn, bool force);
    mappings_result_t get_mappings(bool forceFresh);
    mapping_result_t create_mapping(const models::Mapping& m);
    mapping_result_t update_mapping(const models::Mapping& m);
    std::optional<Error> delete_mapping(std::string_view mapping_id);
    servers_result_t get_servers();

  private:
    http_client_ptr_t client_;
    token_source_ptr_t token_source_;
    std::string url_;
    std::string cacert_fn_;
};

} // namespace mifs::apiclients

#endif
