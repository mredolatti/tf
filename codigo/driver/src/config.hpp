#ifndef MIFS_OPTIONS_HPP
#define MIFS_OPTIONS_HPP

#include "expected.hpp"
#include <string>
#include <unordered_map>

namespace mifs {

class Config
{
    public:

    enum class ParseError
    {
        Ok = 0,
        ErrorOpeningFile = 1,
        ErrorParsingJSON = 2
    };

    struct credentials
    {
        std::string root_certificate_fn;
        std::string client_certificate_fn;
        std::string client_private_key_fn;
    };

    struct is_cfg
    {
        std::string url;
        std::string root_cert;
        std::string token_source_;
    };


    using result_t = util::Expected<Config, ParseError>;
    using fs_creds_by_server_t = std::unordered_map<std::string, credentials>;
    using fs_creds_by_org_t = std::unordered_map<std::string, fs_creds_by_server_t>;

    Config() = delete;
    Config(const Config&) = default;
    Config(Config&&) noexcept = default;
    Config& operator=(const Config&) = default;
    Config& operator=(Config&&) = default;
    ~Config() = default;

    static result_t parse(const std::string& config_file_fn);
    const is_cfg& index_server() const;
    const fs_creds_by_org_t& creds() const;

    private:
    Config(is_cfg is_conf, fs_creds_by_org_t);
    is_cfg index_server_config_;    
    fs_creds_by_org_t creds_;
};

}


#endif
