#ifndef MIFS_HTTP_CLIENT_HPP
#define MIFS_HTTP_CLIENT_HPP

#include "expected.hpp"
#include "http.hpp"

namespace mifs::http
{

namespace detail
{

using curl_handle_t = void *;

struct Context {
    const Request& request; // reference to the one passed by the client
    Response::Builder response;
};

using contexts_t = std::unordered_map<curl_handle_t, Context>;
} // namespace detail

class Client
{
  public:

    class Error
    {
        public:
        Error() = delete;
        Error(int64_t code);
        Error(const Error&) = default;
        Error& operator=(const Error&) = default;
        Error(Error&&) noexcept = default;
        Error& operator=(Error&&) noexcept = default;
        ~Error() = default;

        int64_t get() const;
        const char* message() const;

        private:
        int64_t code_;
    };


    using response_t = util::Expected<Response, Error>;
    using no_response_t = util::Unexpected<Error>;

    Client() = default;
    Client(const Client&) = delete;
    Client(Client&&) noexcept = default;
    Client& operator=(Client&&) noexcept = default;
    ~Client() = default;

    response_t execute(const Request& request);

  private:
    detail::contexts_t contexts_;
};

} // namespace mifs::http

#endif
