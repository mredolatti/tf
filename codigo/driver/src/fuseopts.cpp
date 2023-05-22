#include "fuseopts.hpp"
#include <cstddef>
#include <cstdio>
#include <cstdlib>
#include <cstring>

#define FUSE_USE_VERSION 35
#include <fuse3/fuse.h>
#include <iostream>

namespace mifs::fuseopts
{

enum {
    KEY_HELP,
    KEY_VERSION,
};

static struct fuse_opt mifs_opts[] = {FUSE_OPT_KEY("-V", KEY_VERSION), FUSE_OPT_KEY("--version", KEY_VERSION),
                                      FUSE_OPT_KEY("-h", KEY_HELP), FUSE_OPT_KEY("--help", KEY_HELP),
                                      FUSE_OPT_END};

static int mifs_opt_proc(void *data, const char *arg, int key, struct fuse_args *outargs)
{
    auto *opts{reinterpret_cast<struct options *>(data)};
    if (key == FUSE_OPT_KEY_NONOPT && opts->config == nullptr) {
        // this is the first argument/device to mount, in this case the json cfg file
        opts->config = arg;
        return 0;
    }

    switch (key) {
    case KEY_HELP:
        fprintf(stderr,
                "usage: %s mountpoint [options]\n"
                "\n"
                "general options:\n"
                "    -o opt,[opt...]  mount options\n"
                "    -h   --help      print help\n"
                "    -V   --version   print version\n"
                "\n",
                outargs->argv[0]);
        fuse_opt_add_arg(outargs, "-h");
        exit(0);

    case KEY_VERSION:
        fprintf(stderr, "MIFS version %s\n", "0.1");
        fuse_opt_add_arg(outargs, "--version");
        exit(0);
    }
    return 1;
}

options parse(int *argc, char ***argv)
{
    struct options opts;
    memset(&opts, 0, sizeof(options));
    struct fuse_args custom_args = FUSE_ARGS_INIT(*argc, *argv);
    fuse_opt_parse(&custom_args, &opts, mifs_opts, mifs_opt_proc);

    *argc = custom_args.argc;
    *argv = custom_args.argv;

    return opts;
}

} // namespace mifs::fuseopts
