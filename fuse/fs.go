package fuse

import (
	"time"

	"github.com/golang/glog"
	gofuse "github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

type GitFS struct {}

func (f *GitFS) String() string {
	return "GitFS"
}

func (f *GitFS) SetDebug(debug bool) {
	// TODO: implement
}

func (f *GitFS) GetAttr(name string, context *gofuse.Context) (*gofuse.Attr, gofuse.Status) {
	// TODO: implement
	glog.V(1).Infof("GetAttr(name=%q) called", name)
	return nil, gofuse.ENOENT
}

func (f *GitFS) Chmod(name string, mode uint32, context *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Chown(name string, uid uint32, gid uint32, context *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Utimens(name string, atime *time.Time, mtime *time.Time, context *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Truncate(name string, size uint64, context *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Access(name string, mode uint32, context *gofuse.Context) gofuse.Status {
	// TODO: Implement
	glog.V(1).Infof("Access(name=%q, mode=%#o) called", name, mode)
	return gofuse.ENOENT
}

func (f *GitFS) Link(oldName string, newName string, context *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Mkdir(name string, mode uint32, context *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Mknod(name string, mode uint32, dev uint32, context *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Rename(oldName string, newName string, context *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Rmdir(name string, context *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Unlink(name string, context *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) GetXAttr(name string, attribute string, context *gofuse.Context) ([]byte, gofuse.Status) {
	// TODO: implement
	glog.V(1).Infof("GetXAttr(name=%q, attribute=%q) called", name, attribute)
	return nil, gofuse.ENOENT
}

func (f *GitFS) ListXAttr(name string, context *gofuse.Context) ([]string, gofuse.Status) {
	// TODO: implement
	glog.V(1).Infof("ListXAttr(name=%q) called", name)
	return nil, gofuse.ENOENT
}

func (f *GitFS) RemoveXAttr(name string, attr string, context *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) SetXAttr(name string, attr string, data []byte, flags int, context *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) OnMount(nodeFs *pathfs.PathNodeFs) {
	glog.V(1).Infof("OnMount() called")
}

func (f *GitFS) OnUnmount() {
	glog.V(1).Infof("OnUnmount() called")
}

func (f *GitFS) Open(name string, flags uint32, context *gofuse.Context) (nodefs.File, gofuse.Status) {
	// TODO: implement
	glog.V(1).Infof("Open(name=%q, flags=%#x) called", name, flags)
	return nil, gofuse.ENOENT
}

func (f *GitFS) Create(name string, flags uint32, mode uint32, context *gofuse.Context) (nodefs.File, gofuse.Status) {
	return nil, gofuse.EROFS
}

func (f *GitFS) OpenDir(name string, context *gofuse.Context) ([]gofuse.DirEntry, gofuse.Status) {
	// TODO: implement
	glog.V(1).Infof("OpenDir(name=%q) called", name)
	return nil, gofuse.ENOENT
}

func (f *GitFS) Symlink(value string, linkName string, context *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Readlink(name string, context *gofuse.Context) (string, gofuse.Status) {
	// TODO: implement
	glog.V(1).Infof("Readlink(name=%q) called", name)
	return "", gofuse.ENOENT
}

func (f *GitFS) StatFs(name string) *gofuse.StatfsOut {
	// TODO: implement
	glog.V(1).Infof("StatFs(name=%q) called", name)
	return &gofuse.StatfsOut{}
}