#include "fscatalog.hpp"
#include <fmt/core.h>

namespace mifs::util {

std::string make_catalog_key(const std::string_view org, const std::string_view server);

ServerInfo::ServerInfo(std::string_view organization, std::string_view server, std::string_view files_url, tls::Config tls_config) :
    organization_{std::string{organization}},
    server_{std::string{server}},
    files_url_{std::string{files_url}},
    tls_config_{std::move(tls_config)}
{}
    
const std::string& ServerInfo::organization() const
{
	return organization_;
}

const std::string& ServerInfo::server() const
{
	return server_;
}

const std::string& ServerInfo::files_url() const
{
	return files_url_;
}

const tls::Config& ServerInfo::tls_config() const
{
	return tls_config_;
}


void ServerInfo::update_fetch_url(std::string_view new_url)
{
    files_url_ = new_url;
}


FileServerCatalog::FileServerCatalog(catalog_t catalog) :
    catalog_{std::move(catalog)}
{}

FileServerCatalog::ptr_t FileServerCatalog::createFromCredentialsConfig(const mifs::Config::fs_creds_by_org_t& creds_by_org)
{
    catalog_t catalog;
    for (const auto& for_org: creds_by_org) {
        for (const auto& for_server : for_org.second) {
            catalog.insert({make_catalog_key(for_org.first, for_server.first), ServerInfo{
                for_org.first,
                for_server.first,
                "",
                tls::Config{for_server.second.root_certificate_fn,
                    for_server.second.client_certificate_fn,
                    for_server.second.client_private_key_fn}}});
        };
    }
    return std::make_shared<FileServerCatalog>(std::move(catalog));
}

std::optional<ServerInfo> FileServerCatalog::get(std::string_view org, std::string_view server)
{
    std::lock_guard lk{mtx_};
    const auto it{catalog_.find(make_catalog_key(org, server))};
    return it != catalog_.cend() ? std::make_optional(it->second) : std::nullopt;
}

bool FileServerCatalog::update_fetch_url(std::string_view org, std::string_view server, std::string_view fetch_url)
{
    std::lock_guard lk{mtx_};
    const auto it{catalog_.find(make_catalog_key(org, server))};
    if (it == catalog_.cend()) {
        return false;
    }

    it->second.update_fetch_url(fetch_url);
    return true;
}


std::string make_catalog_key(std::string_view org, std::string_view server)
{
    return fmt::format("{}::{}", org, server);
}

}

