#include "mappings.hpp"
#include "expected.hpp"

#include <rapidjson/document.h>

namespace mifs::models
{

Mapping::Mapping(std::string_view id, std::string_view path, std::string_view org, std::string_view server,
                 std::string_view ref, std::size_t size_bytes, int64_t last_updated)
    : id_{id},
      path_{path},
      size_bytes_{size_bytes},
      ref_{ref},
      org_{org},
      server_{server},
      last_updated_{last_updated}
{
}

const std::string& Mapping::id() const { return id_; }

const std::string& Mapping::path() const { return path_; }

const std::string& Mapping::ref() const { return ref_; }

const std::string& Mapping::org() const { return org_; }

const std::string& Mapping::server() const { return server_; }

std::size_t Mapping::size_bytes() const { return size_bytes_; }

int64_t Mapping::last_updated() const { return last_updated_; }

template <>
Mapping::parse_result_t
Mapping::parse<rapidjson::Document::ValueType>(const rapidjson::Document::ValueType& doc)
{
    if (!doc.HasMember("id") || !doc["id"].IsString()) {
        return util::Unexpected<int>{1};
    }

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

    return Mapping{doc["id"].GetString(),
                   doc.HasMember("path") ? doc["path"].GetString() : "",
                   doc["organizationName"].GetString(),
                   doc["serverName"].GetString(),
                   doc["ref"].GetString(),
                   static_cast<std::size_t>(doc["sizeBytes"].GetInt()),
                   doc["updated"].GetInt64()};
}

template <> int Mapping::serialize<rapidjson::Document>(rapidjson::Document& doc, bool ignore_empty) const
{
    doc.SetObject();
    auto& alloc{doc.GetAllocator()};

    if (!id_.empty() || !ignore_empty) {
        doc.AddMember("id", rapidjson::Value{id_.c_str(), alloc}.Move(), doc.GetAllocator());
    }

    if (!path_.empty() || !ignore_empty) {
        doc.AddMember("path", rapidjson::Value{path_.c_str(), alloc}.Move(), doc.GetAllocator());
    }

    if (!org_.empty() || !ignore_empty) {
        doc.AddMember("organizationName", rapidjson::Value{org_.c_str(), alloc}.Move(), doc.GetAllocator());
    }

    if (!server_.empty() || !ignore_empty) {
        doc.AddMember("serverName", rapidjson::Value{server_.c_str(), alloc}.Move(), doc.GetAllocator());
    }

    if (!ref_.empty() || !ignore_empty) {
        doc.AddMember("ref", rapidjson::Value{ref_.c_str(), alloc}.Move(), doc.GetAllocator());
    }

    if (last_updated_ != -1 || !ignore_empty) {
        doc.AddMember("updated", last_updated_, doc.GetAllocator());
    }

    if (size_bytes_ != -1 || !ignore_empty) {
        doc.AddMember("sizeBytes", size_bytes_, doc.GetAllocator());
    }

    return 0;
}

} // namespace mifs::models
