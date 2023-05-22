#include "isclient.hpp"
#include "expected.hpp"
#include "http.hpp"
#include "jsend.hpp"
#include "mappings.hpp"
#include <fmt/format.h>
#include <iostream>
#include <rapidjson/stringbuffer.h>
#include <rapidjson/writer.h>

namespace mifs::apiclients
{

IndexServerClient::IndexServerClient(http_client_ptr_t http_client, Config config)
    : client_{std::move(http_client)},
      token_source_{std::move(config.token_source)},
      url_{std::move(config.url)},
      cacert_fn_{std::move(config.root_cert_fn)}
{
}

std::optional<Error> IndexServerClient::signin(std::string_view user, std::string_view email,
                                               std::string_view password)
{
    auto request{http::Request::Builder{}
                     .method(http::Method::GET)
                     .uri(fmt::format("{}/api/clients/v1/signup", url_))
                     .tls(tls::Config{cacert_fn_, "", ""})
                     .headers(http::Headers{{"Content-Type", "application/json"}})
                     .body("{" +
                           fmt::format("\"name\": \"{}\", \"email\": \"{}\", \"password\": \"{}\"", user,
                                       email, password) +
                           "}")
                     .build()};

    auto result{client_->execute(request)};
    if (!result) {
        return result.error();
    }

    auto code{result->code()};
    if (code != 200) {
        return Error{HTTPErrorCode{code}};
    }

    return std::nullopt; // todo validate possible failures when parsing
}

IndexServerClient::auth_result_t IndexServerClient::auth(std::string_view email, std::string_view password,
                                                         std::string_view otp)
{
    auto request{
        http::Request::Builder{}
            .method(http::Method::GET)
            .uri(fmt::format("{}/api/clients/v1/login", url_))
            .tls(tls::Config{cacert_fn_, "", ""})
            .headers(http::Headers{{"Content-Type", "application/json"}})
            .body("{" +
                  fmt::format(fmt::runtime("\"email\": \"{}\", \"password\": \"{}\", \"OTP\": \"{}\""), email,
                              password, otp) +
                  "}")
            .build()};

    auto result{client_->execute(request)};
    if (!result) {
        return error_t{result.error()};
    }

    auto code{result->code()};
    if (code != 200) {
        return error_t(Error(code));
    }

    return nsresp::TokenResponse::parse(result->body()); // todo validate possible failures when parsing
}

IndexServerClient::setup2fa_result_t IndexServerClient::setup2fa()
{
    auto token_res{token_source_->get()};
    if (!token_res) {
        return error_t{predefined::token_error};
    }

    auto request{http::Request::Builder{}
                     .method(http::Method::POST)
                     .uri(fmt::format("{}/api/clients/v1/2fa", url_))
                     .tls(tls::Config{cacert_fn_, "", ""})
                     .headers(http::Headers{{"X-MIFS-IS-Session-Token", *token_res}})
                     .build()};

    auto result{client_->execute(request)};
    if (!result) {
        return error_t{Error(result.error())};
    }

    auto code{result->code()};
    if (code != 200) {
        return error_t(Error(code));
    }

    return result->body();
}

std::optional<Error> IndexServerClient::link_fs(std::string organization, std::string server,
                                                std::string_view cert_fn, std::string_view key_fn, bool force)
{
    auto token_res{token_source_->get()};
    if (!token_res) {
        return predefined::token_error;
    }

    auto url{fmt::format("{}/api/clients/v1/organizations/{}/servers/{}/link", url_, organization, server)};
    if (force) {
        url += "?force=true";
    }

    auto request{http::Request::Builder{}
                     .method(http::Method::GET)
                     .uri(url)
                     .follow_redirects(true)
                     .tls(tls::Config{cacert_fn_, cert_fn, key_fn})
                     .headers(http::Headers{{"X-MIFS-IS-Session-Token", *token_res}})
                     .build()};

    auto result{client_->execute(request)};
    if (!result) {
        return Error(result.error());
    }

    auto code{result->code()};
    if (code != 200) {
        return Error(code);
    }

    return std::nullopt;
}

IndexServerClient::mappings_result_t IndexServerClient::get_mappings(bool forceFresh)
{
    auto token_res{token_source_->get()};
    if (!token_res) {
        return error_t{predefined::token_error};
    }

    auto request{http::Request::Builder{}
                     .method(http::Method::GET)
                     .uri(fmt::format("{}/api/clients/v1/mappings?forceUpdate={}", url_, forceFresh))
                     .tls(tls::Config{cacert_fn_, "", ""})
                     .headers(http::Headers{{"X-MIFS-IS-Session-Token", *token_res}})
                     .build()};

    auto result{client_->execute(request)};
    if (!result) {
        return error_t{Error(result.error())};
    }

    auto code{result->code()};
    if (code != 200) {
        auto res{jsend::parse_unsuccessful_response(result->body())};
        if (!res) {
            return error_t{predefined::json_error_unsuccessful};
        }
        return error_t{Error{code, jsend::format_error(*res)}};
    }

    auto response_res{jsend::parse_multi_item_response<models::Mapping>(result->body(), "mappings")};
    if (!response_res) {
        return error_t{predefined::json_error};
    }

    return mappings_result_t{std::move(*response_res)};
}

IndexServerClient::mapping_result_t IndexServerClient::create_mapping(const models::Mapping& m)
{
    auto token_res{token_source_->get()};
    if (!token_res) {
        return error_t{predefined::token_error};
    }

    // TODO(mredolatti): move this block to a separate function
    rapidjson::Document doc;
    m.serialize<rapidjson::Document>(doc, true); // TODO(mredolatti): check this func output!
    rapidjson::StringBuffer sb;
    rapidjson::Writer<rapidjson::StringBuffer> writer(sb);
    doc.Accept(writer);

    auto request{http::Request::Builder{}
                     .method(http::Method::POST)
                     .uri(fmt::format("{}/api/clients/v1/mappings", url_))
                     .tls(tls::Config{cacert_fn_, "", ""})
                     .headers(http::Headers{{"X-MIFS-IS-Session-Token", *token_res},
                                            {"Content-Type", "application/json"}})
                     .body(sb.GetString())
                     .build()};

    auto result{client_->execute(request)};
    if (!result) {
        return error_t{result.error()};
    }

    auto code{result->code()};
    if (code != 200) {
        auto res{jsend::parse_unsuccessful_response(result->body())};
        if (!res) {
            return error_t{predefined::json_error_unsuccessful};
        }
        return error_t{Error{code, jsend::format_error(*res)}};
    }

    auto response_res{jsend::parse_single_item_response<models::Mapping>(result->body(), "mapping")};
    if (!response_res) {
        return error_t{predefined::json_error};
    }

    return mapping_result_t{std::move(*response_res)};
}

IndexServerClient::mapping_result_t IndexServerClient::update_mapping(const models::Mapping& m)
{
    auto token_res{token_source_->get()};
    if (!token_res) {
        return error_t{predefined::token_error};
    }

    // TODO(mredolatti): move this block to a separate function
    rapidjson::Document doc;
    m.serialize<rapidjson::Document>(doc, true); // TODO(mredolatti): check this func output!
    rapidjson::StringBuffer sb;
    rapidjson::Writer<rapidjson::StringBuffer> writer(sb);
    doc.Accept(writer);

    auto request{http::Request::Builder{}
                     .method(http::Method::PUT)
                     .uri(fmt::format("{}/api/clients/v1/mappings/{}", url_, m.id()))
                     .tls(tls::Config{cacert_fn_, "", ""})
                     .headers(http::Headers{{"X-MIFS-IS-Session-Token", *token_res},
                                            {"Content-Type", "application/json"}})
                     .body(sb.GetString())
                     .build()};

    auto result{client_->execute(request)};
    if (!result) {
        return error_t{result.error()};
    }

    auto code{result->code()};
    if (code != 200) {
        auto res{jsend::parse_unsuccessful_response(result->body())};
        if (!res) {
            return error_t{predefined::json_error_unsuccessful};
        }
        return error_t{Error{code, jsend::format_error(*res)}};
    }

    auto response_res{jsend::parse_single_item_response<models::Mapping>(result->body(), "mapping")};
    if (!response_res) {
        return error_t{predefined::json_error};
    }

    return mapping_result_t{std::move(*response_res)};
}

std::optional<Error> IndexServerClient::delete_mapping(std::string_view mapping_id)
{
    auto token_res{token_source_->get()};
    if (!token_res) {
        return predefined::token_error;
    }

    auto request{http::Request::Builder{}
                     .method(http::Method::DELETE)
                     .uri(fmt::format("{}/api/clients/v1/mappings/{}", url_, mapping_id))
                     .tls(tls::Config{cacert_fn_, "", ""})
                     .headers(http::Headers{{"X-MIFS-IS-Session-Token", *token_res}})
                     .build()};

    auto result{client_->execute(request)};
    if (!result) {
        return Error{result.error()};
    }

    auto code{result->code()};
    if (code != 200) {
        auto res{jsend::parse_unsuccessful_response(result->body())};
        if (!res) {
            return predefined::json_error_unsuccessful;
        }
        return Error{code, jsend::format_error(*res)};
    }

    return std::nullopt;
}

IndexServerClient::servers_result_t IndexServerClient::get_servers()
{
    auto token_res{token_source_->get()};
    if (!token_res) {
        return error_t{predefined::token_error};
    }

    auto request{http::Request::Builder{}
                     .method(http::Method::GET)
                     .uri(fmt::format("{}/api/clients/v1/servers", url_))
                     .tls(tls::Config{cacert_fn_, "", ""})
                     .headers(http::Headers{{"X-MIFS-IS-Session-Token", *token_res}})
                     .build()};

    auto result{client_->execute(request)};
    if (!result) {
        return error_t{result.error()};
    }

    auto code{result->code()};
    if (code != 200) {
        auto res{jsend::parse_unsuccessful_response(result->body())};
        if (!res) {
            return error_t{predefined::json_error_unsuccessful};
        }
        return error_t{Error{code, jsend::format_error(*res)}};
    }

    auto response_res{jsend::parse_multi_item_response<models::FileServer>(result->body(), "servers")};
    if (!response_res) {
        return error_t{predefined::json_error};
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

    return Config{.url = cfg.index_server().url,
                  .root_cert_fn = cfg.index_server().root_cert,
                  .token_source = std::move(ts)};
}

} // namespace mifs::apiclients
