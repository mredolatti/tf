#include "errors.hpp"

namespace { // anonymous namespace
 
struct HttpErrorCategory : std::error_category
{
    const char* name() const noexcept override;
    std::string message(int ev) const override;
};
 
const char* HttpErrorCategory::name() const noexcept
{
  return "http_error";
}
 
std::string HttpErrorCategory::message(int ev) const
{
  switch (static_cast<mifs::http::Error>(ev))
  {
    case mifs::http::Error::ConnectionRefused:  return "connection refused by remote host";
    case mifs::http::Error::DNSFailed:          return "dns resolution failed";
    case mifs::http::Error::TLSErr:             return "TLS connection error";
    default:                                    return "unrecognized error";
  }
}
 
const HttpErrorCategory httpErrorCategory {};
 
}

namespace mifs::http {

std::error_code make_error_code(Error e)
{
    return {static_cast<int>(e), httpErrorCategory};
}


} // namespace mifs::http
