#include "fservers.hpp"
#include "mappings.hpp"

#include <unordered_set>
#include <algorithm>

namespace mifs::util {

ServerFile::ServerFile(std::string name, std::size_t size_bytes, bool is_folder) :
    name_{std::move(name)},
    size_bytes_{size_bytes},
    is_folder_{is_folder}
{}

const std::string& ServerFile::name() const
{
    return name_;
}

std::size_t ServerFile::size_bytes() const
{
    return size_bytes_;
}

void FileServersContents::reset(const std::vector<models::Mapping>& mappings)
{
    servers_t wip;
    for (const auto& mapping : mappings) {
        wip[mapping.server()].push_back(ServerFile{mapping.ref(), mapping.size_bytes(), false});
    }

    for (auto&& [server_id, refs] : wip) {
        std::sort(refs.begin(), refs.end(), [](auto&& x, auto&& y) { return x.name() < y.name(); });
        refs.erase(std::unique(refs.begin(), refs.end()));
    }

    std::unique_lock<std::mutex> lk(mutex_);
    servers_ = wip;
}

FileServersContents::list_result_t FileServersContents::ls(std::string_view path)
{
    if (path == "" || path == "/") {
        return std::vector<ServerFile>{ServerFile{"servers", 0, true}};
    } else if (path == "servers" || path == "/servers") {
        std::vector<ServerFile> folders(servers_.size());
        for (auto& [k, v] : servers_) {
            folders.push_back(ServerFile{k, 0, true});
        }
        return folders;
    } else {
        gT
    }
    
}



} // namespace mifs::util
