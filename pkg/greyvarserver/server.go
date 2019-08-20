package greyvarserver;

import (
	"context"
	"fmt"
	"net"
	"log"
	"google.golang.org/grpc"
	pb "github.com/greyvar/server/pkg/greyvarproto"
)

type serverInterface struct {

}

func newServer() *serverInterface {
	s := &serverInterface{};

	return s;
}

func (s *serverInterface) Connect(ctx context.Context, req *pb.ConnectionRequest) (*pb.ConnectionResponse, error) {
	res := new(pb.ConnectionResponse);
	res.ServerVersion = "waffles2";
	fmt.Println("Player connected");
	return res, nil;
}

func (s *serverInterface) PlayerSetup(ctx context.Context, plr *pb.NewPlayer) (*pb.NoResponse, error) {
	return nil, nil;
}

func (s *serverInterface) GetServerFrame(ctx context.Context, req *pb.ClientRequests) (*pb.ServerFrameResponse, error) {
	u := new(pb.PlayerYou);
	u.PlayerId = 137;

	res := new(pb.ServerFrameResponse);
	res.PlayerYou = u;

	return res, nil;
}

func Start() {
	fmt.Println("Server starting");

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 2000));

	if (err != nil) {
		log.Fatalf("failed to listen %v ", err);
	}

	grpcServer := grpc.NewServer();
	pb.RegisterServerInterfaceServer(grpcServer, newServer());
	grpcServer.Serve(lis);
}

