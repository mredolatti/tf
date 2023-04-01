#include <ios>
#include <iostream>
#include <memory>
#include "filemanager.hpp"
#include "fsclient.hpp"
#include "fuse.hpp"
#include "fuseopts.hpp"
#include "http.hpp"
#include "httpc.hpp"
#include "isclient.hpp"
#include "log.hpp"
#include "mappings.hpp"
#include "config.hpp"
#include "fscatalog.hpp"
#include "fstree.hpp"

using namespace mifs;


int main(int argc, char**argv)
{
    auto logger{log::initialize()};
    assert(logger);

    auto options{mifs::fuseopts::parse(&argc, &argv)};
    // TODO(mredolatti): validar options

    auto config{mifs::Config::parse(options.config)};
    assert(config);
    // TODO(mredolatti): validar config

    SPDLOG_LOGGER_INFO(logger, "starting...");

    auto client{std::make_shared<http::Client>()};
    auto fs_catalog{mifs::util::FileServerCatalog::createFromCredentialsConfig((*config).creds())};

    mifs::FileManager fm{
        mifs::apiclients::IndexServerClient{client, mifs::apiclients::IndexServerClient::Config::from_parsed_conf(*config)},
        mifs::apiclients::FileServerClient{client, fs_catalog},
        fs_catalog
    };
    fm.sync();

    ContextData ctx{argv[argc-1], logger, fm};
    init_fuse(argc, argv, ctx);
    return 0;
}
