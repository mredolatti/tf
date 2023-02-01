#ifndef IMFS_FUSE_HPP
#define IMFS_FUSE_HPP

#include "filemanager.hpp"

class ContextData
{
    public:
    ContextData() = delete;
    ContextData(ContextData&) noexcept = default;
    ContextData& operator=(const ContextData&) = delete;
    ContextData& operator=(ContextData&&) noexcept = delete;
    ~ContextData() = default;

    ContextData(std::string mount_point, mifs::log::logger_t logger, mifs::FileManager& fm);
    mifs::log::logger_t& logger();
    mifs::FileManager& file_manager();
    const std::string& mount_point() const;

    private:
    std::string mount_point_;
    mifs::log::logger_t logger_;
    mifs::FileManager& fm_;
};

int init_fuse(int argc, char **argv, ContextData& ctx);

#endif
