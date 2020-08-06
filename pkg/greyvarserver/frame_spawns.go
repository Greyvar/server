package greyvarserver

import(
	pb "github.com/greyvar/server/pkg/greyvarproto"
)

func FramePlayerSpawns(s *serverInterface, serverFrame *pb.ServerFrameResponse, rp RemotePlayer) {
	if !rp.Spawned {
		spawn := pb.EntitySpawn{}
		spawn.X = rp.X;
		spawn.Y = rp.Y;
		spawn.Texture = "playerBob.png"

		serverFrame.EntitySpawn = &spawn;

		rp.Spawned = true
	}
}
