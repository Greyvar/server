package greyvarserver;

import (
	"fmt"
	"net/http"
	log "github.com/sirupsen/logrus"
	"time"
	pb "github.com/greyvar/server/gen/greyvarprotocol"
	"github.com/greyvar/server/pkg/gridFileHandler"
	"github.com/gorilla/websocket"
	"github.com/golang/protobuf/proto"
)

var upgrader = websocket.Upgrader {
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type serverInterface struct {
	remotePlayers map[int64]*RemotePlayer;
	entities []*Entity;
	grids []gridFileHandler.GridFile;

	nextFrame pb.ServerFrameResponse;

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

func (s *serverInterface) onConnected(c *websocket.Conn) {
	log.Info("Player connected");

	res := new(pb.ConnectionResponse);
	res.ServerVersion = "waffles2";

	s.nextFrame = pb.ServerFrameResponse{}
	s.nextFrame.ConnectionResponse = res
	s.sendFrame(c)
}

func (s *serverInterface) sendFrame(c *websocket.Conn) {
	data, err := proto.Marshal(&s.nextFrame);

	if err != nil {
		log.Errorf("Could not marshal obj to protobuf in sendMessage: %v", err);
		return
	}


	log.Infof("msg len = %v", len(data))

	c.WriteMessage(websocket.BinaryMessage, data)

	s.nextFrame = pb.ServerFrameResponse{}
}

func (s *serverInterface) playerSetup(c *websocket.Conn) (*pb.NoResponse, error) {
	md := "?"

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

func (server *serverInterface) handleClientRequests(reqs *pb.ClientRequests) {
	log.Infof("handleClientRequests %v", reqs)
}

func (server *serverInterface) handleConnection(c *websocket.Conn) {
	log.Infof("New Handler")

	server.onConnected(c)

	for {
		_, rawMessage, err := c.ReadMessage()

		if err != nil {
			log.Warnf("Conn readMessage fail: %v", err)
			break
		}

		reqs := &pb.ClientRequests{}
		err = proto.Unmarshal(rawMessage, reqs)

		if err != nil {
			log.Warnf("Unmarshal failure: %v", err)
		}

		server.handleClientRequests(reqs)
	}

	log.Infof("Closing handler")
}

func Start() {
	log.Info("Server starting");

	server := newServer()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("New conn")

		c, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.Errorf("Upgrade: %v", err)
			return
		}

		go server.handleConnection(c)
	})

	go server.mainLoop();

	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

