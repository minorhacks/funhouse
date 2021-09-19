package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/minorhacks/funhouse/service"
	fspb "github.com/minorhacks/funhouse/proto/git_read_fs_proto"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	grpcPort = flag.Int("grpc_port", 8080, "Port of gRPC service")
	basePath   = flag.String("base_path", "/tmp/funhouse", "Path to store cloned repository data")
	singleRepo = flag.String("single_repo", "", "If set, clone and serve a single repository at this URL")
)

func main() {
	flag.Parse()
	if err := app(); err != nil {
		glog.Error(err)
		os.Exit(1)
	}
}

func app() error {
	var s *service.Service
	var err error
	if *singleRepo == "" {
		return fmt.Errorf("--single_repo must be set")
	} else {
		s, err = service.NewSingle(*basePath, *singleRepo)
	}
	if err != nil {
		return fmt.Errorf("failed to create service: %v", err)
	}

	addr := net.JoinHostPort("", strconv.FormatInt(int64(*grpcPort), 10))
	conn, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %q: %v", addr, err)
	}
	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)
	fspb.RegisterGitReadFsServer(grpcServer, s)
	reflection.Register(grpcServer)
	glog.Infof("Listening on %s", addr)
	return grpcServer.Serve(conn)
}
