package greyvarserver;

import (
	"context"
	"fmt"
	"net"
	log "github.com/sirupsen/logrus"
	"time"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	pb "github.com/greyvar/server/pkg/greyvarproto"
	"github.com/greyvar/server/pkg/gridFileHandler"
)

type serverInterface struct {
	remotePlayers map[int64]*RemotePlayer;
	entities []*Entity;
	grids []gridFileHandler.GridFile;

	lastEntityId int64;
}

func (s *serverInterface) nextEntityId() int64 {
	s.lastEntityId += 1;

	return s.lastEntityId;
}

func newServer() *serverInterface {
	s := &serverInterface{};
	s.remotePlayers = make(map[int64]*RemotePlayer);
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
	log.Debug("server tick");
}

func (s *serverInterface) Connect(ctx context.Context, req *pb.ConnectionRequest) (*pb.ConnectionResponse, error) {
	res := new(pb.ConnectionResponse);
	res.ServerVersion = "waffles2";
	log.Info("Player connected");

	return res, nil;
}

func (s *serverInterface) PlayerSetup(ctx context.Context, plr *pb.NewPlayer) (*pb.NoResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx);

	log.WithFields(log.Fields{
		"uuid": md,
	}).Info("PlayerSetup");

	// Register an entity for this new player. In the next server frame the
	// unspawned player will spawn an entity.
	playerEntity := Entity {
		X: 64,
		Y: 64,
		ServerDebugAlias: "player",
		Id: s.nextEntityId(),
	}

	s.entities = append(s.entities, &playerEntity);

	rp := RemotePlayer {
		Username: "bob",
		NeedsGridUpdate: true,
		Spawned: false,
		Entity: &playerEntity,
	}

	s.remotePlayers[playerEntity.Id] = &rp;

	return new(pb.NoResponse), nil;
}

func Start() {
	log.Info("Server starting");

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

