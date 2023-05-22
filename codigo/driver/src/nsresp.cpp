#include "nsresp.hpp"

#include <rapidjson/document.h>

namespace mifs::nsresp
{

TokenResponse::TokenResponse(std::string_view token)
    : token_{token}
{
}

TokenResponse TokenResponse::parse(std::string_view body)
{

    rapidjson::Document doc;
    doc.Parse(body.data());

    if (!doc.HasMember("token") || !doc["token"].IsString()) {
        // TODO
    }

    return TokenResponse{doc["token"].GetString()};
}

const std::string& TokenResponse::token() const { return token_; }

} // namespace mifs::nsresp
