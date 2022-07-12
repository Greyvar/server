package greyvarserver;

import (
	"github.com/gorilla/websocket"
)

type RemotePlayer struct {
	Connection *websocket.Conn;
	Username string;
	NeedsGridUpdate bool;
	Spawned bool;

	Entity *Entity;
}

