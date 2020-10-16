package greyvarserver

import(
	pb "github.com/greyvar/server/pkg/greyvarproto"
	log "github.com/sirupsen/logrus"
)

func FramePlayerSpawns(s *serverInterface, serverFrame *pb.ServerFrameResponse, rp *RemotePlayer) {
	if !rp.Spawned {
		log.WithFields(log.Fields {
			"player": rp.Username,
			"spawned": rp.Spawned,
		}).Info("Spawning player");

		spawn := pb.EntitySpawn{}
		spawn.X = rp.X;
		spawn.Y = rp.Y;
		spawn.Texture = "playerBob.png"

		serverFrame.EntitySpawn = &spawn;

		rp.Spawned = true
	}
}
