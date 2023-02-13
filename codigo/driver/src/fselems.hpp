#ifndef MIFS_TYPES_FSELEM_HPP
#define MIFS_TYPES_FSELEM_HPP

#include <string>
namespace mifs::types {

class FSElemVisitor;

class FSElem
{
    public:
    FSElem() = delete;
    FSElem(const FSElem&) = default;
    FSElem(FSElem&&) = default;
    FSElem& operator=(const FSElem&) = default;
    FSElem& operator=(FSElem&&) = default;
    ~FSElem() = default;

    FSElem(std::string name);
    const std::string& name() const;

    virtual std::size_t size_bytes() const = 0;
    virtual void accept(FSElemVisitor&) const = 0;

    private:
    std::string name_;
};

class FSEFile : public FSElem
{
    public:
    FSEFile(const FSEFile&) = default;
    FSEFile(FSEFile&&) = default;
    FSEFile& operator=(const FSEFile&) = default;
    FSEFile& operator=(FSEFile&&) = default;
    ~FSEFile() = default;

    FSEFile(std::string name, std::string server_id, std::string ref, std::size_t size_bytes);
    const std::string& server_id() const;
    const std::string& ref() const;
    std::size_t size_bytes() const override;
    void accept(FSElemVisitor&) const override;

    private:
    std::string server_id_;
    std::string ref_;
    std::size_t size_bytes_;

};

class FSEFolder : public FSElem
{
    public:
    FSEFolder() = delete;
    FSEFolder(const FSEFolder&) = default;
    FSEFolder(FSEFolder&&) = default;
    FSEFolder& operator=(const FSEFolder&) = default;
    FSEFolder& operator=(FSEFolder&&) = default;
    ~FSEFolder() = default;

    FSEFolder(std::string name);
    std::size_t size_bytes() const override;
    void accept(FSElemVisitor&) const override;
};

class FSELink : public FSElem
{
    public:
    FSELink() = delete;
    FSELink(const FSELink&) = default;
    FSELink(FSELink&&) = default;
    FSELink& operator=(const FSELink&) = default;
    FSELink& operator=(FSELink&&) = default;
    ~FSELink() = default;

    FSELink(std::string name, std::string server_id, std::string ref);
    std::size_t size_bytes() const override;
    void accept(FSElemVisitor&) const override;
    const std::string& server_id() const;
    const std::string& ref() const;

    private:
    std::string server_id_;
    std::string ref_;
};


class FSElemVisitor
{
    public:
    virtual void visit_file(const FSEFile&) = 0;
    virtual void visit_link(const FSELink&) = 0;
    virtual void visit_folder(const FSEFolder&) = 0;
};

} // namespace mifs::types

#endif /* ifndef MIFS_TYPES_FSELEM_HPP */