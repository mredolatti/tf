#include "tls.hpp"

namespace mifs::tls
{

Config::Config(std::string root_ca, std::string client_cert, std::string client_pk)
    : root_ca_fn_{std::move(root_ca)},
      client_cert_fn_{std::move(client_cert)},
      client_pk_fn_{std::move(client_pk)}
{
}

const std::string& Config::root_ca_fn() const { return root_ca_fn_; }

const std::string& Config::client_cert_fn() const { return client_cert_fn_; }

const std::string& Config::client_pk_fn() const { return client_pk_fn_; }

} // namespace mifs::tls
