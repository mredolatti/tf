#ifndef MIFS_IS_CLIENT_HPP
#define MIFS_IS_CLIENT_HPP

#include "http.hpp"
#include "httpc.hpp"
#include "mappings.hpp"
#include "jsend.hpp"
#include "expected.hpp"

#include <memory>
#include <vector>

namespace mifs::apiclients {

class IndexServerClient
{
    public:

    struct Config
    {
        std::string url;
    };

    using http_client_ptr_t = std::shared_ptr<http::Client>;

    using response_t = jsend::Response<models::Mapping>;
    using response_result_t = util::Expected<response_t, int /* TODO */>;
    using no_response_t = util::Unexpected<int /* TODO */>;

    IndexServerClient() = delete;
    IndexServerClient(const IndexServerClient&) = delete;
    IndexServerClient(IndexServerClient&&) = default;
    IndexServerClient& operator=(const IndexServerClient&) = delete;
    IndexServerClient& operator=(IndexServerClient&&) = delete;
    ~IndexServerClient() = default;

    explicit IndexServerClient(http_client_ptr_t http_client, Config config);

    response_result_t get_all();

    private:
    http_client_ptr_t client_;
    std::string url_;
};

} // namespace mifs::apiclients

#endif
