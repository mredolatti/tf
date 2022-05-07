#ifndef MIFS_HTTP_HPP
#define MIFS_HTTP_HPP

#include "tls.hpp"
#include <optional>
#include <string>
#include <sstream>
#include <unordered_map>

namespace mifs::http {

enum class Method {GET, POST, PUT, PATCH, DELETE};

using Headers = std::unordered_map<std::string, std::string>;

class Request
{
    public:
    class Builder
    {
        public:
        Builder() = default;
        Builder(const Builder&) = default;
        Builder(Builder&&) = default;
        Builder& operator=(const Builder&) = default;
        Builder& operator=(Builder&&) = default;
        ~Builder() = default;

        Builder& method(Method m);
        Builder& body(std::string b);
        Builder& uri(std::string u);
        Builder& headers(Headers h);
        Builder& tls(tls::Config tls_conf);
        Request build();

        private:
        Method method_;
        std::string body_;
        std::string uri_;
        Headers headers_;
        std::optional<tls::Config> tls_conf_;
    };

    Request() = delete;
    Request(const Request&) = default;
    Request(Request&&) = default;
    Request& operator=(const Request&) = default;
    Request& operator=(Request&&) = default;
    ~Request() = default;

    Request(Method m, std::string body, std::string uri, Headers headers, std::optional<tls::Config> tls_conf);
    Method method() const;
    const std::string& url() const;
    const std::string& body() const;
    const Headers& headers() const;
    const std::optional<tls::Config>& tls_conf() const;

    private:
    Method method_;
    std::string body_;
    std::string uri_;
    Headers headers_;
    std::optional<tls::Config> tls_conf_;
};

class Response
{
    public:
    
    class Builder
    {
        public:
        Builder() = default;
        Builder(const Builder&) = delete;
        Builder(Builder&&) noexcept = default;
        Builder& operator=(const Builder&) = delete;
        Builder& operator=(Builder&&) noexcept = default;
        ~Builder() noexcept = default;

        Builder& code(int code);
        Builder& add_header(std::string name, std::string value);
        Builder& body_append(std::string data);
        Builder& body_append(const char* data);
        Builder& body_append(const char* data, std::size_t len);
        Response build();

        private:
        int code_;
        std::stringstream body_builder_;
        Headers headers_;
    };

    Response() = delete;
    Response(const Response&) = default;
    Response(Response&&) = default;
    Response& operator=(const Response&) = default;
    Response& operator=(Response&&) = default;
    ~Response() = default;

    Response(int c, std::string b, Headers h);

    int code() const;
    const std::string body() const;
    const Headers& headears() const;

    private:
    int code_;
    std::string body_;
    Headers headers_;
};

}

#endif
