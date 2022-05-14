#include "mappings.hpp"
#include "expected.hpp"

#include <rapidjson/document.h>

namespace mifs::models {

Mapping::Mapping(std::string_view name, std::size_t size_bytes, std::string_view ref, std::string_view server) :
    name_{name},
    size_bytes_{size_bytes},
    ref_{ref},
    server_{server}
{}

const std::string& Mapping::name() const
{
    return name_;
}

const std::string& Mapping::ref() const
{
    return ref_;
}

const std::string& Mapping::server() const
{
    return server_;
}


std::size_t Mapping::size_bytes() const
{
    return size_bytes_;
}

template<>
Mapping::parse_result_t Mapping::parse<rapidjson::Document::ValueType>(const rapidjson::Document::ValueType& doc)
{

    /*
        "userId"
        "serverId"
        "path"
        "ref"
        "updated"
        "deleted"
     */

    if (!doc.HasMember("path") || !doc["path"].IsString()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("path") || !doc["ref"].IsString()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("path") || !doc["serverId"].IsString()) {
        return util::Unexpected<int>{1};
    }

    return Mapping{
        doc["path"].GetString(),
        0, // TODO(mredolatti): send size in mapping
        doc["ref"].GetString(),
        doc["serverId"].GetString()
    };
}


} // namespace mifs::models
