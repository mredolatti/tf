#include "apierror.hpp"

namespace mifs::apiclients
{

Error::Error(http::Client::Error err)
    : data_{err}
{
}

Error::Error(int code, std::string_view message)
    : data_{std::make_pair(code, std::string{message})}
{
}

int64_t Error::get() const
{
    return std::holds_alternative<http_jsend_err>(data_) ? std::get<http_jsend_err>(data_).first
                                                         : std::get<http::Client::Error>(data_).get();
}

const char *Error::message() const
{
    return std::holds_alternative<http_jsend_err>(data_) ? std::get<http_jsend_err>(data_).second.c_str()
                                                         : std::get<http::Client::Error>(data_).message();
}

namespace predefined
{
const Error token_error{-1, "failed to get token"};
const Error json_error{-2, "failed parse successful response JSON body"};
const Error json_error_unsuccessful{-3, "failed parse error response JSON body"};
const Error no_server_data{-4, "associate information for org/server not available in cache"};
} // namespace predefined

} // namespace mifs::apiclients
