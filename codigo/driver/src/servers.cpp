#include "servers.hpp"
#include "expected.hpp"

#include <rapidjson/document.h>

namespace mifs::models
{

FileServer::FileServer(std::string_view id, std::string_view org_id, std::string_view name,
                       std::string_view fetch_url)
    : id_{id},
      org_name_{org_id},
      name_{name},
      fetch_url_{fetch_url}
{
}

const std::string& FileServer::id() const { return id_; }

const std::string& FileServer::org_name() const { return org_name_; }

const std::string& FileServer::name() const { return name_; }

const std::string& FileServer::fetch_url() const { return fetch_url_; }

template <>
FileServer::parse_result_t
FileServer::parse<rapidjson::Document::ValueType>(const rapidjson::Document::ValueType& doc)
{

    /*
    ID                string `json:"id"`
    OrganizationName  string `json:"organizationName"`
    Name              string `json:"name"`
    AuthenticationURL string `json:"authenticationUrl"`
    TokenURL          string `json:"tokenUrl"`
    FileFetchURL      string `json:"fileFetchUrl"`
    ControlEndpoint   string `json:"controlEndpoint"`
*/

    if (!doc.HasMember("id") || !doc["id"].IsString()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("name") || !doc["name"].IsString()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("organizationName") || !doc["organizationName"].IsString()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("fileFetchUrl") || !doc["fileFetchUrl"].IsString()) {
        return util::Unexpected<int>{1};
    }

    return FileServer{
        doc["id"].GetString(),
        doc["organizationName"].GetString(),
        doc["name"].GetString(),
        doc["fileFetchUrl"].GetString(),
    };
}

} // namespace mifs::models
