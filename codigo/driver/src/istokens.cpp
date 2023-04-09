#include "istokens.hpp"

namespace mifs::apiclients
{

EnvISTokenSource::EnvISTokenSource(std::string var_name)
    : var_name_{std::move(var_name)}
{
}

IndexServerTokenSource::token_result_t EnvISTokenSource::get()
{
    if (auto token{std::getenv(var_name_.c_str())}) {
        return std::string{token};
    }
    return util::Unexpected<int>{1};
}

} // namespace mifs::apiclients
