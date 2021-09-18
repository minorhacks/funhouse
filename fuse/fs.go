package fuse

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"
	"regexp"

	"github.com/minorhacks/funhouse/service"

	"github.com/golang/glog"
	gofuse "github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

var (
	commitHashPattern = regexp.MustCompile(`[0-9a-f]{40}`)
)

type GitFS struct {
	ServerAddr string
}

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

	path := strings.FieldsFunc(name, func(c rune) bool { return c == os.PathSeparator })
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
	case len(path) >= 2 && path[0] == "commits":
		// Assume path[1] is the commit hash
		client := &http.Client{}
		var filePath string
		if len(path) == 2 {
			filePath = "/"
		} else {
			filePath = "/" + strings.Join(path[2:], "/")
		}
		r := &service.GetAttrRequest{
			CommitHash: path[1],
			Path: filePath,
		}

		body, err := json.Marshal(r)
		if err != nil {
			glog.Errorf("failed to marshal GetAttrRequest: %v", err)
			return nil, gofuse.EIO
		}

		req, err := http.NewRequest("GET", (&url.URL{Scheme: "http", Host: f.ServerAddr, Path: "getattr"}).String(), bytes.NewReader(body))
		if err != nil {
			glog.Errorf("failed to get attrs for %q: %v", filePath, err)
			return nil, gofuse.EIO
		}
		res, err := client.Do(req)
		if err != nil {
			glog.Errorf("failed to get attrs for %q: %v", filePath, err)
			return nil, gofuse.EIO
		}
		defer res.Body.Close()
		var resData service.GetAttrResponse
		err = json.NewDecoder(res.Body).Decode(&resData)
		if err != nil {
			glog.Errorf("failed to parse attr response: %v", err)
			return nil, gofuse.EIO
		}
		return &gofuse.Attr{
			Mode: resData.FileMode.ToSyscallMode(),
			Size: resData.Size,
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
	return gofuse.ENOSYS
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

	client := &http.Client{}
	r := &service.CatFileRequest{
		CommitHash: path[1],
		Path: strings.Join(path[2:], "/"),
	}
	body, err := json.Marshal(r)
	if err != nil {
		glog.Errorf("failed to marshal CatFileRequest: %v", err)
		return nil, gofuse.EIO
	}
	req, err := http.NewRequest("GET", (&url.URL{Scheme: "http", Host: f.ServerAddr, Path: "cat"}).String(), bytes.NewReader(body))
	if err != nil {
		glog.Errorf("failed to cat file: %v", err)
		return nil, gofuse.EIO
	}
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("failed to cat file: %v", err)
		return nil, gofuse.EIO
	}
	defer res.Body.Close()
	fileBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Failed to read from response body: %v", err)
	}
	return nodefs.NewDataFile(fileBody), gofuse.OK
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

	path := strings.FieldsFunc(name, func(c rune) bool { return c == os.PathSeparator })
	// Simulate top-level dirs
	switch {
	case len(path) == 0:
		return []gofuse.DirEntry{
			{
				Name: "commits",
				Mode: syscall.S_IFDIR,
			},
		}, gofuse.OK
	case len(path) == 1 && path[0] == "commits":
		client := &http.Client{}
		r := &service.ListCommitsRequest{}
		body, err := json.Marshal(r)
		if err != nil {
			glog.Errorf("failed to marshal ListCommitsRequest: %v", err)
			return nil, gofuse.EIO
		}
		req, err := http.NewRequest("GET", (&url.URL{Scheme: "http", Host: f.ServerAddr, Path: "commits"}).String(), bytes.NewReader(body))
		if err != nil {
			glog.Errorf("failed to get commit list: %v", err)
			return nil, gofuse.EIO
		}
		res, err := client.Do(req)
		if err != nil {
			glog.Errorf("failed to get commit list: %v", err)
			return nil, gofuse.EIO
		}
		defer res.Body.Close()
		var resData service.ListCommitsResponse
		err = json.NewDecoder(res.Body).Decode(&resData)
		if err != nil {
			glog.Errorf("failed to parse commit list response: %v", err)
			return nil, gofuse.EIO
		}
		for _, hash := range resData.CommitHashes {
			dirs = append(dirs, gofuse.DirEntry{
				Name: hash,
				Mode: syscall.S_IFDIR,
			})
		}
		return dirs, gofuse.OK
	case len(path) >= 2 && path[0] == "commits":
		// Assume path[1] is the commit hash
		client := &http.Client{}
		var filePath string
		if len(path) == 2 {
			filePath = "/"
		} else {
			filePath = "/" + strings.Join(path[2:], "/")
		}
		r := &service.ListDirRequest{
			CommitHash: path[1],
			Path: filePath,
		}

		body, err := json.Marshal(r)
		if err != nil {
			glog.Errorf("failed to marshal ListCommitsRequest: %v", err)
			return nil, gofuse.EIO
		}

		req, err := http.NewRequest("GET", (&url.URL{Scheme: "http", Host: f.ServerAddr, Path: "listdir"}).String(), bytes.NewReader(body))
		if err != nil {
			glog.Errorf("failed to get dir entry list: %v", err)
			return nil, gofuse.EIO
		}
		res, err := client.Do(req)
		if err != nil {
			glog.Errorf("failed to get dir entry list: %v", err)
			return nil, gofuse.EIO
		}
		defer res.Body.Close()
		var resData service.ListDirResponse
		err = json.NewDecoder(res.Body).Decode(&resData)
		if err != nil {
			glog.Errorf("failed to parse dir entry list response: %v", err)
			return nil, gofuse.EIO
		}
		for _, entry := range resData.Entries {
			dirs = append(dirs, gofuse.DirEntry{
				Name: entry.Name,
				Mode: entry.FileMode.ToSyscallMode(),
			})
		}
		return dirs, gofuse.OK
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
	return "", gofuse.ENOSYS
}

func (f *GitFS) StatFs(name string) *gofuse.StatfsOut {
	// TODO: implement
	glog.V(1).Infof("StatFs(name=%q) called", name)
	return &gofuse.StatfsOut{}
}
