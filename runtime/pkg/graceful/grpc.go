package graceful

import (
	"context"
	"fmt"
	"net"
	"strings"

	"google.golang.org/grpc"
)

// ServeGRPC serves a GRPC server and performs a graceful shutdown if/when ctx is cancelled.
func ServeGRPC(ctx context.Context, server *grpc.Server, port int) error {
	// Calling net.Listen("tcp", ...) will succeed if the port is blocked on IPv4 but not on IPv6.
	// This workaround ensures we get the port on IPv4 (and most likely also on IPv6).
	lis, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err == nil {
		lis.Close()
		lis, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	}
	if err != nil {
		if strings.Contains(err.Error(), "address already in use") {
			return fmt.Errorf("grpc port %d is in use by another process. Either kill that process or pass `--port-grpc PORT` to run Rill on another port", port)
		}
		return err
	}

	cctx, cancel := context.WithCancel(ctx)
	var serveErr error
	go func() {
		serveErr = server.Serve(lis)
		cancel()
	}()

	<-cctx.Done()
	if serveErr == nil {
		server.GracefulStop()
	}

	return serveErr
}
