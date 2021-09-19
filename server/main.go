package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	fspb "github.com/minorhacks/funhouse/proto/git_read_fs_proto"
	"github.com/minorhacks/funhouse/service"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	grpcPort = flag.Int("grpc_port", 8080, "Port of gRPC service")
	basePath = flag.String("base_path", "/tmp/funhouse", "Path to store cloned repository data")
	repoURL  = flag.String("repo_url", "", "Clone and serve a single repository at this URL")
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
	if *repoURL == "" {
		return fmt.Errorf("--repo_url must be set")
	} else {
		s, err = service.New(*basePath, *repoURL)
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
