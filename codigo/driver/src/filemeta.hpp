#ifndef MIFS_FILE_META_HPP
#define MIFS_FILE_META_HPP

#include "expected.hpp"
#include <string>

namespace mifs::models {

class FileMetadata
{
    public:
    using parse_result_t = util::Expected<FileMetadata, int /* TODO */>;

    FileMetadata() = delete;
    FileMetadata(const FileMetadata&) = default;
    FileMetadata(FileMetadata&&) noexcept = default;
    FileMetadata& operator=(const FileMetadata&) = default;
    FileMetadata& operator=(FileMetadata&&) noexcept = default;
    ~FileMetadata() noexcept = default;

    template<typename Serialized>
    static parse_result_t parse(const Serialized& data);

    FileMetadata(std::string id, std::string name, std::size_t size_bytes, std::string notes, std::string patient_id,
            std::string type, std::string content_id, int64_t last_updated, bool deleted);

    const std::string& id() const;
    const std::string& name() const;
    std::size_t size_bytes() const;
    const std::string& notes() const;
    const std::string& patient_id() const;
    const std::string& type() const;
    const std::string& content_id() const;
    int64_t last_updated() const;
    bool deleted() const;

    private:
    std::string id_;
    std::string name_;
    std::size_t size_bytes_;
    std::string notes_;
    std::string patient_id_;
    std::string type_;
    std::string content_id_;
    int64_t last_updated_;
    bool deleted_;
};

std::ostream& operator<<(std::ostream& sink, const FileMetadata& fm);


} // namespace mifs::models
#endif // MIFS_FILE_META_HPP
