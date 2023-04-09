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

// class DirentryStub;

class FileManager
{
  public:
    using list_result_t = util::Expected<std::vector<fstree::views::Wrapper>, int>;
    using stat_result_t = util::Expected<fstree::views::Wrapper, int>;
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

    int touch(std::string_view path);
    int open(std::string_view path, int mode);
    int read(std::string_view path, char *buffer, std::size_t offset, std::size_t count);
    int read(int fd, char *buffer, off_t offset, std::size_t count);
    int write(std::string_view path, const char *buf, size_t size, off_t offset);
    bool flush(std::string_view path);

    bool mkdir(std::string_view path);
    bool rmdir(std::string_view path);

    bool remove(std::string_view path);
    bool rename(std::string_view from, std::string_view to);
    bool link(std::string_view from, std::string_view to);

    void sync();

  private:
    util::FSMirror fs_mirror_;
    util::FileCache file_cache_;
    util::OpenFileTracker open_files_;
    util::FileServerCatalog::ptr_t fs_catalog_;
    apiclients::IndexServerClient is_client_;
    apiclients::FileServerClient fs_client_;
    log::logger_t logger_;

    bool ensure_cached(const std::string& org, const std::string& server, const std::string& ref);
};

} // namespace mifs

#endif
