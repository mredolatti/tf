#ifndef MIFS_APICLIENT_ERRORS_HPP
#define MIFS_APICLIENT_ERRORS_HPP

#include "http.hpp"
#include "httpc.hpp"
#include <variant>

namespace mifs::apiclients
{



class Error
{
  public:
    Error() = delete;
    Error(http::Client::Error err);
    Error(int code, std::string_view message);
    Error(const Error&) = default;
    Error& operator=(const Error&) = default;
    Error(Error&&) noexcept = default;
    Error& operator=(Error&&) noexcept = default;
    ~Error() = default;

    int64_t get() const;
    const char *message() const;

  private:
    using http_jsend_err = std::pair<int, std::string>;
    using data_t = std::variant<http::Client::Error, http_jsend_err>;
    data_t data_;
};

namespace predefined 
{

extern const Error token_error;
extern const Error json_error;
extern const Error json_error_unsuccessful;
extern const Error no_server_data;
} // namespace predefined

} // namespace mifs::apiclients

#endif
