#ifndef MIFS_ERRORS_HPP
#define MIFS_ERRORS_HPP

#include <system_error>

namespace mifs::http {

enum class Error
{
    ConnectionRefused = 10,
    DNSFailed,
    TLSErr,
};

std::error_code make_error_code(Error);

} // namespace mifs::http

namespace std {

template<> struct is_error_code_enum<mifs::http::Error> : true_type {};

}

#endif
