package fuse

import (
	"context"
	"os"
	"regexp"
	"strings"
	"syscall"
	"time"

	fspb "github.com/minorhacks/funhouse/proto/git_read_fs_proto"

	"github.com/golang/glog"
	gofuse "github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"google.golang.org/grpc/codes"
	grpcstat "google.golang.org/grpc/status"
)

var commitHashPattern = regexp.MustCompile(`[0-9a-f]{40}`)

type GitFS struct {
	ServerAddr string
	Client     fspb.GitReadFsClient
}

func (f *GitFS) String() string {
	return "GitFS"
}

func (f *GitFS) SetDebug(debug bool) {
	// TODO: implement
}

func (f *GitFS) GetAttr(name string, ctx *gofuse.Context) (ret *gofuse.Attr, status gofuse.Status) {
	glog.V(1).Infof("GetAttr(name=%q) called", name)
	defer func() {
		if status != gofuse.OK {
			glog.Errorf("GetAttr(name=%q) failed: %v", name, status)
		}
	}()

	path := strings.FieldsFunc(name, func(c rune) bool { return c == '/' })
	// Simulate top-level dirs
	switch {
	case len(path) == 0:
		return &gofuse.Attr{
			Mode: syscall.S_IFDIR,
		}, gofuse.OK
	case len(path) == 1 && path[0] == "commits":
		return &gofuse.Attr{
			Mode: syscall.S_IFDIR,
		}, gofuse.OK
	case len(path) == 1 && path[0] == "branches":
		return &gofuse.Attr{
			Mode: syscall.S_IFDIR,
		}, gofuse.OK
	case len(path) == 2 && path[0] == "commits":
		// The right thing to do here is to query and see which paths are
		// present, to avoid optimistically returning directories where none
		// exist. This is important because GetAttr is called when programs are
		// checking for file existence of files that don't necessarily exist.
		//
		// For now, just make sure path[1] looks like a commit hash; it's
		// unlikely that false positives will be reported since that means that
		// programs are checking for the existence of fabricated hashes.
		if !commitHashPattern.MatchString(path[1]) {
			return nil, gofuse.ENOENT
		}
		return &gofuse.Attr{
			Mode: syscall.S_IFDIR,
		}, gofuse.OK
	case len(path) == 2 && path[0] == "branches":
		// Get the list of branches
		res, err := f.Client.ListBranches(context.TODO(), &fspb.ListBranchesRequest{})
		if err != nil {
			glog.Errorf("GetAttributes(Path=%q) returned error: %v", name, err)
			return nil, errnoFromCode(grpcstat.Convert(err))
		}
		// Return ENOENT if the branch name is not recognized
		if _, ok := res.Branches[path[1]]; !ok {
			glog.Errorf("GetAttributes(Path=%q) branch %q not found", name, path[1])
			return nil, gofuse.ENOENT
		}
		// Return a symlink to the branch's commit
		return &gofuse.Attr{
			Mode: syscall.S_IFLNK,
		}, gofuse.OK
	case len(path) >= 2 && path[0] == "commits":
		// Assume path[1] is the commit hash
		var filePath string
		if len(path) == 2 {
			filePath = "/"
		} else {
			filePath = "/" + strings.Join(path[2:], "/")
		}
		res, err := f.Client.GetAttributes(context.TODO(), &fspb.GetAttributesRequest{
			Commit: path[1],
			Path:   filePath,
		})
		if err != nil {
			glog.Errorf("GetAttributes(Commit=%q, Path=%q) returned error: %v", path[1], filePath, err)
			return nil, errnoFromCode(grpcstat.Convert(err))
		}
		return &gofuse.Attr{
			Mode: toSyscallMode(res.Mode),
			Size: res.SizeBytes,
		}, gofuse.OK
	}
	return nil, gofuse.ENOENT
}

