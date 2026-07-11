package greyvarserver;

import (
	"os"
	"net/http"
	log "github.com/sirupsen/logrus"
	"time"
	pb "github.com/greyvar/server/gen/greyvarprotocol"
	"github.com/jamesread/golure/pkg/dirs"
	"github.com/greyvar/datlib/entdefs"
	"github.com/greyvar/datlib/gridfiles"
	"github.com/greyvar/server/pkg/worlds"
	"github.com/fsnotify/fsnotify"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"crypto/x509"
	"encoding/pem"
)

type serverInterface struct {
	remotePlayers map[string]*RemotePlayer;
	entityInstances map[int64]*Entity;
	entityDefinitions map[string]*entdefs.EntityDefinition
	loadedWorlds map[string]*worlds.World;
	defaultWorldId string;

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
		log.Warnf("entdef read error for %q: %v", definition, err)
	} else {
		s.entityDefinitions[definition] = entdef
	}
}

func defaultWorldId() string {
	if env := os.Getenv("GREYVAR_WORLD"); env != "" {
		return env
	}

	return "isleOfStarting_dev_tiled"
}

func newServer() *serverInterface {
	s := &serverInterface{};
	s.entityInstances = make(map[int64]*Entity)
	s.entityDefinitions = make(map[string]*entdefs.EntityDefinition)
	blockingTileTextures = loadBlockingTileTextures()
	s.loadServerEntdefs()
	s.remotePlayers = make(map[string]*RemotePlayer);

	s.defaultWorldId = defaultWorldId()
	s.loadedWorlds = make(map[string]*worlds.World)

	if _, err := s.ensureWorldLoaded(s.defaultWorldId); err != nil {
		log.Fatalf("Cannot load world %v: %v", s.defaultWorldId, err)
	}

	return s;
}

func (s *serverInterface) ensureWorldLoaded(worldId string) (*worlds.World, error) {
	if world, ok := s.loadedWorlds[worldId]; ok {
		return world, nil
	}

	world, err := worlds.LoadWorld(worldId)
	if err != nil {
		return nil, err
	}

	s.loadedWorlds[worldId] = world
	log.Infof("Loaded world %v (%v) with %v grids, spawnGrid=%v",
		world.ID, world.Title, len(world.Grids), world.SpawnGrid)

	for gridId, grid := range world.Grids {
		s.instantiateGridEntities(worldId, gridId, grid)
	}

	return world, nil
}

func (s *serverInterface) worldForPlayer(rp *RemotePlayer) *worlds.World {
	if rp == nil {
		return nil
	}

	return s.loadedWorlds[rp.CurrentWorldId]
}

func (s *serverInterface) instantiateGridEntities(worldId string, gridId string, gf *gridfiles.Grid) {
	log.Infof("Loading grid entities for %v/%v", worldId, gridId)

	for _, gfEnt := range gf.Entities {
		s.loadEntdef(gfEnt.Definition)

		entdef, ok := s.entityDefinitions[gfEnt.Definition]
		if !ok {
			log.Warnf("Skipping entity with unknown entdef %q on %v/%v", gfEnt.Definition, worldId, gridId)
			continue
		}

		ent := &Entity{
			Definition: gfEnt.Definition,
			State: entdef.InitialState,
			ServerId: s.nextEntityId(),
			X: int32(gfEnt.Col * 16),
			Y: int32(gfEnt.Row * 16),
			GridId: gridId,
			WorldId: worldId,
			Spawned: true,
		}
	
		s.entityInstances[ent.ServerId] = ent
	}
}

