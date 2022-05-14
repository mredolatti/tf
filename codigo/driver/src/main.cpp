#include <ios>
#include <iostream>
#include <memory>
#include "filemanager.hpp"
#include "fuse.hpp"
#include "http.hpp"
#include "httpc.hpp"
#include "isclient.hpp"
#include "log.hpp"
#include "mappings.hpp"

using namespace mifs;

int main(int argc, char**argv)
{
    (void)argc;
    (void)argv;

    /*
    FS_CACERT=${FS_CACERT:-'PKI/root/certs/ca.crt'}
FS_CERT=${FS_CERT:-'PKI/client/certs/client.crt'}
FS_KEY=${FS_KEY:-'PKI/client/private/client.key'}

    */

    auto logger{log::initialize()};
    assert(logger);

    SPDLOG_LOGGER_INFO(logger, "starting...");
    auto client{std::make_shared<http::Client>()};
    mifs::FileManager fm{client};
    fm.sync();

    ContextData ctx{logger, fm};
    init_fuse(argc, argv, ctx);

    return 0;
}
