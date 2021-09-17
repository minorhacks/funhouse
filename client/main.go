package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/minorhacks/funhouse/fuse"

	"github.com/golang/glog"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

var (
	mountPoint = flag.String("mount_point", "", "Location where filesystem should be mounted")

	entryTTL    = flag.Float64("entry_ttl", 1.0, "FUSE entry cache TTL")
	negativeTTL = flag.Float64("negative_ttl", 1.0, "FUSE negative entry cache TTL")
)

func main() {
	flag.Parse()

	if err := app(); err != nil {
		glog.Error(err)
		os.Exit(1)
	}
}

func app() error {
	fs := &fuse.GitFS{}
	pathNodeFs := pathfs.NewPathNodeFs(fs, &pathfs.PathNodeFsOptions{})
	mountState, _, err := nodefs.MountRoot(*mountPoint, pathNodeFs.Root(), &nodefs.Options{
		EntryTimeout:    time.Duration(*entryTTL * float64(time.Second)),
		AttrTimeout:     time.Duration(*entryTTL * float64(time.Second)),
		NegativeTimeout: time.Duration(*negativeTTL * float64(time.Second)),
		PortableInodes:  false,
	})
	if err != nil {
		return fmt.Errorf("failed to mount: %v", err)
	}

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		s := <-sigChan
		glog.Infof("Caught %v; unmounting", s)
		if err := mountState.Unmount(); err != nil {
			glog.Errorf("Error while unmounting: %v", err)
		}
	}()

	mountState.Serve()
	return nil
}
