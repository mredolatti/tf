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
#include <set>
#include <variant>

namespace mifs::util
{

FSMirror::FSMirror()
    : root_{""}
{
}

FSMirror::Error FSMirror::mkdir(path_t path)
{
    // TODO(mredolatti): check path doesn't belong to a server
    return (root_.insert(path, std::make_unique<fstree::InnerNode>(path.filename()))) ? Error::Code::Ok
                                                                                      : Error::Code::AlreadyExists;
}

FSMirror::Error FSMirror::rmdir(std::filesystem::path path)
{
    return root_.drop(path, static_cast<int>(fstree::DropFlags::IF_DIR))
             ? Error::Code::Ok
             : Error::Code::NotFound; // TODO: use a proper codehere
}

FSMirror::Error FSMirror::add_file(std::string_view org, std::string_view server, std::string_view ref,
                                   std::size_t size_bytes, int64_t last_updated)
{
    return root_.insert(fmt::format("servers/{}/{}/{}", org, server, ref),
                        fstree::LeafNode::file(ref, org, server, ref, size_bytes, last_updated))
             ? Error::Code::Ok
             : Error::Code::AlreadyExists;
}

FSMirror::Error FSMirror::link_file(std::string_view id, std::string_view org, std::string_view server,
                                    std::string_view ref, path_t path)
{
    // TODO(mredolatti): check path doesn't belong to a server
    return root_.insert(path, fstree::LeafNode::link(id, path.filename().c_str(), org, server, ref))
             ? Error::Code::Ok
             : Error::Code::AlreadyExists;
}

std::vector<std::pair<std::size_t, FSMirror::Error>>
FSMirror::reset_all(const std::vector<models::Mapping>& mappings)
{
    root_ = fstree::InnerNode{""};
    std::vector<std::pair<std::size_t, FSMirror::Error>> errors;
    for (std::size_t idx{0}; idx < mappings.size(); ++idx) {
        const auto& mapping{mappings[idx]};
        if (auto err{add_file(mapping.org(), mapping.server(), mapping.ref(), mapping.size_bytes(),
                              mapping.last_updated())};
            err.code() != Error::Code::Ok) {
            errors.emplace_back(idx, err);
        }
        if (!mapping.path().empty()) {
            if (auto err{
                    link_file(mapping.id(), mapping.org(), mapping.server(), mapping.ref(), mapping.path())};
                err.code() != Error::Code::Ok) {
                errors.emplace_back(idx, err);
            }
        }
    }

    return errors;
}

FSMirror::list_result_t FSMirror::ls(path_t path)
{
    const auto *node{root_.follow_path(path)};
    if (node == nullptr) {
        return Unexpected<Error>{Error::Code::NotFound};
    }

    return FSMirror::list_result_t{node->children()};
}

FSMirror::info_result_t FSMirror::info(path_t path)
{
    const auto *node{root_.follow_path(path)};
    if (node == nullptr) {
        return Unexpected<Error>{Error::Code::NotFound};
    }
    return FSMirror::info_result_t{node->get()};
}

FSMirror::Error FSMirror::remove(std::filesystem::path path)
{
    return root_.drop(path, static_cast<int>(fstree::DropFlags::IF_FILE))
             ? Error::Code::Ok
             : Error::Code::NotFound; // TODO: use a proper codehere
}

// Errors

const char* errOK = "Ok";
const char* errAlreadyExists = "item already exists";
const char* errCannotLinkInServerFolder = "cannot manipulate links inside the servers subtree";
const char* errCannotAddInLikedFolder = "files must be created in a server's path";
const char* errNotFound = "item not found";
const char* errUnknown = "unknown error";

FSMirror::Error::Error(Code c) : code_{c} {}

FSMirror::Error::Code FSMirror::Error::code() const { return code_; }
const char* FSMirror::Error::message() const
{
    switch (code_) {
    case Code::Ok: return errOK;
    case Code::AlreadyExists: return errAlreadyExists;
    case Code::CannotLinkInServerFolder: return errCannotAddInLikedFolder;
    case Code::CannotAddInLinkedFolder: return errCannotAddInLikedFolder;
    case Code::NotFound: return errNotFound;
    }

    return errUnknown;
}


} // namespace mifs::util
