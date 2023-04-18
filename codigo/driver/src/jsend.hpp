#ifndef MIFS_JSEND_HPP
#define MIFS_JSEND_HPP

#include "expected.hpp"

#include <fmt/format.h>
#include <iostream>
#include <string>
#include <unordered_map>
#include <vector>

#include <rapidjson/document.h>

namespace mifs::jsend
{

enum class JSONParseStatus { OK = 0, Error = 1 };

enum class Status { SUCCESS, FAILURE, ERROR };

namespace detail
{

Status parse_status(std::string_view status);

} // namespace detail

// Base template
template <typename T>
struct BaseResponse {
    using data_t = std::unordered_map<std::string, T>;

    Status status;
    std::string code;
    std::string message;
    data_t data;
};

// Response types
template <typename T>
using SingleItemResponse = BaseResponse<T>;

template <typename T>
using MultipleItemResponse = BaseResponse<std::vector<T>>;

using ErrorResponse = BaseResponse<std::string>;

template <typename T>
util::Expected<MultipleItemResponse<T>, JSONParseStatus> parse_multi_item_response(std::string_view body,
                                                                                   std::string_view resource);
template <typename T>
util::Expected<SingleItemResponse<T>, JSONParseStatus> parse_single_item_response(std::string_view body,
                                                                                  std::string_view resource);

util::Expected<ErrorResponse, JSONParseStatus> parse_unsuccessful_response(std::string_view body);

// helpers

template <typename T>
util::Expected<MultipleItemResponse<T>, JSONParseStatus> parse_multi_item_response(std::string_view body,
                                                                                   std::string_view resource)
{
    rapidjson::Document doc;
    doc.Parse(body.data());

    if (!doc.HasMember("status") || !doc["status"].IsString()) {
        return util::Unexpected<JSONParseStatus>{JSONParseStatus::Error};
    }

    MultipleItemResponse<T> toRet{.status = detail::parse_status(doc["status"].GetString())};

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
    if (!data.HasMember(resource.data()) || !data[resource.data()].IsArray()) {
        return util::Unexpected<JSONParseStatus>(JSONParseStatus::Error);
    }

    const auto& items{data[resource.data()].GetArray()};
    std::vector<T> parsed_items;
    parsed_items.reserve(items.Size());
    for (auto& item : items) {
        auto parse_result{T::parse(item)};
        if (!parse_result) {
            return util::Unexpected<JSONParseStatus>{JSONParseStatus::Error};
        }
        parsed_items.emplace_back(*parse_result);
    }

    typename MultipleItemResponse<T>::data_t::value_type map_data{resource, std::move(parsed_items)};
    toRet.data = typename MultipleItemResponse<T>::data_t{std::move(map_data)};
    return toRet;
}

template <typename T>
util::Expected<SingleItemResponse<T>, JSONParseStatus> parse_single_item_response(std::string_view body,
                                                                                  std::string_view resource)
{
    rapidjson::Document doc;
    doc.Parse(body.data());

    if (!doc.HasMember("status") || !doc["status"].IsString()) {
        return util::Unexpected<JSONParseStatus>{JSONParseStatus::Error};
    }

    SingleItemResponse<T> toRet{.status = detail::parse_status(doc["status"].GetString())};

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
    if (!data.HasMember(resource.data()) || !data[resource.data()].IsObject()) {
        return util::Unexpected<JSONParseStatus>(JSONParseStatus::Error);
    }

    auto parse_result{T::parse(data[resource.data()])};
    if (!parse_result) {
        return util::Unexpected<JSONParseStatus>{JSONParseStatus::Error};
    }

    typename SingleItemResponse<T>::data_t::value_type map_data{resource, std::move(*parse_result)};
    toRet.data = typename SingleItemResponse<T>::data_t{std::move(map_data)};
    return toRet;
}

std::string format_error(const ErrorResponse& r);

} // namespace mifs::jsend

#endif
