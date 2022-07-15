package greyvarserver;

import (
	"fmt"
	"net/http"
	log "github.com/sirupsen/logrus"
	"time"
	pb "github.com/greyvar/server/gen/greyvarprotocol"
	"github.com/greyvar/server/pkg/gridFileHandler"
	"github.com/greyvar/server/pkg/entdefReader"
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
	remotePlayers map[string]*RemotePlayer;
	entityInstances map[int64]*Entity;
	entityDefinitions map[string]*entdefReader.EntityDefinition
	grids []gridFileHandler.GridFile;

	lastEntityId int64;

	frameTime int64;
}

func (s *serverInterface) nextEntityId() int64 {
	s.lastEntityId += 1;

	return s.lastEntityId;
}

func (s *serverInterface) loadServerEntdefs() {
	log.Infof("Loading server entdefs")

	s.loadEntdef("player")
}

func (s *serverInterface) loadEntdef(definition string) {
	if _, ok := s.entityDefinitions[definition]; ok {
		return
	}

	entdef, err := entdefReader.ReadEntdef(definition)

	if err != nil {
		log.Warnf("entdef read error! %v", entdef)
	} else {
		s.entityDefinitions[definition] = entdef
	}
}

func newServer() *serverInterface {
	s := &serverInterface{};
	s.entityInstances = make(map[int64]*Entity)
	s.entityDefinitions = make(map[string]*entdefReader.EntityDefinition)
	s.loadServerEntdefs()
	s.remotePlayers = make(map[string]*RemotePlayer);
	s.loadGrid("dat/worlds/isleOfStarting_dev/grids/1.1.grid")

	return s;
}

func (s *serverInterface) loadGrid(filename string) {
	gf, err := gridFileHandler.ReadGridFile(filename)

	if err != nil {
		fmt.Printf("Cannot load grid: %v", err)
		return
	}

	log.Infof("Loading grid entdefs")

	for _, gfEnt := range gf.Entities {
		s.loadEntdef(gfEnt.Definition)

		ent := &Entity{
			Definition: gfEnt.Definition,
			ServerId: s.nextEntityId(),
			X: gfEnt.X * 16,
			Y: gfEnt.Y * 16,
			Spawned: true,
		}
	
		s.entityInstances[ent.ServerId] = ent
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
		X: 48,
		Y: 48,
		Definition: "player",
		ServerDebugAlias: "player",
		ServerId: s.nextEntityId(),
	}

	s.entityInstances[playerEntity.ServerId] = &playerEntity

	rp := &RemotePlayer {
		Connection: c,
		Username: "bob",
		NeedsGridUpdate: true,
		Spawned: false,
		Entity: &playerEntity,
		KnownEntities: make(map[int64]*Entity),
		KnownEntdefs: make(map[string]bool),
		TimeOfLastMoveRequest: s.frameTime,
	}

	// NOTE: This player needs to get a ConnectionResponse to get things going
	// before it is added to s.remotePlayers - so we send the connection frame
	// outside of the main frame loop. 
	//
	// If we used rp.currentFrame here, it would get overwritten by the frame()
	// loop before a ConnectionResponse was sent.

	connectionFrame := &pb.ServerUpdate{
		ConnectionResponse: &pb.ConnectionResponse {
			ServerVersion: "waffles2",
		},
	} 

	s.sendServerFrame(connectionFrame, rp)

	// NOTE: Now it's safe to add.

	s.remotePlayers[rp.Username] = rp;

	return rp
}

func (s *serverInterface) onDisconnected(c *websocket.Conn) {
	for _, rp := range s.remotePlayers {
		if rp.Connection == c {
			delete(s.entityInstances, rp.Entity.ServerId)
			delete(s.remotePlayers, rp.Username)
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

