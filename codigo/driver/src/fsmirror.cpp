#include "fsmirror.hpp"
#include "expected.hpp"
#include "filemeta.hpp"
#include "fstree.hpp"
#include "mappings.hpp"

#include <algorithm>
#include <cassert>
#include <fmt/format.h>
#include <iostream>
#include <memory>
#include <variant>
#include <set>


namespace mifs::util {

namespace helpers {
std::pair<std::string_view, std::string_view> pop_fname(std::string_view path);
}

FSMirror::FSMirror() :
    root_{""}
{}

FSMirror::Error FSMirror::mkdir(std::string_view path)
{
    // TODO(mredolatti): check path doesn't belong to a server
    auto [folder, fname]{helpers::pop_fname(path)};
    return (root_.insert(path, std::make_unique<filesystem::InnerNode>(std::string{fname})))
        ? Error::Ok
        : Error::AlreadyExists;
}

FSMirror::Error FSMirror::add_file(std::string_view name, std::string_view server, std::string_view ref, std::size_t size_bytes)
{
    return root_.insert(
            fmt::format("servers/{}/{}", server, name),
            filesystem::LeafNode::file(std::string{name}, std::string{server}, std::string{ref}, size_bytes))
        ? Error::Ok
        : Error::AlreadyExists;
}

FSMirror::Error FSMirror::link_file(std::string_view server, std::string_view ref, std::string path)
{
    // TODO(mredolatti): check path doesn't belong to a server
    auto [folder, fname]{helpers::pop_fname(path)};
    return root_.insert(
            path,
            filesystem::LeafNode::link(std::string{fname}, std::string{server}, std::string{ref}))
        ? Error::Ok
        : Error::AlreadyExists;
}

FSMirror::Error FSMirror::reset_mappings(const std::vector<models::Mapping>& mappings)
{
    root_ = filesystem::InnerNode{""};

    bool all_ok{true};

    for (const auto& mapping : mappings) {
//        all_ok &= (add_file(mapping.server(), mapping.ref(), mapping.size_bytes()) == Error::Ok);
        all_ok &= (link_file(mapping.server(), mapping.ref(), mapping.name()) == Error::Ok);
    }

    return all_ok ? Error::Ok : Error::NotFound;
}

FSMirror::Error FSMirror::reset_server(const std::string& server_id, const std::vector<models::FileMetadata>& files)
{
    bool all_ok{true};

    for (const auto& file : files) {
        all_ok &= (add_file(file.name(), server_id, file.id(), file.size_bytes()) == Error::Ok);
    }

    return all_ok ? Error::Ok : Error::NotFound;
}


FSMirror::Error FSMirror::reset_all(
        std::vector<models::Mapping>&& mappings,
        std::unordered_map<std::string, std::vector<models::FileMetadata>>&& files_by_server)
{
    root_ = filesystem::InnerNode{""};

    bool all_ok{true};

    for (const auto& mapping : mappings) {
        all_ok &= (link_file(mapping.server(), mapping.ref(), mapping.name()) == Error::Ok);
    }

    for (auto&& [server_id, files] : files_by_server) {
        for (const auto& file : files) {
            all_ok &= (add_file(file.name(), server_id, file.id(), file.size_bytes()) == Error::Ok);
        }
    }

    return all_ok ? Error::Ok : Error::NotFound;
}



FSMirror::list_result_t FSMirror::ls(std::string_view path)
{
    const auto* node{root_.follow_path(path)};
    if (node == nullptr) {
        return Unexpected<Error>{Error::NotFound};
    }

    return FSMirror::list_result_t{node->children()};
}

FSMirror::info_result_t FSMirror::info(std::string_view path)
{
    const auto* node{root_.follow_path(path)};
    if (node == nullptr) {
        return Unexpected<Error>{Error::NotFound};
    }
    return FSMirror::info_result_t{node->get()};
}

namespace helpers {
std::pair<std::string_view, std::string_view> pop_fname(std::string_view path)
{
    auto slash{path.find_last_of('/')};
    if (slash == std::string_view::npos) {
        return {{}, path};
    }
    return {path.substr(0, slash), path.substr(slash + 1)};
}

}

} // namespace mifs::util
