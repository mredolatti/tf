#include "config.hpp"
#include <document.h>
#include <fstream>
#include <iostream>

namespace mifs {

namespace detail {

constexpr const char* tokenSource = "tokenSource";
constexpr const char* credentials = "credentials";
constexpr const char* root_cert = "rootCertificate";
constexpr const char* client_cert = "clientCertificate";
constexpr const char* client_pk = "clientPrivateKey";

}

util::Expected<Config::is_cfg, Config::ParseError> parse_is_cfg(const rapidjson::Document& doc);
util::Expected<Config::fs_creds_by_org_t, Config::ParseError> parse_fs_creds(const rapidjson::Document::ValueType& doc);

    
Config::Config(is_cfg cfg, fs_creds_by_org_t creds) :
    index_server_config_{std::move(cfg)},
    creds_{std::move(creds)}
{}

Config::result_t Config::parse(const std::string& config_file_fn)
{

    std::ifstream t(config_file_fn);
    std::string raw((std::istreambuf_iterator<char>(t)),
                     std::istreambuf_iterator<char>());

    rapidjson::Document doc;
    doc.Parse(raw.c_str());

    auto is_conf_res{parse_is_cfg(doc)};
    if (!is_conf_res) return util::Unexpected<Config::ParseError>{is_conf_res.error()};

    auto fs_creds_res{parse_fs_creds(doc)};
    if (!fs_creds_res) return util::Unexpected<Config::ParseError>{is_conf_res.error()};

    return Config::result_t{Config{*is_conf_res, *fs_creds_res}};
}

const Config::is_cfg& Config::index_server() const
{
    return index_server_config_;
}

const Config::fs_creds_by_org_t& Config::creds() const
{
    return creds_;
}


util::Expected<Config::is_cfg, Config::ParseError> parse_is_cfg(const rapidjson::Document& doc)
{

    if (!doc.IsObject()) {
        std::cout << "es un " << doc.GetType() << '\n';
    }

    if (!doc.HasMember("indexServer") || !doc["indexServer"].IsObject()) {
        return util::Unexpected<Config::ParseError>{Config::ParseError::ErrorParsingJSON};
    }

    const auto& obj{doc["indexServer"].GetObject()};

    if (!obj.HasMember("tokenSource") || !obj["tokenSource"].IsString()) {
        return util::Unexpected<Config::ParseError>{Config::ParseError::ErrorParsingJSON};
    }

    if (!obj.HasMember("url") || !obj["url"].IsString()) {
        return util::Unexpected<Config::ParseError>{Config::ParseError::ErrorParsingJSON};
    }

    Config::is_cfg cfg;
    cfg.url = obj["url"].GetString();
    cfg.token_source_ = obj["tokenSource"].GetString();
    if (obj.HasMember("rootCert") && obj["rootCert"].IsString()) {
        cfg.root_cert = obj["rootCert"].GetString();
    }

    return util::Expected<Config::is_cfg, Config::ParseError>{cfg};
}

util::Expected<Config::fs_creds_by_org_t, Config::ParseError> parse_fs_creds(const rapidjson::Document::ValueType& doc)
{
    if (!doc.HasMember(detail::credentials) || !doc[detail::credentials].IsObject()) {
        return util::Unexpected<Config::ParseError>{Config::ParseError::ErrorParsingJSON};
    }

    const auto& by_org{doc[detail::credentials].GetObject()};
    Config::fs_creds_by_org_t servers_by_org;
    for (const auto& [org, servers] : by_org) {
        if (!org.IsString() || !servers.IsObject()) {
            return util::Unexpected<Config::ParseError>{Config::ParseError::ErrorParsingJSON};
        }

        const auto& servers_for_org{servers.GetObject()};
        Config::fs_creds_by_server_t creds_by_server;
        for (const auto& [server, creds] : servers_for_org) {
            if (!server.IsString() || !creds.IsObject()) {
                return util::Unexpected<Config::ParseError>{Config::ParseError::ErrorParsingJSON};
            }
            const auto& creds_obj{creds.GetObject()};
            
            if (creds.HasMember(detail::root_cert) && !creds[detail::root_cert].IsString()) {
                return util::Unexpected<Config::ParseError>{Config::ParseError::ErrorParsingJSON};
            }

            if (creds.HasMember(detail::client_cert) && !creds[detail::client_cert].IsString()) {
                return util::Unexpected<Config::ParseError>{Config::ParseError::ErrorParsingJSON};
            }

            if (creds.HasMember(detail::client_pk) && !creds.HasMember(detail::client_pk)) {
                return util::Unexpected<Config::ParseError>{Config::ParseError::ErrorParsingJSON};
            }

            Config::credentials cfs{
                .root_certificate_fn = creds.HasMember(detail::root_cert) ? creds[detail::root_cert].GetString() : "",
                .client_certificate_fn = creds.HasMember(detail::client_cert) ? creds[detail::client_cert].GetString() : "",
                .client_private_key_fn = creds.HasMember(detail::client_pk) ? creds[detail::client_pk].GetString() : ""
            };
            creds_by_server[server.GetString()] = std::move(cfs);
        }
        servers_by_org[org.GetString()] = std::move(creds_by_server);
    }
    return util::Expected<Config::fs_creds_by_org_t, Config::ParseError>{servers_by_org};
}

}
