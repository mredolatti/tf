#include "mappings.hpp"
#include "expected.hpp"

#include <rapidjson/document.h>

namespace mifs::models {

Mapping::Mapping(std::string_view path, std::size_t size_bytes, std::string_view ref, std::string_view org, std::string_view server, int64_t last_updated) :
    path_{path},
    size_bytes_{size_bytes},
    ref_{ref},
    org_{org},
    server_{server},
    last_updated_{last_updated}
{}

const std::string& Mapping::path() const
{
    return path_;
}

const std::string& Mapping::ref() const
{
    return ref_;
}

const std::string& Mapping::org() const
{
    return org_;
}

const std::string& Mapping::server() const
{
    return server_;
}

std::size_t Mapping::size_bytes() const
{
    return size_bytes_;
}

int64_t Mapping::last_updated() const {
    return last_updated_;
}

template<>
Mapping::parse_result_t Mapping::parse<rapidjson::Document::ValueType>(const rapidjson::Document::ValueType& doc)
{
    if (doc.HasMember("path") && !doc["path"].IsString()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("ref") || !doc["ref"].IsString()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("organizationName") || !doc["organizationName"].IsString()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("serverName") || !doc["serverName"].IsString()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("sizeBytes") || !doc["sizeBytes"].IsInt()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("updated") || !doc["updated"].IsInt64()) {
        return util::Unexpected<int>{1};
    }

    return Mapping{
        doc.HasMember("path") ? doc["path"].GetString() : "",
        static_cast<std::size_t>(doc["sizeBytes"].GetInt()),
        doc["ref"].GetString(),
        doc["organizationName"].GetString(),
        doc["serverName"].GetString(),
	doc["updated"].GetInt64()
    };
}


} // namespace mifs::models
