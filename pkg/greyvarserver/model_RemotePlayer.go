package greyvarserver;

import (
	"github.com/coder/websocket"
	pb "github.com/greyvar/server/gen/greyvarprotocol"
)

type RemotePlayer struct {
	Connection *websocket.Conn;
	Username string;
	NeedsGridUpdate bool;
	Spawned bool;

	Entity *Entity;

	CurrentGridId string;
	CurrentWorldId string;

	PendingGridTransition *GridTransitionInfo;

	KnownEntities map[int64]*Entity;
	KnownEntdefs map[string]bool
	PendingDespawns []int64;
	PendingConsoleMessages []string;

	TimeOfLastMoveRequest int64;

	currentFrame *pb.ServerUpdate;

	pendingRequests []*pb.ClientRequests;
}

