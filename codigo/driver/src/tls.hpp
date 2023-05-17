#ifndef MIFS_TLS_HPP
#define MIFS_TLS_HPP

#include <string>

namespace mifs::tls
{

class Config
{
  public:
    const std::string& root_ca_fn() const;
    const std::string& client_cert_fn() const;
    const std::string& client_pk_fn() const;

    Config(std::string_view root_ca, std::string_view client_cert, std::string_view client_pk);

  private:
    std::string root_ca_fn_;
    std::string client_cert_fn_;
    std::string client_pk_fn_;
};

} // namespace mifs::tls

#endif
