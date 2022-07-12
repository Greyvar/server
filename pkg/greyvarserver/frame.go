package greyvarserver;

import (
	pb "github.com/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

func (s *serverInterface) frame() {
	s.currentFrame = &pb.ServerFrameResponse{}

	log.WithFields(log.Fields{
		"Players": (s.remotePlayers),
	}).Debug("frame() tick");

	frameGridUpdates(s)

	for _, player := range s.remotePlayers {
		frameSpawnPlayer(s, player)
	}

	// Note that position updates need to come after spawns, so we don't update
	// the position of an entity that has not yet spawned.
	frameEntityPositions(s);

	s.broadcastServerFrame()
}

func (s *serverInterface) broadcastServerFrame() {
	data, err := proto.Marshal(s.currentFrame);

	if err != nil {
		log.Errorf("Could not marshal obj to protobuf in sendMessage: %v", err);
		return
	}

	for _, player := range s.remotePlayers {
		player.Connection.WriteMessage(websocket.BinaryMessage, data)
	}
}


func frameEntityPositions(s *serverInterface) {
	for _, ent := range s.entities {
		entpos := pb.EntityPosition {
			X: ent.X,
			Y: ent.Y,
			EntityId: ent.Id,
		}

		s.currentFrame.EntityPositions = append(s.currentFrame.EntityPositions, &entpos);
	}
}

