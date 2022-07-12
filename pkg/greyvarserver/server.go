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
	entities map[int64]*Entity;
	grids []gridFileHandler.GridFile;

	currentFrame *pb.ServerFrameResponse;

	lastEntityId int64;
}

func (s *serverInterface) nextEntityId() int64 {
	s.lastEntityId += 1;

	return s.lastEntityId;
}

func newServer() *serverInterface {
	s := &serverInterface{};
	s.entities = make(map[int64]*Entity)
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
		time.Sleep(100 * time.Millisecond)
		s.frame();
	}
}

func (s *serverInterface) onConnected(c *websocket.Conn) (*RemotePlayer) {
	log.Info("Player connected");

	// Register an entity for this new player. In the next server frame the
	// unspawned player will spawn an entity.
	playerEntity := Entity {
		X: 64,
		Y: 64,
		ServerDebugAlias: "player",
		Id: s.nextEntityId(),
	}

	s.entities[playerEntity.Id] = &playerEntity

	rp := RemotePlayer {
		Connection: c,
		Username: "bob",
		NeedsGridUpdate: true,
		Spawned: false,
		Entity: &playerEntity,
	}

	res := new(pb.ConnectionResponse);
	res.ServerVersion = "waffles2";

	s.currentFrame.ConnectionResponse = res

	s.remotePlayers[playerEntity.Id] = &rp;

	return &rp
}

func (s *serverInterface) onDisconnected(c *websocket.Conn) {
	for i, rp := range s.remotePlayers {
		if rp.Connection == c {
			delete(s.entities, rp.Entity.Id)
			delete(s.remotePlayers, i)
			return
		}
	}

	log.Warnf("Could not find remoteplayer to remove in onDisconnected")
}

func (server *serverInterface) handleConnection(c *websocket.Conn) {
	log.Infof("New Handler")

	rp := server.onConnected(c)

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

		server.handleClientRequests(rp, reqs)
	}

	server.onDisconnected(c)

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

