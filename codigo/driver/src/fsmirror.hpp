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
        ErrorAddingFiles = 6,
    };

    using path_t = std::filesystem::path;
    using info_result_t = util::Expected<std::unique_ptr<types::FSElem>, Error>;
    using list_result_t = util::Expected<std::vector<std::unique_ptr<types::FSElem>>, Error>;

    FSMirror();
    FSMirror(const FSMirror&) = delete;
    FSMirror(FSMirror&&) noexcept = default;
    FSMirror& operator=(const FSMirror&) = delete;
    FSMirror& operator=(FSMirror&&) noexcept = default;
    ~FSMirror() = default;

    Error mkdir(std::filesystem::path path);
    Error rmdir(std::filesystem::path path);
    Error add_file(std::string_view org, std::string_view server, std::string_view ref, std::size_t size_bytes, int64_t last_updated);
    Error link_file(std::string_view id, std::string_view org, std::string_view server, std::string_view ref, std::filesystem::path);
    list_result_t ls(std::filesystem::path path);
    info_result_t info(std::filesystem::path path);
    Error remove(std::filesystem::path path);

    Error reset_all(const std::vector<models::Mapping>& mappings);

    private:
    fstree::InnerNode root_;
};

}
#endif