func spawnPositionForGrid(grid *gridfiles.Grid) (int32, int32) {
	row := grid.RowCount / 2
	col := grid.ColCount / 2
	return int32(col * 16), int32(row * 16)
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
					s.migrateGridTiles()
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

func (s *serverInterface) migrateGridTiles() {
	for _, rp := range s.remotePlayers {
		rp.NeedsGridUpdate = true
	}
}

func (s *serverInterface) gridById(worldId string, gridId string) *gridfiles.Grid {
	world := s.loadedWorlds[worldId]
	if world == nil {
		return nil
	}

	return world.Grids[gridId]
}

func (s *serverInterface) mainLoop() {
	for {
		time.Sleep(100 * time.Millisecond)
		s.frame();
	}
}

func (s *serverInterface) onConnected(c *websocket.Conn) (*RemotePlayer) {
	log.Info("Player connected");

	world, err := s.ensureWorldLoaded(s.defaultWorldId)
	if err != nil {
		log.Fatalf("Cannot load default world %v: %v", s.defaultWorldId, err)
	}

	spawnGridId := world.SpawnGrid
	spawnGrid := s.gridById(s.defaultWorldId, spawnGridId)
	if spawnGrid == nil {
		log.Fatalf("Spawn grid %v not found in world %v", spawnGridId, s.defaultWorldId)
	}

	spawnX, spawnY := spawnPositionForGrid(spawnGrid)

	playerEntity := Entity {
		X: spawnX,
		Y: spawnY,
		Definition: "player",
		State: "idle",
		ServerDebugAlias: "player",
		ServerId: s.nextEntityId(),
		GridId: spawnGridId,
		WorldId: s.defaultWorldId,
	}

	s.entityInstances[playerEntity.ServerId] = &playerEntity

	rp := &RemotePlayer {
		Connection: c,
		Username: "bob",
		NeedsGridUpdate: true,
		Spawned: false,
		Entity: &playerEntity,
		CurrentGridId: spawnGridId,
		CurrentWorldId: s.defaultWorldId,
		KnownEntities: make(map[int64]*Entity),
		KnownEntdefs: make(map[string]bool),
		TimeOfLastMoveRequest: s.frameTime,
	}

	connectionFrame := &pb.ServerUpdate{
		ConnectionResponse: &pb.ConnectionResponse {
			ServerVersion: "waffles2",
		},
		Grid: generateGridUpdate(s, rp),
	}

	rp.NeedsGridUpdate = false
	rp.currentFrame = connectionFrame
	frameNewEntdefs(s, rp)

	s.sendServerFrame(rp.currentFrame, rp)

	s.remotePlayers[rp.Username] = rp;

	return rp
}

func (s *serverInterface) onDisconnected(c *websocket.Conn) {
	for username, rp := range s.remotePlayers {
		if rp.Connection == c {
			oldGridId := rp.CurrentGridId
			oldWorldId := rp.CurrentWorldId

			for _, other := range s.remotePlayers {
				if other == rp {
					continue
				}

				if other.CurrentWorldId == oldWorldId && other.CurrentGridId == oldGridId {
					if _, known := other.KnownEntities[rp.Entity.ServerId]; known {
						other.PendingDespawns = append(other.PendingDespawns, rp.Entity.ServerId)
						delete(other.KnownEntities, rp.Entity.ServerId)
					}
				}
			}

			delete(s.entityInstances, rp.Entity.ServerId)
			delete(s.remotePlayers, username)
			return
		}
	}

	log.Warnf("Could not find remoteplayer to remove in onDisconnected")
}

func (server *serverInterface) handleConnection(w http.ResponseWriter, r *http.Request) {
	log.Infof("New connection from %s", r.RemoteAddr)

	client, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Warnf("websocket accept failed: %v", err)
		return
	}
	defer client.Close(websocket.StatusInternalError, "internal error")

	ctx := r.Context()

	rp := server.onConnected(client)

	for {
		reqs := &pb.ClientRequests{}
		err = wsjson.Read(ctx, client, reqs)

		if err != nil {
			log.Warnf("handleConnection read failure: %v", err)
			break
		}

		if reqs.MoveRequest != nil {
			log.WithFields(log.Fields{
				"player":   rp.Username,
				"entityId": rp.Entity.ServerId,
				"x":        reqs.MoveRequest.X,
				"y":        reqs.MoveRequest.Y,
			}).Info("MoveRequest received over websocket")
		}

		rp.pendingRequests = append(rp.pendingRequests, reqs)
	}

	server.onDisconnected(client)

	log.Infof("Closing handler")
}

func findResDir() (string, error) {
	return dirs.GetFirstExistingDirectory("res", []string {
		"../res/",
		"/var/www/html/greyvar-res",
	})
}

func findWebclientDir() (string, error) {
	return dirs.GetFirstExistingDirectory("webclient", []string {
		"/var/www/html/webclient/dist",
		"../webclient/dist/",
	})
}

func printCertificateInfo(certPath string) {
	log.Infof("Reading certificate: %v", certPath)

	certPem, err := os.ReadFile(certPath)

	if err != nil {
		log.Fatalf("Cannot read cert file: %v", err)
	}

	block, _ := pem.Decode(certPem)

	if block == nil || block.Type != "CERTIFICATE" {
		log.Fatalf("Failed to decode certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)

	if err != nil {
		log.Fatalf("Failed to parse certificate: %v", err)
	}

	if len(cert.DNSNames) == 0 {
		log.Fatalf("No SANs in certificate")
	} else {
		for _, name := range cert.DNSNames {
			log.Infof("  SAN: %v", name)
		}

		for _, ip := range cert.IPAddresses {
			log.Infof("  IP : %v", ip)
		}
	}

	log.Infof("  NotBefore: %v", cert.NotBefore)
	log.Infof("  NotAfter: %v", cert.NotAfter)
}

func Start() {
	log.Info("Server starting");

	server := newServer()

	mux := http.NewServeMux()

	resDir, _ := findResDir()
	webclientDir, _ := findWebclientDir()

	mux.HandleFunc("/api", server.handleConnection)
	mux.HandleFunc("/api/debug/blocking-tiles", server.handleBlockingTiles)
	mux.Handle("/res/", http.StripPrefix("/res/", http.FileServer(http.Dir(resDir))))
	mux.Handle("/", http.FileServer(http.Dir(webclientDir)))

	go server.mainLoop();

	cert := "greyvar.crt"
	key := "greyvar.key"

	printCertificateInfo(cert);

	srv := &http.Server {
		Addr: "0.0.0.0:8443",
		Handler: mux,
	}

	log.Infof("Listening on %v", srv.Addr)

	log.Fatal(srv.ListenAndServeTLS(cert, key))
}
