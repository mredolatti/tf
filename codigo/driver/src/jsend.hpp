#ifndef MIFS_JSEND_HPP
#define MIFS_JSEND_HPP

#include "expected.hpp"

#include <iostream>
#include <string>
#include <unordered_map>
#include <vector>

#include <rapidjson/document.h>


namespace mifs::jsend {

enum class Status {
    SUCCESS,
    FAILURE,
    ERROR
};

namespace detail {

Status parse_status(std::string_view status);

} // namespace detail

template<typename T>
struct Response
{
    using data_t = std::unordered_map<std::string, std::vector<T>>;

    Status status;
    std::string code;
    std::string message;
    data_t data;

};

template<typename T>
util::Expected<Response<T>, int> parse(std::string_view body, std::string_view resource)
{

    rapidjson::Document doc;
    doc.Parse(body.data());

    if (!doc.HasMember("status") || !doc["status"].IsString()) {
        return util::Unexpected<int>(-1);
    }

    Response<T> toRet{.status = detail::parse_status(doc["status"].GetString())};

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
        return util::Unexpected<int>(-1);
    }

    const auto& items{data[resource.data()].GetArray()};
    std::vector<T> parsed_items;
    parsed_items.reserve(items.Size());
    for (auto& item : items) {
        auto parse_result{T::parse(item)};
        if (!parse_result) {
            continue;
        }
        parsed_items.emplace_back(*parse_result);
    }

    std::cout << "parseados " << parsed_items.size() << " items\n";

    typename Response<T>::data_t::value_type map_data{resource, std::move(parsed_items)};
    toRet.data = typename Response<T>::data_t{std::move(map_data)};
    return toRet;
    //return util::Expected<Response<T>, int>{std::move(toRet)};
}

} // namespace mifs::jsend

#endif
