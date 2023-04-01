#include "fuseopts.hpp"
#include <cstdio>
#include <cstdlib>
#include <cstddef>
#include <cstring>

#define FUSE_USE_VERSION 35
#include <fuse3/fuse.h>
#include <iostream>

namespace mifs::fuseopts {


enum {
     KEY_HELP,
     KEY_VERSION,
};

static struct fuse_opt mifs_opts[] = {
     {"config=%s", offsetof(options, config), 0},
     FUSE_OPT_KEY("-V",        KEY_VERSION),
     FUSE_OPT_KEY("--version", KEY_VERSION),
     FUSE_OPT_KEY("-h",        KEY_HELP),
     FUSE_OPT_KEY("--help",    KEY_HELP),
     FUSE_OPT_END
};

static int mifs_opt_proc(void *data, const char *arg, int key, struct fuse_args *outargs)
{
     switch (key) {
     case KEY_HELP:
             fprintf(stderr,
                     "usage: %s mountpoint [options]\n"
                     "\n"
                     "general options:\n"
                     "    -o opt,[opt...]  mount options\n"
                     "    -h   --help      print help\n"
                     "    -V   --version   print version\n"
                     "\n"
                     "MIFS options:\n"
                     "    -o config=<path_to_config_file>\n"
                     , outargs->argv[0]);
             fuse_opt_add_arg(outargs, "-h");
             exit(1);

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

}
