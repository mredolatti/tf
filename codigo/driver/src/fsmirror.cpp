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

FSMirror::FSMirror() :
    root_{""}
{}

FSMirror::Error FSMirror::mkdir(path_t path)
{
    // TODO(mredolatti): check path doesn't belong to a server
    return (root_.insert(path, std::make_unique<fstree::InnerNode>(path.filename())))
        ? Error::Ok
        : Error::AlreadyExists;
}

FSMirror::Error FSMirror::rmdir(std::filesystem::path path)
{
    return root_.drop(path, static_cast<int>(fstree::DropFlags::IF_DIR))
        ? Error::Ok
        : Error::NotFound; // TODO: use a proper codehere
}

FSMirror::Error FSMirror::add_file(std::string_view org, std::string_view server, std::string_view ref, std::size_t size_bytes, int64_t last_updated)
{
    return root_.insert(
            fmt::format("servers/{}/{}/{}", org, server, ref),
            fstree::LeafNode::file(ref, std::string{org}, std::string{server}, std::string{ref}, size_bytes, last_updated))
        ? Error::Ok
        : Error::AlreadyExists;
}

FSMirror::Error FSMirror::link_file(std::string_view id, std::string_view org, std::string_view server, std::string_view ref, path_t path)
{
    // TODO(mredolatti): check path doesn't belong to a server
    return root_.insert(path, fstree::LeafNode::link(id, path.filename().c_str(), org, server, ref))
        ? Error::Ok
        : Error::AlreadyExists;
}

FSMirror::Error FSMirror::reset_all(const std::vector<models::Mapping>& mappings) {
    root_ = fstree::InnerNode{""};

    bool mappings_ok{true};
    bool files_ok{true};
    for (const auto& mapping : mappings) {
        files_ok &= (add_file(mapping.org(), mapping.server(), mapping.ref(), mapping.size_bytes(), mapping.last_updated()) == Error::Ok);
        if (!mapping.path().empty()) {
            mappings_ok &= (link_file(mapping.id(), mapping.org(), mapping.server(), mapping.ref(), mapping.path()) == Error::Ok);
        }
    }

    if (!mappings_ok) return Error::ErrorAddingMappings;
    if (!files_ok) return Error::ErrorAddingFiles;
    return Error::Ok;
}


FSMirror::list_result_t FSMirror::ls(path_t path)
{
    const auto* node{root_.follow_path(path)};
    if (node == nullptr) {
        return Unexpected<Error>{Error::NotFound};
    }

    return FSMirror::list_result_t{node->children()};
}

FSMirror::info_result_t FSMirror::info(path_t path)
{
    const auto* node{root_.follow_path(path)};
    if (node == nullptr) {
        return Unexpected<Error>{Error::NotFound};
    }
    return FSMirror::info_result_t{node->get()};
}

FSMirror::Error FSMirror::remove(std::filesystem::path path)
{
    return root_.drop(path, static_cast<int>(fstree::DropFlags::IF_FILE))
        ? Error::Ok
        : Error::NotFound; // TODO: use a proper codehere
}


} // namespace mifs::util
