package greyvarserver;

import (
)

type RemotePlayer struct {
	Username string;
	NeedsGridUpdate bool;
	Spawned bool;

	X int32;
	Y int32;
}
