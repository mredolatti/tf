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
#include "fstree.hpp"
#include "mappings.hpp"

namespace mifs::util
{

class FSMirror
{
  public:
    class Error
    {
        public:

        enum class Code {
            Ok = 0,
            AlreadyExists = 1,
            CannotLinkInServerFolder = 2,
            CannotAddInLinkedFolder = 3,
            NotFound = 4,
        };

        Error(Code c);
        Error(const Error&) = default;
        Error& operator=(const Error&) = default;
        Error(Error&&) noexcept = default;
        Error& operator=(Error&&) noexcept = default;
        ~Error() = default;

        Code code() const;
        const char* message() const;

        private:
        Code code_;
    };

    using path_t = std::filesystem::path;
    using info_result_t = util::Expected<fstree::views::Wrapper, Error>;
    using list_result_t = util::Expected<std::vector<fstree::views::Wrapper>, Error>;

    FSMirror();
    FSMirror(const FSMirror&) = delete;
    FSMirror(FSMirror&&) noexcept = default;
    FSMirror& operator=(const FSMirror&) = delete;
    FSMirror& operator=(FSMirror&&) noexcept = default;
    ~FSMirror() = default;

    Error mkdir(std::filesystem::path path);
    Error rmdir(std::filesystem::path path);
    Error add_file(std::string_view org, std::string_view server, std::string_view ref,
                   std::size_t size_bytes, int64_t last_updated);
    Error link_file(std::string_view id, std::string_view org, std::string_view server, std::string_view ref,
                    std::filesystem::path);
    list_result_t ls(std::filesystem::path path);
    info_result_t info(std::filesystem::path path);
    Error remove(std::filesystem::path path);

    std::vector<std::pair<std::size_t, Error>> reset_all(const std::vector<models::Mapping>& mappings);

  private:
    fstree::InnerNode root_;
};

} // namespace mifs::util
#endif
