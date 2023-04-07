#include "fstree.hpp"
#include "fselems.hpp"
#include <algorithm>
#include <iostream>

namespace mifs::fstree
{

namespace helpers
{
std::pair<std::string, std::filesystem::path> strip_first(const std::filesystem::path& p);

} // namespace helpers

Node::Node(std::string name)
    : name_{std::move(name)}
{
}

const std::string& Node::name() const { return name_; }

// ----------------------------------

InnerNode::InnerNode(std::string path)
    : Node{std::move(path)}
{
}

bool InnerNode::insert(path_t path, node_ptr_t node)
{
    using vt = map_t::value_type;
    if (path.empty() || path == ".") {
        return false;
    }

    auto [head, tail]{helpers::strip_first(path)};
    if (tail.empty() || tail == ".") {
        return children_.insert(vt{head, std::move(node)}).second;
    }
    auto it{children_.find(head)};
    if (it == children_.end()) {
        it = children_.insert(vt{head, std::make_unique<InnerNode>(head)}).first;
    }
    return it->second->insert(tail, std::move(node));
}

bool InnerNode::drop(path_t path, int flags)
{
    using vt = map_t::value_type;
    if (path.empty() || path == ".") {
        return (flags & static_cast<int>(DropFlags::IF_DIR)) != 0;
    }

    auto [head, tail]{helpers::strip_first(path)};
    if (tail.empty() || tail == ".") { // we're staing at the folder that has the item to be deleted,
                                       // ask the item if it's ok to delete it, and if so do it
        auto it{children_.find(head)};
        if (it != children_.end() && it->second->drop(tail, flags)) {
            this->children_.erase(it);
            return true;
        }
        return false;
    }

    auto it{children_.find(head)};
    if (it == children_.end()) {
        it = children_.insert(vt{head, std::make_unique<InnerNode>(head)}).first;
    }
    return it->second->drop(tail, flags);
}

std::unique_ptr<types::FSElem> InnerNode::get() const { return std::make_unique<types::FSEFolder>(name_); }

std::vector<std::unique_ptr<types::FSElem>> InnerNode::children() const
{
    using vt = map_t::value_type;
    std::vector<std::unique_ptr<types::FSElem>> to_ret;
    to_ret.reserve(children_.size());
    std::transform(children_.cbegin(), children_.cend(), std::back_inserter(to_ret),
                   [](const vt& item) { return item.second->get(); });
    return to_ret;
}

const Node *InnerNode::follow_path(path_t path) const
{
    if (path.empty() || path == ".") {
        return this;
    }

    auto [head, tail]{helpers::strip_first(path)};
    if (auto it{children_.find(std::string{head})}; it != children_.end()) {
        return it->second->follow_path(tail);
    }

    return nullptr;
}

void InnerNode::print(std::size_t depth) const
{
    std::cout << std::string(depth, '-') << name_ << '\n';
    for (const auto& [_, child] : children_) {
        child->print(depth + 1);
    }
}

// ----------------------------------

LeafNode::LeafNode(std::string_view id, std::string_view name, std::size_t size_bytes, std::string_view org,
                   std::string_view server, std::string_view ref, int64_t last_updated, bool link)
    : Node{std::string{name}},
      mapping_id_{id},
      org_name_{org},
      server_name_{server},
      ref_{ref},
      size_bytes_{size_bytes},
      last_updated_{last_updated},
      link_{link}
{
}

std::unique_ptr<LeafNode> LeafNode::link(std::string_view id, std::string_view name, std::string_view org,
                                         std::string_view server, std::string_view ref)
{
    return std::make_unique<LeafNode>(id, name, 0, org, server, ref, 0, true);
}

std::unique_ptr<LeafNode> LeafNode::file(std::string_view name, std::string_view org, std::string_view server,
                                         std::string ref, std::size_t size_bytes, int64_t last_updated)
{
    return std::make_unique<LeafNode>("", name, size_bytes, org, server, ref, last_updated, false);
}

bool LeafNode::insert(path_t path, std::unique_ptr<Node> node) { return false; }

bool LeafNode::drop(path_t path, int flags)
{

    if ((flags & static_cast<int>(DropFlags::RECURSIVE)) != 0) {
        return true;
    }

    if (link_ && (flags & static_cast<int>(DropFlags::IF_FILE))) {
        return true;
    }

    return false;
}

std::unique_ptr<types::FSElem> LeafNode::get() const
{
    using rt = std::unique_ptr<types::FSElem>;
    return (link_) ? rt{std::make_unique<types::FSELink>(mapping_id_, name_, org_name_, server_name_, ref_)}
                   : rt{std::make_unique<types::FSEFile>(name_, org_name_, server_name_, ref_, size_bytes_,
                                                         last_updated_)};
}

std::vector<std::unique_ptr<types::FSElem>> LeafNode::children() const { return {}; }

const Node *LeafNode::follow_path(path_t path) const
{
    return (path.empty() || path == ".") ? this : nullptr;
}

void LeafNode::print(std::size_t depth) const { std::cout << std::string(depth, '-') << name_ << '\n'; }

// ----------------------------------
// ----------------------------------

namespace helpers
{

std::pair<std::string, std::filesystem::path> strip_first(const std::filesystem::path& p)
{
    return p.relative_path().empty()
               ? std::make_pair(std::string{*p.begin()}, std::filesystem::path{})
               : std::make_pair(std::string{*p.begin()}, p.lexically_relative(*p.begin()));
}

} // namespace helpers

} // namespace mifs::fstree
