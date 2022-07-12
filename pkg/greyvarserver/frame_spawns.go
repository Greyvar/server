package greyvarserver

import(
	pb "github.com/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
)

func frameSpawnPlayer(s *serverInterface, rp *RemotePlayer) {
	if !rp.Spawned {
		log.WithFields(log.Fields {
			"player": rp.Username,
			"spawned": rp.Spawned,
		}).Info("Spawning player");

		// Spawn this entity for all players in the next server frame.
		spawn := pb.EntitySpawn{}
		spawn.EntityId = rp.Entity.Id;
		spawn.PrimaryColor = 0xff0000ff;
		spawn.X = rp.Entity.X;
		spawn.Y = rp.Entity.Y;
		spawn.Texture = "playerBob.png"

		s.currentFrame.EntitySpawns = append(s.currentFrame.EntitySpawns, &spawn);

		// Now send a joining message
		playerJoin :=  pb.PlayerJoined{}
		playerJoin.Username = rp.Username;

		s.currentFrame.PlayerJoined = &playerJoin;

		rp.Spawned = true
	}
}
