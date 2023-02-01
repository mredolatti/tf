#include "fstree.hpp"
#include "fselems.hpp"
#include <algorithm>
#include <iostream>

namespace mifs::filesystem {

namespace helpers {
std::pair<std::string_view, std::string_view> remove1st(std::string_view path);
} // namespace helpers


Node::Node(std::string name) :
    name_{std::move(name)}
{}

const std::string& Node::name() const
{
    return name_;
}

// ----------------------------------

InnerNode::InnerNode(std::string path) :
    Node{std::move(path)}
{}

bool InnerNode::insert(std::string_view path, std::unique_ptr<Node> node)
{
    using vt = map_t::value_type;
    if (path.empty()) {
        return false;
    }

    auto [head, tail]{helpers::remove1st(path)};
    if (tail.empty()) {
        return children_.insert(vt{std::string{head}, std::move(node)}).second;
    }
    auto it{children_.find(std::string{head})}; 
    if (it == children_.end()) {
        it = children_.insert(vt{std::string{head}, std::make_unique<InnerNode>(std::string{head})}).first;
    }
    return it->second->insert(tail, std::move(node));
}

std::unique_ptr<types::FSElem> InnerNode::get() const
{
    return std::make_unique<types::FSEFolder>(name_);
}

std::vector<std::unique_ptr<types::FSElem>> InnerNode::children() const
{
    using vt = map_t::value_type;
    std::vector<std::unique_ptr<types::FSElem>> to_ret;
    to_ret.reserve(children_.size());
    std::transform(children_.cbegin(), children_.cend(), std::back_inserter(to_ret),
        [](const vt& item) { return item.second->get(); });
    return to_ret;
}

const Node* InnerNode::follow_path(std::string_view path) const
{
    if (path.empty()) {
        return this;
    }

    auto [head, tail]{helpers::remove1st(path)};
    if (auto it{children_.find(std::string{head})}; it != children_.end()) {
        return it->second->follow_path(tail);
    }

    return nullptr;
}

void InnerNode::print(std::size_t depth) const
{
    std::cout << std::string(depth, '-') << name_ << '\n';
    for (const auto& [_, child]: children_) {
        child->print(depth + 1);
    }
}

// ----------------------------------

LeafNode::LeafNode(std::string name, std::size_t size_bytes, std::string server_id, std::string ref, bool link) :
    Node{std::move(name)},
    size_bytes_{size_bytes},
    server_id_{std::move(server_id)},
    ref_{std::move(ref)},
    link_{link}
{}

std::unique_ptr<LeafNode> LeafNode::link(std::string name, std::string server_id, std::string ref)
{
    return std::make_unique<LeafNode>(std::move(name), 0, std::move(server_id), std::move(ref), true); 
}

std::unique_ptr<LeafNode> LeafNode::file(std::string name, std::string server_id, std::string ref, std::size_t size_bytes)
{
    return std::make_unique<LeafNode>(std::move(name), size_bytes, std::move(server_id), std::move(ref), false);
}

bool LeafNode::insert(std::string_view path, std::unique_ptr<Node> node)
{
    return false;
}

std::unique_ptr<types::FSElem> LeafNode::get() const
{
    using rt = std::unique_ptr<types::FSElem>;
    return (link_)
        ? rt{std::make_unique<types::FSELink>(name_, server_id_, ref_)}
        : rt{std::make_unique<types::FSEFile>(name_, server_id_, ref_, size_bytes_)};
}

std::vector<std::unique_ptr<types::FSElem>> LeafNode::children() const
{
    return {};
}

const Node* LeafNode::follow_path(std::string_view path) const
{
    return (path.empty()) ? this : nullptr;
}

void LeafNode::print(std::size_t depth) const
{
    std::cout << std::string(depth, '-') << name_ << '\n';
}

// ----------------------------------
// ----------------------------------

namespace helpers {
std::pair<std::string_view, std::string_view> remove1st(std::string_view path)
{
    if (auto idx{path.find('/')}; idx != std::string_view::npos) {
        return {path.substr(0, idx), path.substr(idx+1)};
    }
    return {path, std::string_view{}};
}
} // namespace helpers

} // namespace mifs::filesystem
