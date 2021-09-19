package fuse

import (
	"syscall"
	"fmt"
	
	fspb "github.com/minorhacks/funhouse/proto/git_read_fs_proto"
)

func toSyscallMode(m fspb.FileMode) uint32 {
	switch m {
	case fspb.FileMode_MODE_DIR:
		return syscall.S_IFDIR | 0o555
	case fspb.FileMode_MODE_REGULAR:
		return syscall.S_IFREG | 0o444
	case fspb.FileMode_MODE_EXECUTABLE:
		return syscall.S_IFREG | 0o555
	case fspb.FileMode_MODE_SYMLINK:
		return syscall.S_IFLNK | 0o555
	default:
		panic(fmt.Sprintf("Unhandled filemode: %v", m))
	}
}
