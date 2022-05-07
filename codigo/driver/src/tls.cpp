#include "tls.hpp"

namespace mifs::tls {

const std::string& Config::root_ca_fn() const
{
    return root_ca_fn_;
}

const std::string& Config::client_cert_fn() const
{
    return client_cert_fn_;
}

const std::string& Config::client_pk_fn() const
{
    return client_pk_fn_;
}

}
