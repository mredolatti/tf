#ifndef MIFS_SERVERS_HPP
#define MIFS_SERVERS_HPP

#include <string>

#include "expected.hpp"

namespace mifs::models
{

class FileServer
{
  public:
    using parse_result_t = util::Expected<FileServer, int /* TODO */>;

    FileServer() = delete;
    FileServer(const FileServer&) = default;
    FileServer(FileServer&&) = default;
    FileServer& operator=(const FileServer&) = default;
    FileServer& operator=(FileServer&&) = default;
    ~FileServer() = default;

    FileServer(std::string_view id, std::string_view org_id, std::string_view name,
               std::string_view fetch_url);

    template <typename Serialized> static parse_result_t parse(const Serialized& data);

    const std::string& id() const;
    const std::string& org_name() const;
    const std::string& name() const;
    const std::string& fetch_url() const;

  private:
    std::string id_;
    std::string org_name_;
    std::string name_;
    std::string fetch_url_;
};

} // namespace mifs::models

#endif
