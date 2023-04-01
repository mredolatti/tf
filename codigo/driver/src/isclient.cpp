#include "isclient.hpp"
#include "http.hpp"
#include "jsend.hpp"
#include "expected.hpp"
#include "mappings.hpp"
#include <fmt/format.h>
#include <iostream>


namespace mifs::apiclients {

namespace detail {

using parse_result_t = util::Expected<IndexServerClient::mappings_response_t, int /* TODO */>;
using parse_error_t = util::Unexpected<int /* TODO */>;

}

IndexServerClient::IndexServerClient(http_client_ptr_t http_client, Config config) :
    client_{std::move(http_client)},
    token_source_{std::move(config.token_source)},
    url_{std::move(config.url)},
    cacert_fn_{std::move(config.root_cert_fn)}
{}

IndexServerClient::mappings_result_t IndexServerClient::get_mappings()
{
    auto token_res{token_source_->get()};
    if (!token_res) {
	    return no_response_t{-2};
    }

    auto request{http::Request::Builder{}
        .method(http::Method::GET)
        .uri(fmt::format("{}/api/clients/v1/mappings", url_))
	.tls(tls::Config{cacert_fn_, "", ""})
	.headers(http::Headers{{"X-MIFS-IS-Session-Token", *token_res}})
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

    auto response_res{jsend::parse<models::Mapping>((*result).body(), "mapping")};
    if (!response_res) {
	    return no_response_t{-2};
    }

    return mappings_result_t{std::move(*response_res)};
}

IndexServerClient::servers_result_t IndexServerClient::get_servers()
{
    auto token_res{token_source_->get()};
    if (!token_res) {
	    return no_response_t{-2};
    }

    auto request{http::Request::Builder{}
        .method(http::Method::GET)
        .uri(fmt::format("{}/api/clients/v1/servers", url_))
	.tls(tls::Config{cacert_fn_, "", ""})
	.headers(http::Headers{{"X-MIFS-IS-Session-Token", *token_res}})
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

    auto response_res{jsend::parse<models::FileServer>((*result).body(), "server")};
    if (!response_res) {
	    return no_response_t{-2};
    }

    return servers_response_t{*response_res};
}

IndexServerClient::Config IndexServerClient::Config::from_parsed_conf(const mifs::Config& cfg)
{

    token_source_ptr_t ts;
    const auto& ts_data{cfg.index_server().token_source_};
    if (ts_data.size() > 5 && ts_data.rfind("env::", 0) == 0) {
        ts = std::make_unique<EnvISTokenSource>(ts_data.substr(5));
    }

    return Config{
        .url = cfg.index_server().url,
        .root_cert_fn = cfg.index_server().root_cert,
        .token_source = std::move(ts)
    };
}


} // namespace mifs::apiclients
