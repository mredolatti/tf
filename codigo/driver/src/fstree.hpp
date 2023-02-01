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

    Node(std::string name);
    const std::string& name() const;
    virtual bool insert(std::string_view path, std::unique_ptr<Node> node) = 0;
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
    bool insert(std::string_view path, std::unique_ptr<Node> node) override;
    std::unique_ptr<types::FSElem> get() const override;
    std::vector<std::unique_ptr<types::FSElem>> children() const override;
    const Node* follow_path(std::string_view path) const override;
    void print(std::size_t depth = 1) const override;

    private:
    using map_t = std::unordered_map<std::string, std::unique_ptr<Node>>;
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

    LeafNode(std::string name, std::size_t size_bytes, std::string server_id, std::string ref, bool link);
    static std::unique_ptr<LeafNode> link(std::string name, std::string server_id, std::string ref);
    static std::unique_ptr<LeafNode> file(std::string name, std::string server_id, std::string ref, std::size_t bytes);

    bool insert(std::string_view path, std::unique_ptr<Node> node) override;
    std::unique_ptr<types::FSElem> get() const override;
    std::vector<std::unique_ptr<types::FSElem>> children() const override;
    const Node* follow_path(std::string_view path) const override;
    void print(std::size_t depth = 1) const override;

    private:
    std::size_t size_bytes_;
    std::string server_id_;
    std::string ref_;
    bool link_;
};

} // namespace mifs::filesystem
#endif // MIFS_FILESYSTEM_TREE_HPP
