#ifndef MIFS_UTIL_GTREE_HPP
#define MIFS_UTIL_GTREE_HPP

#include <memory>
#include <optional>
#include <string>
#include <unordered_map>
#include <variant>
#include <vector>

#include "expected.hpp"
#include "filemeta.hpp"
#include "fselems.hpp"
#include "fstree.hpp"
#include "mappings.hpp"

namespace mifs::util {

class FSMirror
{
    public:
    enum class Error
    {
        Ok = 0,
        AlreadyExists = 1,
        CannotLinkInServerFolder = 2,
        CannotAddInLinkedFolder = 3,
        NotFound = 4,
	ErrorAddingMappings = 5,
	ErrorAddingFiles = 6
    };

    using info_result_t = util::Expected<std::unique_ptr<types::FSElem>, Error>;
    using list_result_t = util::Expected<std::vector<std::unique_ptr<types::FSElem>>, Error>;

    FSMirror();
    FSMirror(const FSMirror&) = delete;
    FSMirror(FSMirror&&) noexcept = default;
    FSMirror& operator=(const FSMirror&) = delete;
    FSMirror& operator=(FSMirror&&) noexcept = default;
    ~FSMirror() = default;

    Error mkdir(std::string_view path);
    Error add_file(std::string_view name, std::string_view server, std::string_view ref, std::size_t size_bytes, int64_t last_updated);
    Error link_file(std::string_view org, std::string_view server, std::string_view ref, std::string path);
    list_result_t ls(std::string_view path);
    info_result_t info(std::string_view path);

    Error reset_all(
            std::vector<models::Mapping>&& mappings,
            std::unordered_map<std::string, std::vector<models::FileMetadata>>&& files_by_server);
    Error reset_all(const std::vector<models::Mapping>& mappings);
    Error reset_mappings(const std::vector<models::Mapping>& mappings);
    Error reset_server(const std::string& server_id, const std::vector<models::FileMetadata>& mappings);


    private:
    filesystem::InnerNode root_;
};

}
#endif
