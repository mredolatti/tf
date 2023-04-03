#include "httpc.hpp"
#include <curl/curl.h>
#include <iostream>
#include <charconv>

namespace mifs::httpc::detail{

int code_from_status_line(std::string_view line)
{
    auto first_space{line.find(' ')};
    if (first_space == std::string_view::npos) {
        return -1;
    }

    auto sc_sv{line.substr(first_space + 1, 3)};
    int status_code;
    auto [_, ec]{std::from_chars(sc_sv.data(), sc_sv.data() + sc_sv.size(), status_code)};
    if (ec == std::errc{}) {
        return status_code;
    }
    return -1;
}

class CurlStaticInitializer
{
    public:
    CurlStaticInitializer() { curl_global_init(CURL_GLOBAL_DEFAULT); }
    CurlStaticInitializer(CurlStaticInitializer&&) = delete;
    CurlStaticInitializer& operator=(const CurlStaticInitializer&) = delete;
    CurlStaticInitializer& operator=(CurlStaticInitializer&&) = delete;
    ~CurlStaticInitializer() { curl_global_cleanup(); }
};

class CurlEasyHandle
{
    public:
    CurlEasyHandle() : handle_{curl_easy_init()} {}
    CurlEasyHandle(const CurlEasyHandle&) = delete;
    CurlEasyHandle(CurlEasyHandle&&) = delete;
    CurlEasyHandle& operator=(const CurlEasyHandle&) = delete;
    CurlEasyHandle& operator=(CurlEasyHandle&&) = delete;
    ~CurlEasyHandle() { curl_easy_cleanup(handle_); }

    CURL* operator*() { return handle_; }

    private:
    CURL* handle_;
};

class CurlHeaders
{
    public:
    CurlHeaders() : headers_{nullptr} {}
    CurlHeaders(const CurlHeaders&) = delete;
    CurlHeaders(CurlHeaders&&) = delete;
    CurlHeaders& operator=(const CurlHeaders&) = delete;
    CurlHeaders& operator=(CurlHeaders&&) = delete;
    ~CurlHeaders() { curl_slist_free_all(headers_); }

    curl_slist* operator*() { return headers_; }
    void add(const char* header_line) { headers_ = curl_slist_append(headers_, header_line); }

    private:
    curl_slist* headers_;
};

static CurlStaticInitializer curl_static_initializer;

}

extern "C"
{
    static size_t write_cb(char *ptr, size_t size, size_t nmemb, void *data)
    {
        using context_t = mifs::http::detail::Context;
        auto realsize{size * nmemb};
        auto* const context{reinterpret_cast<context_t*>(data)};
        if (context != nullptr) {
            // TODO(mredolatti): LOG!
            context->response.body_append(ptr, realsize);
        }
        return realsize;
    } 

    static int trace_cb(CURL*, curl_infotype , char* data, size_t, void* userptr)
    {
        using context_t = mifs::http::detail::Context;
        auto* const context{reinterpret_cast<context_t*>(userptr)};
        // TODO(mredolatti): LOG!
        return 0;
    }

    static size_t header_callback(const char *buffer, size_t size, size_t nitems, void *userdata)
    {
        using context_t = mifs::http::detail::Context;
        auto total_size{size*nitems};
        if (total_size == 2 && buffer[0] == '\r' && buffer[1] == '\n') {
            return total_size;
        }

        auto* const context{reinterpret_cast<context_t*>(userdata)};
        if (context == nullptr) {
            std::cout << "context es null\n";
            std::abort();
        }
        auto line{std::string_view(buffer, total_size)};
        const auto separator_idx{line.find(':')};
        if (separator_idx == -1) {
            //TODO(mredolatti): LOG
            auto sc{mifs::httpc::detail::code_from_status_line(line)};
            context->response.code(sc);
            return total_size;
        }

        // We have a header:
        auto name{std::string(buffer, separator_idx)};
        auto value{std::string(buffer + separator_idx + 2 , line.size() - separator_idx - 4)}; // skipping ": " and "\r\n"
        // TODO(mredolatti): LOG!
        context->response.add_header(std::move(name), std::move(value));
        return total_size;
    }
}

namespace mifs::http {

MaybeResponse Client::execute(const Request& request)
{
    httpc::detail::CurlEasyHandle handle;

    // callbacks
    curl_easy_setopt(*handle, CURLOPT_WRITEFUNCTION, write_cb);
    curl_easy_setopt(*handle, CURLOPT_HEADERFUNCTION, header_callback);
    curl_easy_setopt(*handle, CURLOPT_DEBUGFUNCTION, trace_cb);
    curl_easy_setopt(*handle, CURLOPT_VERBOSE, 1L);

    // url
    curl_easy_setopt(*handle, CURLOPT_URL, request.url().c_str());

    // tls
    if (request.tls_conf()) {
        const auto& cfg{request.tls_conf().value()};
        curl_easy_setopt(*handle, CURLOPT_SSL_VERIFYPEER, 1L);
        curl_easy_setopt(*handle, CURLOPT_SSLCERTTYPE, "PEM");
        curl_easy_setopt(*handle, CURLOPT_CAINFO, cfg.root_ca_fn().empty() ? nullptr : cfg.root_ca_fn().c_str());
        curl_easy_setopt(*handle, CURLOPT_SSLKEY, cfg.client_pk_fn().empty() ? nullptr : cfg.client_pk_fn().c_str());    
        curl_easy_setopt(*handle, CURLOPT_SSLCERT, cfg.client_cert_fn().empty() ? nullptr : cfg.client_cert_fn().c_str());
    }

    // headers
    httpc::detail::CurlHeaders headers;
    for (const auto& [name, value]: request.headers()) {
        std::string tmp{name + ": " + value};
        headers.add(tmp.c_str());
    }
    curl_easy_setopt(*handle, CURLOPT_HTTPHEADER, *headers);

    if (!request.body().empty()) {
        curl_easy_setopt(*handle, CURLOPT_POSTFIELDS, request.body().c_str());
    }

    switch (request.method()) {
        case Method::PATCH: curl_easy_setopt(*handle, CURLOPT_CUSTOMREQUEST, "PATCH"); break;
        case Method::PUT:   curl_easy_setopt(*handle, CURLOPT_CUSTOMREQUEST, "PUT");   break;
        case Method::POST:  curl_easy_setopt(*handle, CURLOPT_CUSTOMREQUEST, "POST");  break;
        default: ;
    }

    // contexto
    detail::Context context{.request = request};
    curl_easy_setopt(*handle, CURLOPT_PRIVATE, &context);
    curl_easy_setopt(*handle, CURLOPT_HEADERDATA, &context);
    curl_easy_setopt(*handle, CURLOPT_WRITEDATA, &context);

    auto res{curl_easy_perform(*handle)};
    if (res != CURLE_OK) {
        return NoResponse{res};
    }

    return context.response.build();
}

}
