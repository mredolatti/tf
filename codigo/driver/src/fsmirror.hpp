#ifndef MIFS_UTIL_GTREE_HPP
#define MIFS_UTIL_GTREE_HPP

#include <memory>
#include <optional>
#include <string>
#include <unordered_map>
#include <variant>
#include <vector>

#include "expected.hpp"
#include "mappings.hpp"

namespace mifs::util {

namespace detail {

class FSElem
{
    public:
    FSElem() = delete;
    FSElem(const FSElem&) = default;
    FSElem(FSElem&&) = default;
    FSElem& operator=(const FSElem&) = default;
    FSElem& operator=(FSElem&&) = default;
    ~FSElem() = default;

    FSElem(std::string name, std::size_t bytes, bool is_folder);
    const std::string& name() const;
    std::size_t size_bytes() const;
    bool is_folder() const;

    private:
    std::string name_;
    std::size_t size_bytes_;
    bool is_folder_;
};

class FileTreeNode
{
    public:
    using directory_t = std::unordered_map<std::string, std::unique_ptr<FileTreeNode>>;
    using file_meta_t = models::Mapping;
    using data_t = std::variant<directory_t, file_meta_t>;
    
    FileTreeNode() = delete;
    FileTreeNode(const FileTreeNode&) = delete;
    FileTreeNode(FileTreeNode&&) = default;
    FileTreeNode& operator=(const FileTreeNode&) = delete;
    FileTreeNode& operator=(FileTreeNode&&) = default;
    ~FileTreeNode() = default;

    FileTreeNode(std::string_view name, data_t data);
    FileTreeNode* get(std::string_view path);
    FileTreeNode* add(std::string_view path, const models::Mapping* mapping = nullptr);

    void print() const;
    bool is_folder() const;
    const data_t& elem() const;
    const std::string& name() const;
    directory_t& as_directory();
    file_meta_t& as_filemeta();
    const directory_t& as_cdirectory() const;
    const file_meta_t& as_cfilemeta() const;

    private:
    void print_rec(const FileTreeNode* start, int level) const;
    std::string name_;
    data_t data_;
};

}

class FSMirror
{
    public:
    enum class Error
    {
        Ok = 0,
        AlreadyExists = 1,
        CannotLinkInServerFolder = 2,
        CannotAddInLinkedFolder = 3,
        NotFound = 4
    };

    using info_result_t = util::Expected<detail::FSElem, Error>;
    using list_result_t = util::Expected<std::vector<detail::FSElem>, Error>;

    FSMirror();
    FSMirror(const FSMirror&) = delete;
    FSMirror(FSMirror&&) noexcept = default;
    FSMirror& operator=(const FSMirror&) = delete;
    FSMirror& operator=(FSMirror&&) noexcept = default;
    ~FSMirror() = default;

    Error mkdir(std::string_view path);
    Error add_file(std::string_view server, std::string_view ref, std::size_t size_bytes);
    Error link_file(std::string_view server, std::string_view ref, std::string path);
    Error reset(const std::vector<models::Mapping>& mappings);
    list_result_t ls(std::string_view path);
    info_result_t info(std::string_view path);

    private:
    detail::FileTreeNode root_;
};

//template<
//    typename K,
//    typename V,
//    template<typename...> class C = std::unordered_map>
//class Node
//{
//    public:
//    using container_t = C<K,Node>;
//    using path_t = std::vector<K>;
//    using data_t = std::variant<V, container_t, detail::not_set>;
//
//    enum class InsertResult 
//    {
//        OK = 0,
//        CannotWrite,
//        CannotTraverse
//    };
//    using insert_error_t = std::pair<InsertResult, std::size_t>;
//
//    enum class NodeStatus
//    {
//        Folder = 0,
//        NotFound,
//        Empty
//    };
//    using get_result_t = Expected<V, NodeStatus>;
//
//
//    Node();
//    Node(V);
//    Node(container_t);
//    Node(const Node&) = delete;
//    Node(Node&&) noexcept = default;
//    Node& operator=(const Node&) = delete;
//    Node& operator=(Node&&) = delete;
//    ~Node() = default;
//
//    insert_error_t insert(const path_t& path, V item);
//    get_result_t get(const path_t& path);
//
//    private:
//    data_t contents_;
//    insert_error_t insert(const path_t& path, std::size_t idx, data_t *current, V value);
//    get_result_t get(const path_t& path, std::size_t idx, const data_t* current);
//
//};
//
//template<typename K, typename V, template<typename...> class C>
//Node<K, V, C>::Node()
//    : contents_{detail::not_set{}}
//{}
//
//template<typename K, typename V, template<typename...> class C>
//Node<K, V, C>::Node(V val)
//    : contents_{std::move(val)}
//{}
//
//template<typename K, typename V, template<typename...> class C>
//Node<K, V, C>::Node(container_t c)
//    : contents_{std::move(c)}
//{}
//
//template<typename K, typename V, template<typename...> class C>
//typename Node<K, V, C>::insert_error_t Node<K, V, C>::insert(const path_t& path, V value)
//{
//    return insert(path, 0, &contents_, std::move(value));
//}
//
//template<typename K, typename V, template<typename...> class C>
//typename Node<K, V, C>::get_result_t Node<K, V, C>::get(const path_t& path)
//{
//    return get(path, 0, &contents_);
//}
//
//template<typename K, typename V, template<typename...> class C>
//typename Node<K, V, C>::insert_error_t Node<K, V, C>::insert(const path_t& path, std::size_t idx, data_t* current, V value)
//{
//    assert(current); // should never be null
//    if (idx == path.size()) { // we need to insert here
//        if (std::holds_alternative<V>(*current) || std::holds_alternative<detail::not_set>(*current)) {
//            *current = std::move(value);
//            return {InsertResult::OK, idx};
//        }
//        return {InsertResult::CannotWrite, idx}; // we expect a value and have a non-empty subfolder 
//    }
//
//    if (std::holds_alternative<V>(*current)) { // need to keep traversing folders and have a value
//        return {InsertResult::CannotTraverse, idx};
//    }
//
//    if (std::holds_alternative<detail::not_set>(*current)) { // no value set, turn in into a map of nodes
//        *current = container_t{};
//    }
//
//    auto& subtree{std::get<container_t>(*current)};
//    return insert(path, idx+1, &subtree[path[idx]], std::move(value));
//}
//
//template<typename K, typename V, template<typename...> class C>
//typename Node<K, V, C>::get_result_t Node<K, V, C>::get(const path_t& path, std::size_t idx, const data_t* current)
//{
//    if (idx == path.size()) { // this is what we're looking for
//    }
//}
//
//
//template<typename K, typename V, template<typename...> class C>
//bool can_write(const typename Node<V, V, C>::data_t& data) {
//    using node_ptr_t = typename Node<V, V, C>::node_ptr_t;
//    return std::holds_alternative<V>(data) || std::holds_alternative<detail::not_set>(data);
//}
//
//} // namespace mifs::util
}
#endif
