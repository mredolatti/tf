#ifndef MIFS_OPEN_FILE_TRACKER_HPP
#define MIFS_OPEN_FILE_TRACKER_HPP

#include <functional>
#include <mutex>
#include <optional>
#include <string>
#include <unordered_map>
namespace mifs::util {

struct OpenFile
{
    int fd;
    std::string name;
    std::size_t offset;
    int mode;
};

class OpenFileTracker
{
    public:
    OpenFileTracker();
    OpenFileTracker(const OpenFileTracker&) = delete;
    OpenFileTracker(OpenFileTracker&&) noexcept = delete;
    OpenFileTracker& operator=(const OpenFileTracker&) = delete;
    OpenFileTracker& operator=(OpenFileTracker&&) noexcept = delete;
    ~OpenFileTracker() noexcept = default;

    int open(std::string_view path, int mode);
    std::optional<std::reference_wrapper<OpenFile>> get(int fd);
    void close(int fd);

    private:
    std::unordered_map<int, OpenFile> open_files_;
    std::mutex mutex_;
    int last_fd_;
};


} // namespace mifs::util
#endif // MIFS_OPEN_FILE_TRACKER_HPP
