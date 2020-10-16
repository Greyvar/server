package greyvarserver;

import (
	"context"
	"fmt"
	"net"
	"log"
	"time"
	"google.golang.org/grpc"
	pb "github.com/greyvar/server/pkg/greyvarproto"
	"github.com/greyvar/server/pkg/gridFileHandler"
)

type serverInterface struct {
	remotePlayers []RemotePlayer;
	grids []gridFileHandler.GridFile;
}

func newServer() *serverInterface {
	s := &serverInterface{};
	s.loadGrid("dat/worlds/isleOfStarting_dev/grids/1.1.grid")

	return s;
}

func (s *serverInterface) loadGrid(filename string) {
	gf, err := gridFileHandler.ReadGridFile(filename)

	if err != nil {
		fmt.Printf("Cannot load grid: %v", err)
		return
	}

	s.grids = append(s.grids, *gf);
}

func (s *serverInterface) mainLoop() {
	for {
		time.Sleep(1 * time.Second)
		s.tick();
	}
}

func (s *serverInterface) tick() {
	fmt.Println("server tick");
}

func (s *serverInterface) Connect(ctx context.Context, req *pb.ConnectionRequest) (*pb.ConnectionResponse, error) {
	res := new(pb.ConnectionResponse);
	res.ServerVersion = "waffles2";
	fmt.Println("Player connected");
	return res, nil;
}

func (s *serverInterface) PlayerSetup(ctx context.Context, plr *pb.NewPlayer) (*pb.NoResponse, error) {
	rp := RemotePlayer {
		Username: "bob",
		NeedsGridUpdate: true,
		Spawned: false,
		X: 32,
		Y: 32,
	}

	fmt.Println("PlayerSetup");

	s.remotePlayers = append(s.remotePlayers, rp);
	return new(pb.NoResponse), nil;
}

func Start() {
	fmt.Println("Server starting");

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 2000));

	if (err != nil) {
		log.Fatalf("failed to listen %v ", err);
	}

	server := newServer();

	grpcServer := grpc.NewServer();
	pb.RegisterServerInterfaceServer(grpcServer, server);

	go server.mainLoop();

	grpcServer.Serve(lis);
}

