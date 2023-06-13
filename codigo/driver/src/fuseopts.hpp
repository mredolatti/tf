#ifndef MIFS_FUSE_OPTS_HPP
#define MIFS_FUSE_OPTS_HPP

#include <filesystem>

namespace mifs::fuseopts
{

struct options {
    const char *config = nullptr;
    std::filesystem::path mount_point;
};

options parse(int *argc, char ***argv);

} // namespace mifs::fuseopts

#endif
