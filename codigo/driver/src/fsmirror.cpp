#include "fsmirror.hpp"
#include "expected.hpp"
#include "mappings.hpp"

#include <algorithm>
#include <cassert>
#include <fmt/format.h>
#include <iostream>
#include <memory>
#include <variant>


namespace mifs::util {
namespace detail {
namespace helpers {

std::pair<std::string_view, std::string_view> remove1st(std::string_view path);
std::unique_ptr<FileTreeNode> make_node(std::string_view name, const models::Mapping* mapping);

} // namespace helperes

// FSElem

FSElem::FSElem(std::string name, std::size_t bytes, bool is_folder) :
    name_{std::move(name)},
    size_bytes_{bytes},
    is_folder_{is_folder}
{}

const std::string& FSElem::name()   const { return name_; }
std::size_t FSElem::size_bytes()    const { return size_bytes_; }
bool FSElem::is_folder()            const { return is_folder_; }

// FileTreeNode

FileTreeNode::FileTreeNode(std::string_view name, data_t data) :
    name_{name},
    data_{std::move(data)}
{}

FileTreeNode* FileTreeNode::get(std::string_view path)
{
    if (path.empty()) {
        return this;
    }

    if (is_folder()) {
        auto& as_dir{as_cdirectory()};
        auto [head, tail]{helpers::remove1st(path)};
        if (auto it{as_dir.find(std::string{head})}; it != as_dir.end()) {
            assert(it->second); // shold NOT be null
            return it->second->get(tail);
        }
    }
    return nullptr;
}

FileTreeNode* FileTreeNode::add(std::string_view path, const models::Mapping* mapping)
{
    using vt = directory_t::value_type;

    if (!is_folder()) { // we reached a leaf prematurely
        return nullptr;
    }

    auto& as_dir{as_directory()};
    auto [head, tail]{helpers::remove1st(path)};
    auto it{as_dir.find(std::string{head})};
    if (!tail.empty()) { // need to keep traversing the tree
        if (it == as_dir.end()) {
            it = as_dir.insert(vt{std::string{head}, helpers::make_node(head, nullptr)}).first;
        }
        return it->second->add(tail, mapping);
    }

    if (it == as_dir.end()) {
        it = as_dir.insert(vt{std::string{head}, helpers::make_node(head, mapping)}).first;
        return it->second.get();
    }

    // if we're trying to create a new folder (null mapping), and it already exists,
    // just return a pointer to it. If it exists and it's a file, return nullptr.
    if (mapping == nullptr) {
        return it->second->is_folder()
            ? it->second.get()
            : nullptr;
    }

    if (it->second->is_folder()) {
        return nullptr;
    }

    it->second->data_ = *mapping;
    return it->second.get();
}

bool FileTreeNode::is_folder() const
{
    return std::holds_alternative<FileTreeNode::directory_t>(data_);
}

const std::string& FileTreeNode::name() const
{
    return name_;
}

FileTreeNode::directory_t& FileTreeNode::as_directory()
{
    return std::get<directory_t>(data_);
}

FileTreeNode::file_meta_t& FileTreeNode::as_filemeta()
{
    return std::get<file_meta_t>(data_);
}

const FileTreeNode::directory_t& FileTreeNode::as_cdirectory() const
{
    return std::get<directory_t>(data_);
}

const FileTreeNode::file_meta_t& FileTreeNode::as_cfilemeta() const
{
    return std::get<file_meta_t>(data_);
}

void FileTreeNode::print() const
{
    print_rec(this, 1);
}

const FileTreeNode::data_t& FileTreeNode::elem() const
{
    return data_;
}

void FileTreeNode::print_rec(const FileTreeNode* start, int level) const
{
    if (!start->is_folder()) { // this shold only happen at the root
        std::cout << std::string(level, '*') << " <F> " << start->as_cfilemeta().name() << '\n';
        return;
    }

    for (const auto& [key, value] : start->as_cdirectory()) {
        std::cout << std::string(level, '*') << (value->is_folder() ? " <D> " : " <F> ") << key << '\n';
        if (value->is_folder()) {
            print_rec(value.get(), level + 1);
        }
    }
}

// Helpers
namespace helpers {

std::pair<std::string_view, std::string_view> remove1st(std::string_view path)
{
    if (auto idx{path.find('/')}; idx != std::string_view::npos) {
        return {path.substr(0, idx), path.substr(idx+1)};
    }
    return {path, std::string_view{}};
}

std::unique_ptr<FileTreeNode> make_node(std::string_view name, const models::Mapping* mapping)
{
    return (mapping == nullptr)
        ? std::make_unique<FileTreeNode>(name, FileTreeNode::directory_t{})
        : std::make_unique<FileTreeNode>(name, *mapping);
}

} // namespace helpers
} // namespace detail

FSMirror::FSMirror() :
    root_{detail::FileTreeNode{"", detail::FileTreeNode::directory_t{}}}
{}

FSMirror::Error FSMirror::mkdir(std::string_view path)
{
    // TODO(mredolatti): check path doesn't belong to a server
    return (root_.add(path, nullptr) != nullptr)
        ? Error::Ok
        : Error::AlreadyExists;
}

FSMirror::Error FSMirror::add_file(std::string_view server, std::string_view ref, std::size_t size_bytes)
{
    // TODO(mredolatti): Add file to server
    return Error::Ok;
}

FSMirror::Error FSMirror::link_file(std::string_view server, std::string_view ref, std::string path)
{
    // TODO(mredolatti): link file to server ref
    return Error::Ok;
}

FSMirror::Error FSMirror::reset(const std::vector<models::Mapping>& mappings)
{
    root_ = detail::FileTreeNode{"", detail::FileTreeNode::directory_t{}};
    for (auto&& mapping : mappings) {
        //root_.add(fmt::format("servers/{}/{}", mapping.server(), mapping.ref()));
        root_.add(mapping.name(), &mapping);
    }
    return Error::Ok;
}

FSMirror::list_result_t FSMirror::ls(std::string_view path)
{
    const auto* node{root_.get(path)};
    if (node == nullptr) {
        return Unexpected<Error>{Error::NotFound};
    }

    if (!node->is_folder()) {
        const auto& as_file{node->as_cfilemeta()};
        return list_result_t{std::vector<detail::FSElem>{
            detail::FSElem{as_file.name(), as_file.size_bytes(), false}
        }};
    }

    const auto& as_dir{node->as_cdirectory()};
    if (as_dir.empty()) {
        return list_result_t{std::vector<detail::FSElem>{}};
    }

    std::vector<detail::FSElem> result;
    //(as_dir.size());
    std::transform(as_dir.cbegin(), as_dir.cend(), std::back_inserter(result), [](const auto& kv) {
        const auto& [k, v]{kv};
        return detail::FSElem{
            k, 
            (v->is_folder()) ? 0 : v->as_cfilemeta().size_bytes(),
            v->is_folder()
        };
    });
    return list_result_t{std::move(result)};
}


FSMirror::info_result_t FSMirror::info(std::string_view path)
{
    const auto* node{root_.get(path)};
    if (node == nullptr) {
        return util::Unexpected<Error>{Error::NotFound};
    }

    return info_result_t{detail::FSElem{node->name(), 123, node->is_folder()}};
}


} // namespace mifs::util
