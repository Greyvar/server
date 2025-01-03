package greyvarserver;

import (
	"os"
	"context"
	"path/filepath"
	"net/http"
	log "github.com/sirupsen/logrus"
	"time"
	pb "github.com/greyvar/server/gen/greyvarprotocol"
	"github.com/greyvar/datlib/entdefs"
	"github.com/greyvar/datlib/gridfiles"
	"github.com/fsnotify/fsnotify"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type serverInterface struct {
	remotePlayers map[string]*RemotePlayer;
	entityInstances map[int64]*Entity;
	entityDefinitions map[string]*entdefs.EntityDefinition
	grids []*gridfiles.Grid;

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

	entdef, err := entdefs.ReadEntdef(definition)

	if err != nil {
		log.Warnf("entdef read error! %v", entdef)
	} else {
		s.entityDefinitions[definition] = entdef
	}
}

func newServer() *serverInterface {
	s := &serverInterface{};
	s.entityInstances = make(map[int64]*Entity)
	s.entityDefinitions = make(map[string]*entdefs.EntityDefinition)
	s.loadServerEntdefs()
	s.remotePlayers = make(map[string]*RemotePlayer);
	
	filename := "dat/worlds/gen/grids/0.grid"

	s.loadGrid(filename)
	go s.watchGridFile(filename)

	return s;
}

func (s *serverInterface) loadGrid(filename string) *gridfiles.Grid {
	gf, err := gridfiles.ReadGrid(filename)

	if err != nil {
		log.Errorf("Cannot load grid: %v", err)
		return nil
	}

	log.Infof("Loading grid entdefs")

	for _, gfEnt := range gf.Entities {
		s.loadEntdef(gfEnt.Definition)

		ent := &Entity{
			Definition: gfEnt.Definition,
			State: s.entityDefinitions[gfEnt.Definition].InitialState, 
			ServerId: s.nextEntityId(),
			X: int32(gfEnt.Row * 16),
			Y: int32(gfEnt.Col * 16),
			Spawned: true,
		}
	
		s.entityInstances[ent.ServerId] = ent
	}

	s.grids = append(s.grids, gf);

	return gf
}

func (s *serverInterface) watchGridFile(filename string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		lastProcessedEvent := time.Now().Unix()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Errorf("The grid inotify watcher failed")
					return
				}

				if ((time.Now().Unix() - lastProcessedEvent) < 2) {
					log.Warnf("Debouncing probable duplicate inotify event")
					continue
				}

				lastProcessedEvent = time.Now().Unix()

				log.Printf("event: %v\n", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					oldGrid := s.findGridByFilename(filename)
					newGrid := s.loadGrid(filename)

					s.migrateGridTiles(oldGrid, newGrid)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(filename)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func (s *serverInterface) unloadGrid(tounload *gridfiles.Grid) {
	for _, g := range s.grids {
		if g == tounload {
			log.Infof("unloading grid")
			return
		}
	}

	log.Errorf("Cannot find the grid to unload!")
}

func (s *serverInterface) migrateGridTiles(oldGrid *gridfiles.Grid, newGrid *gridfiles.Grid) {
	/**
	for _, oldTile := range oldGrid.Tiles {
		for _, newTile := range newGrid.Tiles {
			if oldTile.Row == newTile.Row && oldTile.Col == newTile.Col {
				oldTile.Texture = newTile.Texture
				continue
			}
		}
	}
	**/

	for _, rp := range s.remotePlayers {
		rp.NeedsGridUpdate = true
	}
}

func (s *serverInterface) findGridByFilename(filename string) *gridfiles.Grid {
	for _, g := range s.grids {
		if g.Filename == filename {
			return g;
		}
	}

	log.Errorf("Could not find grid by filename: %v", filename)

	return nil;
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
		X: 16 * 9,
		Y: 16 * 8,
		Definition: "player",
		State: "idle",
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

func (server *serverInterface) handleConnection(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	defer c.Close(websocket.StatusInternalError, "internal error")
	
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer cancel()

	log.Infof("New Handler")

	rp := server.onConnected(c)

	for {
		reqs := &pb.ClientRequests{}
		err = wsjson.Read(ctx, c, reqs)

		if err != nil {
			log.Warnf("Unmarshal failure: %v", err)
			break
		}

		rp.pendingRequests = append(rp.pendingRequests, reqs)
	}

	server.onDisconnected(c)

	log.Infof("Closing handler")
}

func findResDir() string {
	return firstExistingDir("res", []string {
		"/var/www/html/greyvar-client/res",
		"../webclient/res/",
	})
}

func findWebclientDir() string {
	return firstExistingDir("webclient", []string {
		"../webclient/dist/",
	})
}

func firstExistingDir(name string, directoriesToSearch []string) string {
	for _, dir := range directoriesToSearch {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			dir, _ := filepath.Abs(dir)
			log.Infof("Found %v dir: %v", name, dir)
			return dir 
		}
	}

	log.Fatalf("Could not find %v dir", name)

	return ""
}

func Start() {
	log.Info("Server starting");

	server := newServer()

	mux := http.NewServeMux()

	mux.HandleFunc("/api", server.handleConnection)
	mux.Handle("/res/", http.StripPrefix("/res/", http.FileServer(http.Dir(findResDir()))))
	mux.Handle("/", http.FileServer(http.Dir(findWebclientDir())))

	go server.mainLoop();

	cert := "greyvar.crt"
	key := "greyvar.key" 

	srv := &http.Server {
		Addr: "0.0.0.0:8443",
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServeTLS(cert, key))
}

