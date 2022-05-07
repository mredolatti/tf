#include "isclient.hpp"
#include "http.hpp"
#include "jsend.hpp"
#include "expected.hpp"
#include "mappings.hpp"
#include <iostream>


namespace mifs::apiclients {

namespace detail {

using parse_result_t = util::Expected<IndexServerClient::response_t, int /* TODO */>;
using parse_error_t = util::Unexpected<int /* TODO */>;

}

IndexServerClient::IndexServerClient(http_client_ptr_t http_client) :
    client_{std::move(http_client)}
{}

IndexServerClient::response_result_t IndexServerClient::get_all()
{
    auto request{http::Request::Builder{}
        .method(http::Method::GET)
        .uri("http://index-server:9876/mappings")
        .build()};

    auto result{client_->execute(request)};
    if (!result) {
        return no_response_t{-1};
    }

    auto code{(*result).code()};
    if ( code != 200) {
        std::cout << "code: " << code << '\n';
        //std::cout << "body: " << (*result).body() << '\n';
        return no_response_t{code};
    }

    return jsend::parse<models::Mapping>((*result).body(), "mapping");
}


} // namespace mifs::apiclients
