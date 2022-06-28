package greyvarserver;

type Entity struct {
	Id int64;

	Texture string;
	Spawned bool;

	X int32;
	Y int32;

	ServerDebugAlias string;
}
