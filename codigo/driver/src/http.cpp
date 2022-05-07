#include "http.hpp"

namespace mifs::http {

Request::Builder& Request::Builder::method(Method m)    { method_ = m; return *this; }
Request::Builder& Request::Builder::body(std::string b) { body_ = std::move(b); return *this; }
Request::Builder& Request::Builder::uri(std::string u)  { uri_ = std::move(u); return *this; }
Request::Builder& Request::Builder::headers(Headers h)  { headers_ = std::move(h); return *this; }
Request::Builder& Request::Builder::tls(tls::Config c)  { tls_conf_ = std::move(c); return *this; }

Request Request::Builder::build()
{
    return Request{method_, std::move(body_), std::move(uri_), std::move(headers_), std::move(tls_conf_)}; 
}

Request:: Request(Method m, std::string body, std::string uri, Headers headers, std::optional<tls::Config> tls_conf) :
    method_{m},
    body_{std::move(body)},
    uri_{std::move(uri)},
    headers_{std::move(headers)},
    tls_conf_{std::move(tls_conf)}
{}

Method Request::method() const
{ 
    return method_; 
}

const std::string& Request::url() const
{
    return uri_;
}


const std::string& Request::body() const
{
    return body_;
}

const Headers& Request::headers() const  
{
    return headers_; 
}

const std::optional<tls::Config>& Request::tls_conf() const
{
    return tls_conf_;
}

// Response

Response::Builder& Response::Builder::code(int code) { code_ = code; return *this; }
Response::Builder& Response::Builder::add_header(std::string name, std::string value) { headers_[name] = std::move(value); return *this; }
Response::Builder& Response::Builder::body_append(std::string data) { body_builder_ << data; return *this; }
Response::Builder& Response::Builder::body_append(const char* data) { body_builder_ << data; return *this; }
Response::Builder& Response::Builder::body_append(const char* data, std::size_t len) { body_builder_ << std::string{data, len}; return *this; }

Response Response::Builder::build()
{
    return Response{code_, body_builder_.str(), std::move(headers_)};
}

Response::Response(int c, std::string b, Headers h) :
    code_{c},
    body_{std::move(b)},
    headers_{std::move(h)}
{}

int Response::code() const
{
    return code_;
}

const std::string Response::body() const
{
    return body_;
}

const Headers& Response::headears() const
{
    return headers_;
}

} // namespace mifs::http

