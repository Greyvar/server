package greyvarserver;

import (
	"github.com/gorilla/websocket"
	pb "github.com/greyvar/server/gen/greyvarprotocol"
)

type RemotePlayer struct {
	Connection *websocket.Conn;
	Username string;
	NeedsGridUpdate bool;
	Spawned bool;

	Entity *Entity;

	KnownEntities map[int64]*Entity;
	KnownEntdefs map[string]bool

	TimeOfLastMoveRequest int64;

	currentFrame *pb.ServerUpdate;
}

