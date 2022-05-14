#ifndef MIFS_FILEMANAGER_HPP
#define MIFS_FILEMANAGER_HPP

#include <vector>
#include <string_view>

#include "fsmirror.hpp"
#include "expected.hpp"
#include "isclient.hpp"
#include "log.hpp"

namespace mifs {

class DirentryStub;

class FileManager
{
    public:
    using list_result_t = util::Expected<std::vector<DirentryStub>, int>;
    using stat_result_t = util::Expected<DirentryStub, int>;
    using http_client_ptr_t = std::shared_ptr<http::Client>;

    explicit FileManager(http_client_ptr_t http_client);
    FileManager() = delete;
    FileManager(const FileManager&) = delete;
    FileManager(FileManager&&) = default;
    FileManager& operator=(const FileManager&) = delete;
    FileManager& operator=(FileManager&&) = delete;
    ~FileManager() = default;

    list_result_t list(std::string_view path);
    stat_result_t stat(std::string_view path);
    void sync();

    private:
    util::FSMirror fs_mirror_;
    apiclients::IndexServerClient is_client_;
    log::logger_t logger_;
};

class DirentryStub
{
    public:
    DirentryStub(std::string name, size_t size, bool is_directory);
    DirentryStub() = delete;
    DirentryStub(const DirentryStub&) = default;
    DirentryStub(DirentryStub&&) = default;
    DirentryStub& operator=(const DirentryStub&) = default;
    DirentryStub& operator=(DirentryStub&) = default;
    ~DirentryStub() = default;

    static DirentryStub from_fsmeta(const util::detail::FSElem& fselem);

    const std::string& name() const;
    size_t size() const;
    bool is_directory() const;

    private:
    std::string name_;
    size_t size_;
    bool is_directory_;
};

} // namespace mifs

#endif
