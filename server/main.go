package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	fspb "github.com/minorhacks/funhouse/proto/git_read_fs_proto"
	"github.com/minorhacks/funhouse/service"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"github.com/gorilla/mux"
)

var (
	grpcPort = flag.Int("grpc_port", 8080, "Port of gRPC service")
	httpPort = flag.Int("http_port", 8081, "Port of HTTP service")
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

	httpAddr := net.JoinHostPort("", strconv.FormatInt(int64(*httpPort), 10))
	router := mux.NewRouter()
	router.HandleFunc("/push", s.PushHook).Methods("POST")
	httpServer := &http.Server{
		Handler: router,
		Addr: httpAddr,
		WriteTimeout: 15*time.Second,
		ReadTimeout: 15*time.Second,
	}
	go func() {
		glog.Infof("HTTP server listening on %s", httpAddr)
		glog.Fatal(httpServer.ListenAndServe())
	}()

	glog.Infof("Listening on %s", addr)
	return grpcServer.Serve(conn)
}
