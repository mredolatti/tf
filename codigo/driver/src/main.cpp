#include <ios>
#include <iostream>
#include <memory>
#include "filemanager.hpp"
#include "fsclient.hpp"
#include "fuse.hpp"
#include "http.hpp"
#include "httpc.hpp"
#include "isclient.hpp"
#include "log.hpp"
#include "mappings.hpp"

#include "fstree.hpp"

using namespace mifs;


int main(int argc, char**argv)
{
    (void)argc;
    (void)argv;

    //FS_CACERT=${FS_CACERT:-'PKI/root/certs/ca.crt'}
    //FS_CERT=${FS_CERT:-'PKI/client/certs/client.crt'}
    //FS_KEY=${FS_KEY:-'PKI/client/private/client.key'}

    auto logger{log::initialize()};
    assert(logger);

    SPDLOG_LOGGER_INFO(logger, "starting...");

    auto client{std::make_shared<http::Client>()};
    mifs::apiclients::IndexServerClient isc{client, mifs::apiclients::IndexServerClient::Config{"http://index-server:9876"}};
    mifs::apiclients::FileServerClient fsc{client, mifs::apiclients::FileServerClient::server_infos_t{
        {
            "fs1",
            mifs::apiclients::detail::ServerInfo{"fs1", "https://file-server:9877", tls::Config{
                "/home/martin/Projects/tf/codigo/PKI/root/certs/ca.crt",
                "/home/martin/Projects/tf/codigo/PKI/client/certs/client.crt",
                "/home/martin/Projects/tf/codigo/PKI/client/private/client.key"
            }}
        }
    }};

    /*
    auto res{fsc.get_all("fs1")};

    if (res) {
        for (const auto& file: (*res).data["files"]) {
            std::cout << "- " << file << std::endl;
        }
    }
*/
    mifs::FileManager fm{std::move(isc), std::move(fsc)};
    fm.sync();

    ContextData ctx{argv[argc-1], logger, fm};
    init_fuse(argc, argv, ctx);

    return 0;
}
