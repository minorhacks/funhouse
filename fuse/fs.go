package fuse

import (
	"syscall"
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

func (f *GitFS) GetAttr(name string, context *gofuse.Context) (ret *gofuse.Attr, status gofuse.Status) {
	glog.V(1).Infof("GetAttr(name=%q) called", name)
	defer func() {
		if status != gofuse.OK {
			glog.Errorf("GetAttr(name=%q) failed: %v", name, status)
		}
	}()

	// Simulate top-level dirs
	switch name {
	case "":
		return &gofuse.Attr{
			Mode: syscall.S_IFDIR,
		}, gofuse.OK
	case "commits":
		return &gofuse.Attr{
			Mode: syscall.S_IFDIR,
		}, gofuse.OK
	}
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

func (f *GitFS) Access(name string, mode uint32, context *gofuse.Context) (status gofuse.Status) {
	// TODO: Implement
	glog.V(1).Infof("Access(name=%q, mode=%#o) called", name, mode)
	defer func() {
		if status != gofuse.OK {
			glog.Errorf("Access(name=%q) failed: %v", name, status)
		}
	}()
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

func (f *GitFS) GetXAttr(name string, attribute string, context *gofuse.Context) (xattr []byte, status gofuse.Status) {
	glog.V(1).Infof("GetXAttr(name=%q, attribute=%q) called", name, attribute)
	defer func() {
		if status != gofuse.OK {
			glog.Errorf("GetXAttr(name=%q) failed: %v", name, status)
		}
	}()

	// Assume there are never any extended attributes for now.
	return nil, gofuse.OK
}

func (f *GitFS) ListXAttr(name string, context *gofuse.Context) (xattrs []string, status gofuse.Status) {
	// TODO: implement
	glog.V(1).Infof("ListXAttr(name=%q) called", name)
	defer func() {
		if status != gofuse.OK {
			glog.Errorf("ListXAttr(name=%q) failed: %v", name, status)
		}
	}()

	// Assume there are never any extended attributes for now.
	return nil, gofuse.OK
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

func (f *GitFS) Open(name string, flags uint32, context *gofuse.Context) (file nodefs.File, status gofuse.Status) {
	// TODO: implement
	glog.V(1).Infof("Open(name=%q, flags=%#x) called", name, flags)
	defer func() {
		if status != gofuse.OK {
			glog.Errorf("Open(name=%q) failed: %v", name, status)
		}
	}()
	return nil, gofuse.ENOENT
}

func (f *GitFS) Create(name string, flags uint32, mode uint32, context *gofuse.Context) (nodefs.File, gofuse.Status) {
	return nil, gofuse.EROFS
}

func (f *GitFS) OpenDir(name string, context *gofuse.Context) (dirs []gofuse.DirEntry, status gofuse.Status) {
	glog.V(1).Infof("OpenDir(name=%q) called", name)
	defer func() {
		if status != gofuse.OK {
			glog.Errorf("OpenDir(name=%q) failed: %v", name, status)
		}
	}()

	// Simulate top-level dirs
	switch name {
	case "":
		return []gofuse.DirEntry{
			{
				Name: "commits",
				Mode: syscall.S_IFDIR,
			},
		}, gofuse.OK
	}

	return nil, gofuse.ENOENT
}

func (f *GitFS) Symlink(value string, linkName string, context *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Readlink(name string, context *gofuse.Context) (link string, status gofuse.Status) {
	// TODO: implement
	glog.V(1).Infof("Readlink(name=%q) called", name)
	defer func() {
		if status != gofuse.OK {
			glog.Errorf("Readlink(name=%q) failed: %v", name, status)
		}
	}()
	return "", gofuse.ENOENT
}

func (f *GitFS) StatFs(name string) *gofuse.StatfsOut {
	// TODO: implement
	glog.V(1).Infof("StatFs(name=%q) called", name)
	return &gofuse.StatfsOut{}
}