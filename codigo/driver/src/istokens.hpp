#ifndef MIFS_IS_TOKENS_HPP
#define MIFS_IS_TOKENS_HPP

#include "expected.hpp"
#include <string>

namespace mifs::apiclients
{

class IndexServerTokenSource
{
  public:
    using token_result_t = util::Expected<std::string, int>;
    virtual token_result_t get() = 0;
};

class EnvISTokenSource : public IndexServerTokenSource
{
  public:
    explicit EnvISTokenSource(std::string var_name);
    token_result_t get();

  private:
    std::string var_name_;
};

} // namespace mifs::apiclients

#endif
