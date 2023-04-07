#include "openfiles.hpp"
#include <mutex>
#include <optional>


namespace mifs::util {

OpenFileTracker::OpenFileTracker() :
    last_fd_{1024}
{}


int OpenFileTracker::open(std::string_view path, int mode)
{
    std::lock_guard<std::mutex> lk{mutex_};
    auto curr = last_fd_++;
    open_files_[curr] = OpenFile{.fd = curr, .name = std::string{path}, .offset = 0, .mode=mode};
    return curr;
}

std::optional<std::reference_wrapper<OpenFile>> OpenFileTracker::get(int fd)
{
    std::lock_guard<std::mutex> lk{mutex_};
    auto it{open_files_.find(fd)};
    if (it == open_files_.end()) {
        return std::nullopt;
    }
    return it->second;
}

void OpenFileTracker::close(int fd)
{
    std::lock_guard<std::mutex> lk{mutex_};
    open_files_.erase(fd);
}


} // namespace mifs::util
