package greyvarserver

import(
	pb "github.com/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
)

func frameSpawnEntities(s *serverInterface, rp* RemotePlayer) {
	for _, entinst := range s.entityInstances {
		if _, known := rp.KnownEntities[entinst.ServerId]; !known {
			//entdef := s.entityDefinitions[entinst.Definition]

			log.WithFields(log.Fields {
				"ent": entinst.Definition,
				"rp": rp.Username,
			}).Info("Spawning entity for player")

			spawn := pb.EntitySpawn{}
			spawn.Definition = entinst.Definition
			spawn.InitialState = entinst.State
			spawn.EntityId = entinst.ServerId
			spawn.X = entinst.X
			spawn.Y = entinst.Y
			spawn.PrimaryColor = 0xff0000ff;

			rp.currentFrame.EntitySpawns = append(rp.currentFrame.EntitySpawns, &spawn)

			rp.KnownEntities[entinst.ServerId] = entinst;
		}
	}
}

func frameSpawnPlayer(s *serverInterface, rp *RemotePlayer) {
	if !rp.Spawned {
		log.WithFields(log.Fields {
			"player": rp.Username,
		}).Info("Spawning player");

		// The entity spawner will take care of spawning this RP's entity

		playerJoin :=  pb.PlayerJoined{}
		playerJoin.Username = rp.Username;

		for _, player := range s.remotePlayers {
			player.currentFrame.PlayerJoined = &playerJoin;
		}

		rp.Spawned = true
	}
}
