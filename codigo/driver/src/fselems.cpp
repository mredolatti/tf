#include "fselems.hpp"

namespace mifs::types {

FSElem::FSElem(std::string name) :
    name_{std::move(name)}
{}

const std::string& FSElem::name() const
{
    return name_;
}

// ---------------------------------------

FSEFile::FSEFile(std::string name, std::string server_id, std::string ref, std::size_t size_bytes) :
    FSElem{std::move(name)},
    server_id_{std::move(server_id)},
    ref_{std::move(ref)},
    size_bytes_{size_bytes}
{}

const std::string& FSEFile::server_id() const
{
    return server_id_;
}

const std::string& FSEFile::ref() const
{
    return ref_;
}

std::size_t FSEFile::size_bytes() const
{
    return size_bytes_;
}

void FSEFile::accept(FSElemVisitor& v) const
{
    v.visit_file(*this);
}


// -------------------------------------------

FSEFolder::FSEFolder(std::string name) :
    FSElem{std::move(name)}
{}

std::size_t FSEFolder::size_bytes() const
{
    // TODO: is this ok?
    return 0;
}

void FSEFolder::accept(FSElemVisitor& v) const
{
    v.visit_folder(*this);
}

// -------------------------------------------

FSELink::FSELink(std::string name, std::string server_id, std::string ref) :
    FSElem{std::move(name)},
    server_id_{std::move(server_id)},
    ref_{std::move(ref)}
{}

std::size_t FSELink::size_bytes() const
{
    return 0;
}

const std::string& FSELink::server_id() const
{
    return server_id_;
}

const std::string& FSELink::ref() const
{
    return ref_;
}

void FSELink::accept(FSElemVisitor& v) const
{
    v.visit_link(*this);
}

} // namespace mifs::types
