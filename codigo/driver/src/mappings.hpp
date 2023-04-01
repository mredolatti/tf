#ifndef MIFS_MAPPINGS_HPP
#define MIFS_MAPPINGS_HPP

#include <string>

#include "expected.hpp"

namespace mifs::models {

class Mapping
{
    
    public:

    using parse_result_t = util::Expected<Mapping, int /* TODO */>;

    Mapping() = delete;
    Mapping(const Mapping&) = default;
    Mapping(Mapping&&) = default;
    Mapping& operator=(const Mapping&) = default;
    Mapping& operator=(Mapping&&) = default;
    ~Mapping() = default;

    Mapping(std::string_view name, std::size_t size_bytes, std::string_view ref, std::string_view org, std::string_view server, int64_t last_updated);

    template<typename Serialized>
    static parse_result_t parse(const Serialized& data);

    const std::string& path() const;
    std::size_t size_bytes() const;
    const std::string& ref() const;
    const std::string& org() const;
    const std::string& server() const;
    int64_t last_updated() const;

    private:
    std::string path_;
    std::string ref_;
    std::string org_;
    std::string server_;
    int64_t last_updated_;
    std::size_t size_bytes_;
};

} // namespace mifs::models

#endif
