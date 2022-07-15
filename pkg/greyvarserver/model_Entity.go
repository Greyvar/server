package greyvarserver;

type Entity struct {
	ServerId int64;

	Texture string;
	Spawned bool;

	Definition string;
	State string;

	X int32;
	Y int32;

	ServerDebugAlias string;
}
