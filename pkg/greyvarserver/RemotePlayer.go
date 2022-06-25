package greyvarserver;

import (
)

type RemotePlayer struct {
	Username string;
	NeedsGridUpdate bool;
	Spawned bool;

	Entity *Entity;
}
