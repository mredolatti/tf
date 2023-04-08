#ifndef MIFS_FILESYSTEM_TREE_HPP
#define MIFS_FILESYSTEM_TREE_HPP

#include <filesystem>
#include <memory>
#include <string>
#include <unordered_map>
#include <variant>
#include <vector>

namespace mifs::fstree
{

namespace views
{

enum class Type {
    Folder = 0,
    File = 1,
    Link = 2,
};

struct Folder {
    std::string name;
    constexpr const static Type type = Type::Folder;

    const std::string& get_name() const;
    std::size_t get_size_bytes() const;
    int get_last_updated_seconds() const;
};

struct File {
    std::string organization_name;
    std::string server_name;
    std::string ref;
    std::size_t size_bytes;
    int64_t last_updated;
    constexpr const static Type type = Type::File;

    const std::string& get_name() const;
    std::size_t get_size_bytes() const;
    int get_last_updated_seconds() const;
};

struct Link : public File {
    std::string id;
    std::string name;
    constexpr const static Type type = Type::Link;

    const std::string& get_name() const;
    std::size_t get_size_bytes() const;
    int get_last_updated_seconds() const;
};

class Wrapper
{
  public:
    explicit Wrapper(Folder);
    explicit Wrapper(File);
    explicit Wrapper(Link);
    Wrapper() = delete;
    Wrapper(const Wrapper&) = default;
    Wrapper(Wrapper&&) noexcept = default;
    Wrapper& operator=(const Wrapper&) = default;
    Wrapper& operator=(Wrapper&&) = default;
    ~Wrapper() = default;

    Folder *folder();
    File *file();
    Link *link();
    const Folder *folder() const;
    const File *file() const;
    const Link *link() const;

    Type type() const;
    const std::string& name() const;
    std::size_t size_bytes() const;
    int last_updated_seconds() const;

  private:
    std::variant<Folder, File, Link> wrapped_;
};

} // namespace views

enum class DropFlags : int {
    IF_FILE = (1 << 0),
    IF_DIR = (1 << 1),
    RECURSIVE = (1 << 2),
};

class Node
{
  public:
    Node() = delete;
    Node(const Node&) = default;
    Node(Node&&) = default;
    Node& operator=(const Node&) = default;
    Node& operator=(Node&&) = default;
    ~Node() = default;

    using path_t = std::filesystem::path;
    using node_ptr_t = std::unique_ptr<Node>;

    Node(std::string name);
    const std::string& name() const;
    virtual bool insert(path_t path, node_ptr_t node) = 0;
    virtual bool drop(path_t path, int flags) = 0;
    virtual views::Wrapper get() const = 0;
    virtual std::vector<views::Wrapper> children() const = 0;
    virtual const Node *follow_path(path_t path) const = 0;
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
    bool insert(path_t path, node_ptr_t node) override;
    bool drop(path_t path, int flags) override;
    views::Wrapper get() const override;
    std::vector<views::Wrapper> children() const override;
    const Node *follow_path(path_t path) const override;
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

    LeafNode(std::string_view id, std::string_view name, std::size_t size_bytes, std::string_view org,
             std::string_view server, std::string_view ref, int64_t last_updated, bool link);

    static leaf_ptr_t link(std::string_view id, std::string_view name, std::string_view org,
                           std::string_view server, std::string_view ref);

    static leaf_ptr_t file(std::string_view name, std::string_view org, std::string_view server,
                           std::string_view ref, std::size_t size_bytes, int64_t last_updated);

    bool insert(path_t path, std::unique_ptr<Node> node) override;
    bool drop(path_t path, int flags) override;
    views::Wrapper get() const override;
    std::vector<views::Wrapper> children() const override;
    const Node *follow_path(path_t path) const override;
    void print(std::size_t depth = 1) const override;

  private:
    std::string mapping_id_;
    std::string org_name_;
    std::string server_name_;
    std::string ref_;
    std::size_t size_bytes_;
    int64_t last_updated_;
    bool link_;
};

} // namespace mifs::fstree
#endif // MIFS_FILESYSTEM_TREE_HPP
