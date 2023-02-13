#ifndef MIFS_FILEMANAGER_HPP
#define MIFS_FILEMANAGER_HPP

#include <vector>
#include <string_view>

#include "filecache.hpp"
#include "fsclient.hpp"
#include "fselems.hpp"
#include "fsmirror.hpp"
#include "expected.hpp"
#include "isclient.hpp"
#include "fservers.hpp"

#include "log.hpp"
#include "openfiles.hpp"

namespace mifs {

//class DirentryStub;

class FileManager
{
    public:
    using list_result_t = util::Expected<std::vector<std::unique_ptr<types::FSElem>>, int>;
    using stat_result_t = util::Expected<std::unique_ptr<types::FSElem>, int>;
    using http_client_ptr_t = std::shared_ptr<http::Client>;

    explicit FileManager(apiclients::IndexServerClient is_client, apiclients::FileServerClient fs_client);
    FileManager() = delete;
    FileManager(const FileManager&) = delete;
    FileManager(FileManager&&) = delete;
    FileManager& operator=(const FileManager&) = delete;
    FileManager& operator=(FileManager&&) = delete;
    ~FileManager() = default;

    list_result_t list(std::string_view path);
    stat_result_t stat(std::string_view path);
    int open(std::string_view path, int mode);
    int read(std::string_view path, char *buffer, std::size_t offset, std::size_t count);
    int read(int fd, char *buffer, std::size_t offset, std::size_t count);

    void sync();

    private:
    util::FSMirror fs_mirror_;
    util::FileCache file_cache_;
    util::OpenFileTracker open_files_;
    apiclients::IndexServerClient is_client_;
    apiclients::FileServerClient fs_client_;
    log::logger_t logger_;

    bool ensure_cached(std::string server, std::string ref);
};

//class DirentryStub
//{
//    public:
//    DirentryStub(std::string name, size_t size, bool is_directory);
//    DirentryStub() = delete;
//    DirentryStub(const DirentryStub&) = default;
//    DirentryStub(DirentryStub&&) = default;
//    DirentryStub& operator=(const DirentryStub&) = default;
//    DirentryStub& operator=(DirentryStub&) = default;
//    ~DirentryStub() = default;
//
//    static DirentryStub from_fsmeta(const util::detail::FSElem& fselem);
//
//    const std::string& name() const;
//    size_t size() const;
//    bool is_directory() const;
//
//    private:
//    std::string name_;
//    size_t size_;
//    bool is_directory_;
//};

} // namespace mifs

#endif
