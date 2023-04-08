#include "fselems.hpp"

namespace mifs::types
{

FSElem::FSElem(std::string name)
    : name_{std::move(name)}
{
}

const std::string& FSElem::name() const { return name_; }

// ---------------------------------------

FSEFile::FSEFile(std::string name, std::string_view org, std::string_view server, std::string_view ref,
                 std::size_t size_bytes, int64_t last_updated)
    : FSElem{std::move(name)},
      org_{org},
      server_{server},
      ref_{std::move(ref)},
      size_bytes_{size_bytes},
      last_updated_(last_updated)
{
}

const std::string& FSEFile::org() const { return org_; }

const std::string& FSEFile::server() const { return server_; }

const std::string& FSEFile::ref() const { return ref_; }

std::size_t FSEFile::size_bytes() const { return size_bytes_; }

int64_t FSEFile::last_updated() const { return last_updated_; }

void FSEFile::accept(FSElemVisitor& v) const { v.visit_file(*this); }

// -------------------------------------------

FSEFolder::FSEFolder(std::string name)
    : FSElem{std::move(name)}
{
}

std::size_t FSEFolder::size_bytes() const { return 0; }

void FSEFolder::accept(FSElemVisitor& v) const { v.visit_folder(*this); }

// -------------------------------------------

FSELink::FSELink(std::string_view id, std::string_view name, std::string_view org_name,
                 std::string_view server_name, std::string_view ref)
    : FSElem{std::string{name}},
      mapping_id_{id},
      org_name_{org_name},
      server_name_{server_name},
      ref_{ref}
{
}

const std::string& FSELink::id() const { return mapping_id_; }

const std::string& FSELink::org_name() const { return org_name_; }

const std::string& FSELink::server_name() const { return server_name_; }

const std::string& FSELink::ref() const { return ref_; }

std::size_t FSELink::size_bytes() const { return 0; }

void FSELink::accept(FSElemVisitor& v) const { v.visit_link(*this); }

} // namespace mifs::types