func (f *GitFS) Chmod(name string, mode uint32, ctx *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Chown(name string, uid uint32, gid uint32, ctx *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Utimens(name string, atime *time.Time, mtime *time.Time, ctx *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Truncate(name string, size uint64, ctx *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Access(name string, mode uint32, ctx *gofuse.Context) (status gofuse.Status) {
	// TODO: Implement
	glog.V(1).Infof("Access(name=%q, mode=%#o) called", name, mode)
	defer func() {
		if status != gofuse.OK {
			glog.Errorf("Access(name=%q) failed: %v", name, status)
		}
	}()
	return gofuse.ENOSYS
}

func (f *GitFS) Link(oldName string, newName string, ctx *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Mkdir(name string, mode uint32, ctx *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Mknod(name string, mode uint32, dev uint32, ctx *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Rename(oldName string, newName string, ctx *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Rmdir(name string, ctx *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Unlink(name string, ctx *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) GetXAttr(name string, attribute string, ctx *gofuse.Context) (xattr []byte, status gofuse.Status) {
	glog.V(1).Infof("GetXAttr(name=%q, attribute=%q) called", name, attribute)
	defer func() {
		if status != gofuse.OK {
			glog.Errorf("GetXAttr(name=%q) failed: %v", name, status)
		}
	}()

	// Assume there are never any extended attributes for now.
	return nil, gofuse.OK
}

func (f *GitFS) ListXAttr(name string, ctx *gofuse.Context) (xattrs []string, status gofuse.Status) {
	glog.V(1).Infof("ListXAttr(name=%q) called", name)
	defer func() {
		if status != gofuse.OK {
			glog.Errorf("ListXAttr(name=%q) failed: %v", name, status)
		}
	}()

	// Assume there are never any extended attributes for now.
	return nil, gofuse.OK
}

func (f *GitFS) RemoveXAttr(name string, attr string, ctx *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) SetXAttr(name string, attr string, data []byte, flags int, ctx *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) OnMount(nodeFs *pathfs.PathNodeFs) {
	glog.V(1).Infof("OnMount() called")
}

func (f *GitFS) OnUnmount() {
	glog.V(1).Infof("OnUnmount() called")
}

func (f *GitFS) Open(name string, flags uint32, ctx *gofuse.Context) (file nodefs.File, status gofuse.Status) {
	glog.V(1).Infof("Open(name=%q, flags=%#x) called", name, flags)
	defer func() {
		if status != gofuse.OK {
			glog.Errorf("Open(name=%q, flags=%#x) failed: %v", name, flags, status)
		}
	}()

	// TODO: Do we need to check flags here?

	path := strings.FieldsFunc(name, func(c rune) bool { return c == '/' })
	// Expect the first elements to be ["commit", "<COMMIT HASH>"]
	if len(path) < 3 {
		return nil, gofuse.ENOENT
	}
	if path[0] != "commits" {
		return nil, gofuse.ENOENT
	}

	res, err := f.Client.GetFile(context.TODO(), &fspb.GetFileRequest{
		Commit: path[1],
		Path:   strings.Join(path[2:], "/"),
	})
	if err != nil {
		glog.Errorf("GetFile(Commit=%q, Path=%q) returned error: %v", path[1], strings.Join(path[2:], "/"), err)
		return nil, gofuse.EIO
	}
	return nodefs.NewDataFile(res.Contents), gofuse.OK
}

func (f *GitFS) Create(name string, flags uint32, mode uint32, ctx *gofuse.Context) (nodefs.File, gofuse.Status) {
	return nil, gofuse.EROFS
}

func (f *GitFS) OpenDir(name string, ctx *gofuse.Context) (dirs []gofuse.DirEntry, status gofuse.Status) {
	glog.V(1).Infof("OpenDir(name=%q) called", name)
	defer func() {
		if status != gofuse.OK {
			glog.Errorf("OpenDir(name=%q) failed: %v", name, status)
		}
	}()

	path := strings.FieldsFunc(name, func(c rune) bool { return c == os.PathSeparator })
	// Simulate top-level dirs
	switch {
	case len(path) == 0:
		return []gofuse.DirEntry{
			{
				Name: "commits",
				Mode: syscall.S_IFDIR,
			},
			{
				Name: "branches",
				Mode: syscall.S_IFDIR,
			},
		}, gofuse.OK
	case len(path) == 1 && path[0] == "commits":
		res, err := f.Client.ListCommits(context.TODO(), &fspb.ListCommitsRequest{})
		if err != nil {
			glog.Errorf("ListCommits() returned error: %v", err)
			return nil, gofuse.EIO
		}
		for _, hash := range res.Commits {
			dirs = append(dirs, gofuse.DirEntry{
				Name: hash,
				Mode: syscall.S_IFDIR,
			})
		}
		return dirs, gofuse.OK
	case len(path) == 1 && path[0] == "branches":
		res, err := f.Client.ListBranches(context.TODO(), &fspb.ListBranchesRequest{})
		if err != nil {
			glog.Errorf("OpenDir(Path=%q) returned error: %v", name, err)
			return nil, errnoFromCode(grpcstat.Convert(err))
		}
		for branchName := range res.Branches {
			dirs = append(dirs, gofuse.DirEntry{
				Name: branchName,
				Mode: syscall.S_IFLNK,
			})
		}
		return dirs, gofuse.OK
	case len(path) >= 2 && path[0] == "commits":
		// Assume path[1] is the commit hash
		var filePath string
		if len(path) == 2 {
			filePath = "/"
		} else {
			filePath = "/" + strings.Join(path[2:], "/")
		}
		res, err := f.Client.ListDir(context.TODO(), &fspb.ListDirRequest{
			Commit: path[1],
			Path:   filePath,
		})
		if err != nil {
			glog.Errorf("ListDir(Commit=%q, Path=%q) returned error: %v", path[1], filePath, err)
			return nil, gofuse.EIO
		}
		for _, entry := range res.Entries {
			dirs = append(dirs, gofuse.DirEntry{
				Name: entry.Name,
				Mode: toSyscallMode(entry.Mode),
			})
		}
		return dirs, gofuse.OK
	}

	return nil, gofuse.ENOENT
}

func (f *GitFS) Symlink(value string, linkName string, ctx *gofuse.Context) gofuse.Status {
	return gofuse.EROFS
}

func (f *GitFS) Readlink(name string, ctx *gofuse.Context) (link string, status gofuse.Status) {
	// TODO: implement
	glog.V(1).Infof("Readlink(name=%q) called", name)
	defer func() {
		if status != gofuse.OK {
			glog.Errorf("Readlink(name=%q) failed: %v", name, status)
		}
	}()

	path := strings.FieldsFunc(name, func(c rune) bool { return c == '/' })

	switch {
	case len(path) == 2 && path[0] == "branches":
		res, err := f.Client.ListBranches(context.TODO(), &fspb.ListBranchesRequest{})
		if err != nil {
			glog.Errorf("Readlink(Path=%q) returned error: %v", name, err)
			return "", errnoFromCode(grpcstat.Convert(err))
		}
		commit, ok := res.Branches[path[1]]
		if !ok {
			glog.Errorf("Readlink(Path=%q) branch %q not found", name, path[1])
			return "", gofuse.ENOENT
		}
		return "../commits/"+commit, gofuse.OK
	}

	return "", gofuse.ENOSYS
}

func (f *GitFS) StatFs(name string) *gofuse.StatfsOut {
	// TODO: implement
	glog.V(1).Infof("StatFs(name=%q) called", name)
	return &gofuse.StatfsOut{}
}

func errnoFromCode(s *grpcstat.Status) gofuse.Status {
	switch s.Code() {
	case codes.NotFound:
		return gofuse.ENOENT
	case codes.Unimplemented:
		return gofuse.ENOSYS
	case codes.Internal:
		return gofuse.EIO
	default:
		return gofuse.EIO
	}
}
