#ifndef MIFS_FILEMANAGER_HPP
#define MIFS_FILEMANAGER_HPP

#include <string_view>
#include <vector>

#include "expected.hpp"
#include "filecache.hpp"
#include "fsclient.hpp"
#include "fservers.hpp"
#include "fsmirror.hpp"
#include "isclient.hpp"

#include "log.hpp"
#include "openfiles.hpp"

namespace mifs
{

class FileManager
{
  public:
    enum class Error {
        Ok = 0,

        NotFound = 1,
        AlreadyExists = 2,
        NotAFile = 3,
        NotALink = 4,
        NotAForlder = 5,
        CannotWriteInNonServerPath = 6,
        InvalidLinkDestination = 7,
        InvalidLinkSource = 8,
        ServerTreeManipulation = 9,

        FailedToFetchMappings = 101,
        FiledToUpdateRemoteMapping = 102,
        FiledToReadFileFromServer = 103,
        FiledToWriteFileInServer = 104,
        FailedToFetchServerInfos = 110,

        InternalCacheError = 200,
        InternalRepresentationError = 201,

        Unknown = 999999,
    };

    using list_result_t = util::Expected<std::vector<fstree::views::Wrapper>, Error>;
    using stat_result_t = util::Expected<fstree::views::Wrapper, Error>;
    using http_client_ptr_t = std::shared_ptr<http::Client>;

    explicit FileManager(apiclients::IndexServerClient is_client, apiclients::FileServerClient fs_client,
                         util::FileServerCatalog::ptr_t fs_catalog);
    FileManager() = delete;
    FileManager(const FileManager&) = delete;
    FileManager(FileManager&&) = delete;
    FileManager& operator=(const FileManager&) = delete;
    FileManager& operator=(FileManager&&) = delete;
    ~FileManager() = default;

    list_result_t list(std::string_view path);
    stat_result_t stat(std::string_view path);

    Error touch(std::string_view path);
    int open(std::string_view path, int mode);
    std::pair<int, Error> read(std::string_view path, char *buffer, std::size_t offset, std::size_t count);
    std::pair<int, Error> read(int fd, char *buffer, off_t offset, std::size_t count);
    std::pair<int, Error> write(std::string_view path, const char *buf, size_t size, off_t offset);
    Error flush(std::string_view path);

    Error mkdir(std::string_view path);
    Error rmdir(std::string_view path);

    Error remove(std::string_view path);
    Error rename(std::string_view from, std::string_view to);
    Error link(std::string_view from, std::string_view to);

    Error sync();

  private:
    util::FSMirror fs_mirror_;
    util::FileCache file_cache_;
    util::OpenFileTracker open_files_;
    util::FileServerCatalog::ptr_t fs_catalog_;
    apiclients::IndexServerClient is_client_;
    apiclients::FileServerClient fs_client_;
    log::logger_t logger_;

    Error ensure_cached(const std::string& org, const std::string& server, const std::string& ref);
};

} // namespace mifs

#endif
