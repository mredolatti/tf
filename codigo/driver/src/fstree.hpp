#ifndef MIFS_FILESYSTEM_TREE_HPP
#define MIFS_FILESYSTEM_TREE_HPP

#include "fselems.hpp"

#include <memory>
#include <string>
#include <unordered_map>
#include <vector>


namespace mifs::filesystem {

class Node
{
    public:

    Node() = delete;
    Node(const Node&) = default;
    Node(Node&&) = default;
    Node& operator=(const Node&) = default;
    Node& operator=(Node&&) = default;
    ~Node() = default;

    using node_ptr_t = std::unique_ptr<Node>;

    Node(std::string name);
    const std::string& name() const;
    virtual bool insert(std::string_view path, node_ptr_t node) = 0;
    virtual std::unique_ptr<types::FSElem> get() const = 0;
    virtual std::vector<std::unique_ptr<types::FSElem>> children() const = 0;
    virtual const Node* follow_path(std::string_view path) const = 0;
    virtual void print(std::size_t depth = 1) const = 0;

    protected:
    std::string name_;
};

class InnerNode : public Node
{
    public:
    InnerNode() = delete;
    InnerNode(const InnerNode&) = default;
    InnerNode(InnerNode&&) = default;
    InnerNode& operator=(const InnerNode&) = default;
    InnerNode& operator=(InnerNode&&) = default;
    ~InnerNode() = default;

    InnerNode(std::string path);
    bool insert(std::string_view path, node_ptr_t node) override;
    std::unique_ptr<types::FSElem> get() const override;
    std::vector<std::unique_ptr<types::FSElem>> children() const override;
    const Node* follow_path(std::string_view path) const override;
    void print(std::size_t depth = 1) const override;

    private:
    using map_t = std::unordered_map<std::string, node_ptr_t>;
    map_t children_;
};

class LeafNode : public Node
{
    public:
    LeafNode() = delete;
    LeafNode(const LeafNode&) = default;
    LeafNode(LeafNode&&) = default;
    LeafNode& operator=(const LeafNode&) = default;
    LeafNode& operator=(LeafNode&&) = default;
    ~LeafNode() = default;

    using leaf_ptr_t = std::unique_ptr<LeafNode>;

    LeafNode(std::string_view name, std::size_t size_bytes, std::string_view org, std::string_view server, std::string ref, int64_t last_updated, bool link);
    static leaf_ptr_t link(std::string_view name, std::string_view org, std::string_view server, std::string_view ref);
    static leaf_ptr_t file(std::string_view name, std::string_view org, std::string_view server, std::string ref, std::size_t size_bytes, int64_t last_updated);

    bool insert(std::string_view path, std::unique_ptr<Node> node) override;
    std::unique_ptr<types::FSElem> get() const override;
    std::vector<std::unique_ptr<types::FSElem>> children() const override;
    const Node* follow_path(std::string_view path) const override;
    void print(std::size_t depth = 1) const override;

    private:
    std::size_t size_bytes_;
    std::string org_name_;
    std::string server_name_;
    std::string ref_;
    int64_t last_updated_;
    bool link_;
};

} // namespace mifs::filesystem
#endif // MIFS_FILESYSTEM_TREE_HPP
