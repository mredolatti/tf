#ifndef MIFS_FUSE_OPTS_HPP
#define MIFS_FUSE_OPTS_HPP

namespace mifs::fuseopts
{

struct options {
    const char *config;
};

options parse(int *argc, char ***argv);

} // namespace mifs::fuseopts

#endif
