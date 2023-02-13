#include "filemeta.hpp"
#include <fmt/format.h>
#include <rapidjson/document.h>

namespace mifs::models {

FileMetadata::FileMetadata(std::string id, std::string name, std::size_t size_bytes, std::string notes, std::string patient_id,
            std::string type, std::string content_id, int64_t last_updated, bool deleted) :
    id_{std::move(id)},
    name_{std::move(name)},
    size_bytes_{size_bytes},
    notes_{std::move(notes)},
    patient_id_{std::move(patient_id)},
    type_{std::move(type)},
    content_id_{std::move(content_id)},
    last_updated_{last_updated},
    deleted_{deleted}
{}

const std::string& FileMetadata::id() const
{
    return id_;
}

const std::string& FileMetadata::name() const
{
    return name_;
}

std::size_t FileMetadata::size_bytes() const {
    return size_bytes_;
}

const std::string& FileMetadata::notes() const
{
    return notes_;
}

const std::string& FileMetadata::patient_id() const
{
    return patient_id_;
}

const std::string& FileMetadata::type() const
{
    return type_;
}

const std::string& FileMetadata::content_id() const
{
    return content_id_;
}

int64_t FileMetadata::last_updated() const
{
    return last_updated_;
}

bool FileMetadata::deleted() const
{
    return deleted_;
}

template<>
FileMetadata::parse_result_t FileMetadata::parse<rapidjson::Document::ValueType>(const rapidjson::Document::ValueType& doc)
{

/*
 `json:"id"`
 `json:"name"`
 `json:"notes"`
 `json:"patientId"`
 `json:"type"`
 `json:"contentId"`
 `json:"lastUpdated"`
 `json:"deleted"`
*/

    if (!doc.HasMember("id") || !doc["id"].IsString()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("name") || !doc["name"].IsString()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("type") || !doc["type"].IsString()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("contentId") || !doc["contentId"].IsString()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("lastUpdated") || !doc["lastUpdated"].IsInt64()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("deleted") || !doc["deleted"].IsBool()) {
        return util::Unexpected<int>{1};
    }

    return FileMetadata{
        doc["id"].GetString(),
        doc["name"].GetString(),
        11, // TODO(mredolatti): get size from server
        doc.HasMember("notes") ? doc["notes"].GetString() : "",
        doc.HasMember("patientId") ? doc["patientId"].GetString() : "",
        doc["type"].GetString(),
        doc["contentId"].GetString(),
        doc["lastUpdated"].GetInt64(),
        doc["deleted"].GetBool()
    };
}

std::ostream& operator<<(std::ostream& sink, const FileMetadata& fm)
{
    sink << fmt::format("[id={}, name={}, size={}, type={}]", fm.id(), fm.size_bytes(), fm.name(), fm.type());
    return sink;
}


} // namespace mifs::models


