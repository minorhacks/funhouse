package fuse

import (
	"time"

	"github.com/golang/glog"
	gofuse "github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

type GitFile struct{}

func (f *GitFile) SetInode(inode *nodefs.Inode) {
	glog.V(1).Infof("SetInode() called")
}

func (f *GitFile) String() string {
	return "GitFile"
}

func (f *GitFile) InnerFile() nodefs.File {
	return nil
}

func (f *GitFile) Read(dest []byte, offset int64) (res gofuse.ReadResult, status gofuse.Status) {
	glog.V(1).Infof("Read(dest=len(%d), offset=%#x)", len(dest), offset)
	defer func() {
		if status != gofuse.OK {
			glog.Error("Read(dest=len(%d), offset=%#x) error: %v", len(dest), offset, status)
		}
	}()

	// TODO: implement
	return gofuse.ReadResultData(nil), gofuse.EIO
}

func (f *GitFile) Write(data []byte, offset int64) (written uint32, status gofuse.Status) {
	glog.V(1).Infof("Write(data=len(%d), offset=%#x)", len(data), offset)
	defer func() {
		if status != gofuse.OK {
			glog.Error("Write(data=len(%d), offset=%#x) error: %v", len(data), offset, status)
		}
	}()

	// TODO: implement
	return 0, gofuse.ENOSYS
}

func (f *GitFile) GetLk(owner uint64, lk *gofuse.FileLock, flags uint32, out *gofuse.FileLock) (status gofuse.Status) {
	glog.V(1).Infof("GetLk(owner=%d, flags=%#x) called", owner, flags)
	defer func() {
		if status != gofuse.OK {
			glog.Error("GetLk(owner=%d, flags=%#x) error: %v", owner, flags, status)
		}
	}()

	// TODO: implement
	return gofuse.ENOSYS
}

func (f *GitFile) SetLk(owner uint64, lk *gofuse.FileLock, flags uint32) (status gofuse.Status) {
	glog.V(1).Infof("SetLk(owner=%d, flags=%#x) called", owner, flags)
	defer func() {
		if status != gofuse.OK {
			glog.Error("SetLk(owner=%d, flags=%#x) error: %v", owner, flags, status)
		}
	}()

	// TODO: implement
	return gofuse.ENOSYS
}

func (f *GitFile) SetLkw(owner uint64, lk *gofuse.FileLock, flags uint32) (status gofuse.Status) {
	glog.V(1).Infof("SetLkw(owner=%d, flags=%#x) called", owner, flags)
	defer func() {
		if status != gofuse.OK {
			glog.Error("SetLkw(owner=%d, flags=%#x) error: %v", owner, flags, status)
		}
	}()

	// TODO: implement
	return gofuse.ENOSYS
}

func (f *GitFile) Flush() (status gofuse.Status) {
	glog.V(1).Infof("Flush() called")
	defer func() {
		if status != gofuse.OK {
			glog.Error("Flush() error: %v", status)
		}
	}()

	// TODO: implement
	return gofuse.ENOSYS
}

func (f *GitFile) Release() {
	glog.V(1).Infof("Release() called")
}

func (f *GitFile) Fsync(flags int) (status gofuse.Status) {
	glog.V(1).Infof("Fsync(flags=%#x) called", flags)
	defer func() {
		if status != gofuse.OK {
			glog.Error("Fsync(flags=%#x) error: %v", flags, status)
		}
	}()

	// TODO: implement
	return gofuse.ENOSYS
}

func (f *GitFile) Truncate(size uint64) (status gofuse.Status) {
	glog.V(1).Infof("Truncate(size=%d) called", size)
	defer func() {
		if status != gofuse.OK {
			glog.Error("Truncate(size=%d) error: %v", size, status)
		}
	}()

	// TODO: implement
	return gofuse.ENOSYS
}

func (f *GitFile) GetAttr(out *gofuse.Attr) (status gofuse.Status) {
	glog.V(1).Infof("GetAttr() called")
	defer func() {
		if status != gofuse.OK {
			glog.Error("GetAttr() error: %v", status)
		}
	}()

	// TODO: implement
	return gofuse.ENOSYS
}

func (f *GitFile) Chown(uid uint32, gid uint32) (status gofuse.Status) {
	glog.V(1).Infof("Chown(uid=%d, gid=%d) called", uid, gid)
	defer func() {
		if status != gofuse.OK {
			glog.Error("Chown(uid=%d, gid=%d) error: %v", uid, gid, status)
		}
	}()

	// TODO: implement
	return gofuse.ENOSYS
}

func (f *GitFile) Chmod(perms uint32) (status gofuse.Status) {
	glog.V(1).Infof("Chmod(perms=%#o) called", perms)
	defer func() {
		if status != gofuse.OK {
			glog.Error("Chmod(perms=%#o) error: %v", perms, status)
		}
	}()

	// TODO: implement
	return gofuse.ENOSYS
}

func (f *GitFile) Utimens(atime *time.Time, mtime *time.Time) (status gofuse.Status) {
	glog.V(1).Infof("Utimens(atime=%q, mtime=%q) called", atime, mtime)
	defer func() {
		if status != gofuse.OK {
			glog.Error("Utimens(atime=%q, mtime=%q) error: %v", atime, mtime, status)
		}
	}()

	// TODO: implement
	return gofuse.ENOSYS
}

func (f *GitFile) Allocate(offset uint64, size uint64, mode uint32) (status gofuse.Status) {
	glog.V(1).Infof("Allocate(offset=%#x, size=%#x, mode=%#o) called", offset, size, mode)
	defer func() {
		if status != gofuse.OK {
			glog.Error("Allocate(offset=%#x, size=%#x, mode=%#o) error: %v", offset, size, mode, status)
		}
	}()

	// TODO: implement
	return gofuse.ENOSYS
}
