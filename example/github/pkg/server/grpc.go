package server

import (
	"context"

	igrpc "github.com/at15/go.ice/ice/transport/grpc"
	dlog "github.com/dyweb/gommon/log"

	pb "github.com/at15/go.ice/example/github/pkg/icehubpb"
	mygrpc "github.com/at15/go.ice/example/github/pkg/transport/grpc"
)

var _ mygrpc.IceHubServer = (*GrpcServer)(nil)

type GrpcServer struct {
	log *dlog.Logger
}

func NewGrpcServer() (*GrpcServer, error) {
	s := &GrpcServer{}
	dlog.NewStructLogger(log, s)
	return s, nil
}

func (srv *GrpcServer) Ping(ctx context.Context, ping *pb.Ping) (*pb.Pong, error) {
	addr := igrpc.RemoteAddr(ctx)
	srv.log.Infof("got ping from addr %s", addr)
	return &pb.Pong{Name: "pong to addr " + addr + " your name is " + ping.Name}, nil
}
