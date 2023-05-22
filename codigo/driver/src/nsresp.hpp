#ifndef MIFS_NSRESP_HPP
#define MIFS_NSRESP_HPP

#include <string_view>
#include <string>

namespace mifs::nsresp {

class TokenResponse
{
    public:
    TokenResponse() = delete;
    TokenResponse(std::string_view token);
    TokenResponse(const TokenResponse&) = default;
    TokenResponse(TokenResponse&&) noexcept = default;
    TokenResponse& operator=(const TokenResponse&) = default;
    TokenResponse& operator=(TokenResponse&&) noexcept = default;
    ~TokenResponse() = default;;

    static TokenResponse parse(std::string_view body);

    const std::string& token() const;

    private:
    std::string token_;
};

} // namespace mifs::nsresp

#endif
