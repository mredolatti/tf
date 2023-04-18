#include "jsend.hpp"

#include <sstream>

namespace mifs::jsend
{
namespace detail
{

Status parse_status(std::string_view status)
{
    if ("success" == status) {
        return Status::SUCCESS;
    } else if ("error" == status) {
        return Status::ERROR;
    } else if ("fail" == status) {
        return Status::FAILURE;
    }

    return Status::ERROR;
}

} // namespace detail

util::Expected<ErrorResponse, JSONParseStatus> parse_unsuccessful_response(std::string_view body)
{
    rapidjson::Document doc;
    doc.Parse(body.data());

    if (!doc.HasMember("status") || !doc["status"].IsString()) {
        return util::Unexpected<JSONParseStatus>{JSONParseStatus::Error};
    }

    ErrorResponse toRet{.status = detail::parse_status(doc["status"].GetString())};

    if (doc.HasMember("message") && doc["message"].IsString()) {
        toRet.message = doc["message"].GetString();
    }

    if (doc.HasMember("code") && doc["code"].IsString()) {
        toRet.code = doc["code"].GetString();
    }

    if (!doc.HasMember("data") || !doc["data"].IsObject()) {
        return toRet;
    }

    const auto& data{doc["data"].GetObject()};
    for (auto const& item : data) {
        if (!item.name.IsString() || !item.value.IsString()) {
            continue;
        }
        toRet.data[item.name.GetString()] = item.value.GetString();
    }
    return toRet;
}

std::string format_error(const ErrorResponse& r)
{
    switch (r.status) {
    case Status::ERROR: {
        if (r.data.empty()) {
            return r.message;
        }

        std::stringstream ss;
        ss << r.message << ":\n\n";
        for (const auto& item : r.data) {
            ss << fmt::format("- [{}]: {}\n", item.first, item.second);
        }
        ss << '\n';
        return ss.str();
    }
    case Status::FAILURE: {
        auto it{r.data.begin()};
        if (it == r.data.end()) {
            return "";
        }
        return fmt::format("error in field/parameter '{}': {}", it->first, it->second);
    }
    case Status::SUCCESS:
    default: return "";
    }
}

} // namespace mifs::jsend
