package grpc

/**
 * Created by  ♅ Salfi Farooq on 23/06/17.
 */

import (
	"net"
	"google.golang.org/grpc"
	"github.com/graphicweave/ox/ox_idl"
	"github.com/graphicweave/ox/grpc/event"
)

func StartServer( addr string )(error) {
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	ox_idl.RegisterOXServer(grpcServer, &event.OXService{})
	grpcServer.Serve(listener)

	return nil
}
