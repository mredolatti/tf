#include "filemanager.hpp"
#include "expected.hpp"
#include "isclient.hpp"
#include "log.hpp"
#include "mappings.hpp"
#include <algorithm>

namespace mifs {

DirentryStub::DirentryStub(std::string name, size_t size, bool is_directory) :
    name_{std::move(name)},
    size_{size},
    is_directory_{is_directory}
{}

DirentryStub DirentryStub::from_fsmeta(const util::detail::FSElem& fselem)
{
    return DirentryStub{fselem.name(), fselem.size_bytes(), fselem.is_folder()};
}

const std::string& DirentryStub::name() const
{
    return name_;
}

size_t DirentryStub::size() const
{
    return size_;
}

bool DirentryStub::is_directory() const
{
    return is_directory_;
}


FileManager::FileManager(http_client_ptr_t http_client) :
    is_client_{apiclients::IndexServerClient{std::move(http_client)}},
    logger_{log::get()}
{}

FileManager::list_result_t FileManager::list(std::string_view path)
{
    auto result{fs_mirror_.ls((path.size() > 0 && path[0] == '/') ? path.substr(1) : path)};
    if (!result) {
        return util::Unexpected<int>{static_cast<int>(result.error())};
    }

    auto& fsmetas{*result};
    std::vector<DirentryStub> des;
    std::transform(fsmetas.begin(), fsmetas.end(), std::back_inserter(des), DirentryStub::from_fsmeta);
    return des;
}

FileManager::stat_result_t FileManager::stat(std::string_view path)
{
    auto result{fs_mirror_.info((path.size() > 0 && path[0] == '/') ? path.substr(1) : path)};
    if (!result) {
        return util::Unexpected<int>{static_cast<int>(result.error())};
    }

    auto fsmeta{*result};
    return DirentryStub::from_fsmeta(fsmeta);
}

void FileManager::sync()
{
    auto res{is_client_.get_all()};
    if (!res) {
        SPDLOG_LOGGER_ERROR(logger_, "failed to fetch mappings from index server");
        return;
    }
    fs_mirror_.reset((*res).data["mapping"]);
}

} // namespace mifs
