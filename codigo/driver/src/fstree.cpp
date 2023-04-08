#include "fstree.hpp"
#include <algorithm>
#include <iostream>

namespace mifs::fstree
{

namespace helpers
{

std::pair<std::string, std::filesystem::path> strip_first(const std::filesystem::path& p);

} // namespace helpers

namespace views
{

Wrapper::Wrapper(Folder f)
    : wrapped_{std::move(f)}
{
}

Wrapper::Wrapper(File f)
    : wrapped_{std::move(f)}
{
}

Wrapper::Wrapper(Link l)
    : wrapped_{std::move(l)}
{
}

Folder *Wrapper::folder() { return std::get_if<Folder>(&wrapped_); }
File *Wrapper::file() { return std::get_if<File>(&wrapped_); }
Link *Wrapper::link() { return std::get_if<Link>(&wrapped_); };
const Folder *Wrapper::folder() const { return std::get_if<Folder>(&wrapped_); }
const File *Wrapper::file() const { return std::get_if<File>(&wrapped_); }
const Link *Wrapper::link() const { return std::get_if<Link>(&wrapped_); };

const std::string& Link::get_name() const { return name; }
std::size_t Link::get_size_bytes() const { return size_bytes; }
int Link::get_last_updated_seconds() const { return last_updated; }

const std::string& File::get_name() const { return ref; }
std::size_t File::get_size_bytes() const { return size_bytes; }
int File::get_last_updated_seconds() const { return last_updated; }

const std::string& Folder::get_name() const { return name; }
std::size_t Folder::get_size_bytes() const { return 0; }
int Folder::get_last_updated_seconds() const { return 0; }

Type Wrapper::type() const
{
    return std::visit([](auto&& v) { return v.type; }, wrapped_);
}

const std::string& Wrapper::name() const
{
    return std::visit([](auto&& v) -> const std::string& { return v.get_name(); }, wrapped_);
}

std::size_t Wrapper::size_bytes() const
{
    return std::visit([](auto&& v) -> std::size_t { return v.get_size_bytes(); }, wrapped_);
}

int Wrapper::last_updated_seconds() const
{
    return std::visit([](auto&& v) -> int { return v.get_last_updated_seconds(); }, wrapped_);
}

} // namespace views

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

views::Wrapper InnerNode::get() const { return views::Wrapper{views::Folder{.name = name_}}; }

std::vector<views::Wrapper> InnerNode::children() const
{
    using vt = map_t::value_type;
    std::vector<views::Wrapper> to_ret;
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
                                         std::string_view ref, std::size_t size_bytes, int64_t last_updated)
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

views::Wrapper LeafNode::get() const
{
    views::File f{.organization_name = org_name_,
                  .server_name = server_name_,
                  .ref = ref_,
                  .size_bytes = size_bytes_,
                  .last_updated = last_updated_};

    return link_ ? views::Wrapper{views::Link{std::move(f), mapping_id_, name_}} : views::Wrapper{f};
}

std::vector<views::Wrapper> LeafNode::children() const { return {}; }

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
