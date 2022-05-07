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

    Mapping(std::string name, std::size_t size_bytes);

    template<typename Serialized>
    static parse_result_t parse(const Serialized& data);

    const std::string& name() const;
    std::size_t size_bytes() const;

    private:
    std::string name_;
    std::size_t size_bytes_;
};

} // namespace mifs::models

#endif
