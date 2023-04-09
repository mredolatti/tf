#include "filemeta.hpp"
#include <fmt/format.h>
#include <rapidjson/document.h>

namespace mifs::models
{

FileMetadata::FileMetadata(std::string id, std::string name, std::size_t size_bytes, std::string notes,
                           std::string patient_id, std::string type, std::string content_id,
                           int64_t last_updated, bool deleted)
    : id_{std::move(id)},
      name_{std::move(name)},
      size_bytes_{size_bytes},
      notes_{std::move(notes)},
      patient_id_{std::move(patient_id)},
      type_{std::move(type)},
      content_id_{std::move(content_id)},
      last_updated_{last_updated},
      deleted_{deleted}
{
}

const std::string& FileMetadata::id() const { return id_; }

const std::string& FileMetadata::name() const { return name_; }

std::size_t FileMetadata::size_bytes() const { return size_bytes_; }

const std::string& FileMetadata::notes() const { return notes_; }

const std::string& FileMetadata::patient_id() const { return patient_id_; }

const std::string& FileMetadata::type() const { return type_; }

const std::string& FileMetadata::content_id() const { return content_id_; }

int64_t FileMetadata::last_updated() const { return last_updated_; }

bool FileMetadata::deleted() const { return deleted_; }

template <>
FileMetadata::parse_result_t
FileMetadata::parse<rapidjson::Document::ValueType>(const rapidjson::Document::ValueType& doc)
{
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

    if (!doc.HasMember("sizeBytes") || !doc["sizeBytes"].IsInt64()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("lastUpdated") || !doc["lastUpdated"].IsInt64()) {
        return util::Unexpected<int>{1};
    }

    if (!doc.HasMember("deleted") || !doc["deleted"].IsBool()) {
        return util::Unexpected<int>{1};
    }

    return FileMetadata{doc["id"].GetString(),
                        doc["name"].GetString(),
                        static_cast<size_t>(doc["sizeBytes"].GetInt()),
                        doc.HasMember("notes") ? doc["notes"].GetString() : "",
                        doc.HasMember("patientId") ? doc["patientId"].GetString() : "",
                        doc["type"].GetString(),
                        doc["contentId"].GetString(),
                        doc["lastUpdated"].GetInt64(),
                        doc["deleted"].GetBool()};
}

template <> int FileMetadata::serialize<rapidjson::Document>(rapidjson::Document& doc, bool ignore_empty) const
{
    doc.SetObject();
    auto& alloc{doc.GetAllocator()};

    if (!id_.empty() || !ignore_empty) {
        doc.AddMember("id", rapidjson::Value{id_.c_str(), alloc}.Move(), doc.GetAllocator());
    }

    if (!name_.empty() || !ignore_empty) {
        doc.AddMember("name", rapidjson::Value{name_.c_str(), alloc}.Move(), doc.GetAllocator());
    }

    if (!notes_.empty() || !ignore_empty) {
        doc.AddMember("notes", rapidjson::Value{notes_.c_str(), alloc}.Move(), doc.GetAllocator());
    }

    if (!patient_id_.empty() || !ignore_empty) {
        doc.AddMember("patientId", rapidjson::Value{patient_id_.c_str(), alloc}.Move(), doc.GetAllocator());
    }

    if (!type_.empty() || !ignore_empty) {
        doc.AddMember("type", rapidjson::Value{type_.c_str(), alloc}.Move(), doc.GetAllocator());
    }

    if (last_updated_ != -1 || !ignore_empty) {
        doc.AddMember("lastUpdated", last_updated_, doc.GetAllocator());
    }

    if (size_bytes_ != -1 || !ignore_empty) {
        doc.AddMember("sizeBytes", size_bytes_, doc.GetAllocator());
    }

    if (deleted_ || !ignore_empty) {
        doc.AddMember("deleted", deleted_, doc.GetAllocator());
    }

    return 0;
}


std::ostream& operator<<(std::ostream& sink, const FileMetadata& fm)
{
    sink << fmt::format("[id={}, name={}, size={}, type={}]", fm.id(), fm.size_bytes(), fm.name(), fm.type());
    return sink;
}

} // namespace mifs::models
