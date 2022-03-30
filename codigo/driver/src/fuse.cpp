#define FUSE_USE_VERSION 35

#include <fuse3/fuse.h>
#include <string.h>
#include <unistd.h>
#include <dirent.h>
#include <errno.h>
#include <iostream>


static void* mifs_init(struct fuse_conn_info *conn, struct fuse_config *cfg)
{
    (void) conn;
    return nullptr;
}

static int mifs_getattr(const char *path, struct stat *stbuf,
               struct fuse_file_info *fi)
{
    (void) fi;
    auto res{lstat(path, stbuf)};
    if (res == -1) {
        return -errno;
    }

    return 0;
}

static int mifs_access(const char *path, int mask)
{
    auto res{access(path, mask)};
    if (res == -1) {
        return -errno;
    }

    return 0;
}

static int mifs_readlink(const char *path, char *buf, size_t size)
{
    auto res{readlink(path, buf, size - 1)};
    if (res == -1) {
        return -errno;
    }

    buf[res] = '\0';
    return 0;
}


static int mifs_readdir(
    const char *path, 
    void *buf, 
    fuse_fill_dir_t filler, 
    off_t offset, 
    struct fuse_file_info *fi,
    enum fuse_readdir_flags flags
) {

    (void) offset;
    (void) fi;
    (void) flags;

    auto dp{opendir(path)};
    if (dp == nullptr) {
        return -errno;
    }

    struct dirent *de;
    while ((de = readdir(dp)) != nullptr) {
        struct stat st;
        memset(&st, 0, sizeof(st));
        st.st_ino = de->d_ino;
        st.st_mode = de->d_type << 12;
        if (filler(buf, de->d_name, &st, 0, static_cast<enum fuse_fill_dir_flags>(0))) {
            break;
        }
    }

    closedir(dp);
    return 0;
}

static int mifs_mkdir(const char *path, mode_t mode)
{
    auto res{mkdir(path, mode)};
    if (res == -1) {
        return -errno;
    }

    return 0;
}

static int mifs_unlink(const char *path)
{
    auto res{unlink(path)};
    if (res == -1) {
        return -errno;
    }

    return 0;
}

static int mifs_rmdir(const char *path)
{
    auto res{rmdir(path)};
    if (res == -1) {
        return -errno;
    }

    return 0;
}

static int mifs_symlink(const char *from, const char *to)
{
    auto res{symlink(from, to)};
    if (res == -1) {
        return -errno;
    }

    return 0;
}

static int mifs_rename(const char *from, const char *to, unsigned int flags)
{
    if (flags) {
        return -EINVAL;
    }

    auto res{rename(from, to)};
    if (res == -1) {
        return -errno;
    }

    return 0;
}

static int mifs_link(const char *from, const char *to)
{
    auto res{link(from, to)};
    if (res == -1) {
        return -errno;
    }

    return 0;
}

static int mifs_chmod(const char *path, mode_t mode, struct fuse_file_info *fi)
{
    (void) fi;

    auto res{chmod(path, mode)};
    if (res == -1) {
        return -errno;
    }

    return 0;
}

static int mifs_chown(const char *path, uid_t uid, gid_t gid, struct fuse_file_info *fi)
{
    (void) fi;

    auto res{lchown(path, uid, gid)};
    if (res == -1) {
        return -errno;
    }

    return 0;
}

static int mifs_truncate(const char *path, off_t size,
            struct fuse_file_info *fi)
{

    auto res{(fi != nullptr) ? ftruncate(fi->fh, size) : truncate(path, size)};
    if (res == -1) {
        return -errno;
    }

    return 0;
}

static int mifs_create(const char *path, mode_t mode,
              struct fuse_file_info *fi)
{
    auto res{open(path, fi->flags, mode)};
    if (res == -1) {
        return -errno;
    }

    fi->fh = res;
    return 0;
}

static int mifs_open(const char *path, struct fuse_file_info *fi)
{
    auto res{open(path, fi->flags)};
    if (res == -1) {
        return -errno;
    }

    fi->fh = res;
    return 0;
}

static int mifs_read(const char *path, char *buf, size_t size, off_t offset, struct fuse_file_info *fi)
{
    auto fd{(fi == nullptr) ? open(path, O_RDONLY) : fi->fh};
    if (fd == -1) {
        return -errno;
    }

    auto res{pread(fd, buf, size, offset)};
    if (res == -1) {
        res = -errno;
    }

    if(fi == nullptr) {
        close(fd);
    }

    return res;
}

static int mifs_write(const char *path, const char *buf, size_t size, off_t offset, struct fuse_file_info *fi)
{
    auto fd{(fi == nullptr) ? open(path, O_WRONLY) : fi->fh};
    if (fd == -1) {
        return -errno;
    }

    auto res{pwrite(fd, buf, size, offset)};
    if (res == -1) {
        res = -errno;
    }

    if(fi == nullptr) {
        close(fd);
    }

    return res;
}

static int mifs_statfs(const char *path, struct statvfs *stbuf)
{
    auto res{statvfs(path, stbuf)};
    if (res == -1) {
        return -errno;
    }

    return 0;
}

static int mifs_release(const char *path, struct fuse_file_info *fi)
{
    (void) path;
    close(fi->fh);
    return 0;
}

static int mifs_fsync(const char *path, int isdatasync,
             struct fuse_file_info *fi)
{
    (void) path;
    (void) isdatasync;
    (void) fi;
    return 0;
}

static int mifs_opendir(const char *dir, struct fuse_file_info *info) {
    (void)info;
    return 0;
}

static const struct fuse_operations mifs_oper = {
    .getattr    = mifs_getattr,
    .mkdir      = mifs_mkdir,
    .unlink     = mifs_unlink,
    .rmdir      = mifs_rmdir,
    .rename     = mifs_rename,
    .link       = mifs_link,
    .chmod      = mifs_chmod,
    .chown      = mifs_chown,
    .truncate   = mifs_truncate,
    .open       = mifs_open,
    .read       = mifs_read,
    .write      = mifs_write,
    .statfs     = mifs_statfs,
    .release    = mifs_release,
    .fsync      = mifs_fsync,
    .opendir    = mifs_opendir,
    .readdir    = mifs_readdir,
    .init       = mifs_init,
    .access     = mifs_access,
    .create     = mifs_create,
};

int init_fuse(int argc, char *argv[])
{
    umask(0);
    return fuse_main(argc, argv, &mifs_oper, NULL);
}
