#include "mappings.hpp"
#include "expected.hpp"

#include <rapidjson/document.h>

namespace mifs::models {

Mapping::Mapping(std::string name, std::size_t size_bytes) :
    name_{std::move(name)},
    size_bytes_{size_bytes}
{}

const std::string& Mapping::name() const { return name_; }

std::size_t Mapping::size_bytes() const { return size_bytes_; }

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

    return Mapping{doc["path"].GetString(), 0};
}


} // namespace mifs::models
