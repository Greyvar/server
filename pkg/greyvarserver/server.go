package greyvarserver;

import (
	"os"
	"context"
	"net/http"
	log "github.com/sirupsen/logrus"
	"time"
	pb "github.com/greyvar/server/gen/greyvarprotocol"
	"github.com/jamesread/golure/pkg/dirs"
	"github.com/greyvar/datlib/entdefs"
	"github.com/greyvar/datlib/gridfiles"
	"github.com/fsnotify/fsnotify"
	"crypto/x509"
	"encoding/pem"
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
	log.Infof("New connection")

	client, err := websocket.Accept(w, r, nil)
	defer client.Close(websocket.StatusInternalError, "internal error")
	
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer cancel()

	log.Infof("New Handler")

	rp := server.onConnected(client)

	for {
		reqs := &pb.ClientRequests{}
		err = wsjson.Read(ctx, client, reqs)

		log.Infof("C: %+v", client);

		if err != nil {
			log.Infof("")
			log.Warnf("handleConnection unmarshal failure: %v", err)
			break
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

	// Print certificate expiry
	log.Infof("  NotBefore: %v", cert.NotBefore)
	log.Infof("  NotAfter: %v", cert.NotAfter)
}

func getNewApiHandler() (string, http.Handler, *http.Server) {
	apiServer := api.NewServer()

	path, handler := clientapiconnect.NewGreyvarApiHandler(apiServer)

	return path, handler, apiServer
}

func Start() {
	log.Info("Server starting");

	server := newServer()

	mux := http.NewServeMux()

	resDir, _ := findResDir()
	webclientDir, _ := findWebclientDir()

	apipath, apihandler, apiserver := getNewApiHandler()

	mux.HandleFunc("/api", server.handleConnection)
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

