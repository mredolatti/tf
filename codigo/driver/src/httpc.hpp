#ifndef MIFS_HTTP_CLIENT_HPP
#define MIFS_HTTP_CLIENT_HPP

#include "http.hpp"
#include "expected.hpp"

namespace mifs::http {

using MaybeResponse = util::Expected<Response, int>;
using NoResponse = util::Unexpected<int>;

namespace detail {

    using curl_handle_t = void*;

    struct Context
    {
        const Request& request; // reference to the one passed by the client
        Response::Builder response;
    };

    using contexts_t = std::unordered_map<curl_handle_t, Context>;
}

class Client
{
    public:
    Client() = default;
    Client(const Client&) = delete;
    Client(Client&&) noexcept = default;
    Client& operator=(Client&&) noexcept = default;
    ~Client() = default;

    MaybeResponse execute(const Request& request);

    private:
    detail::contexts_t contexts_;
};

}

#endif
