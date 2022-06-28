package greyvarserver;

import (
	"context"
	pb "github.com/greyvar/server/gen/greyvarprotocol"
	log "github.com/sirupsen/logrus"
)

/**
A server frame represents the entire state of the game. Client's can request 
updates, such as moving a player. The update may or may not be accepted by the
server (eg, player is trying to walk through a wall, or even walk too quickly). 
*/
func (s *serverInterface) GetServerFrame(ctx context.Context, req *pb.ClientRequests) (*pb.ServerFrameResponse, error) {
	processClientRequests(s, req);

	return buildServerFrame(s), nil;
}

func frameEntityPositions(s *serverInterface, frame *pb.ServerFrameResponse) {
	for _, ent := range s.entities {
		entpos := pb.EntityPosition {
			X: ent.X,
			Y: ent.Y,
			EntityId: ent.Id,
		}

		frame.EntityPositions = append(frame.EntityPositions, &entpos);
	}
}

func buildServerFrame(s *serverInterface) *pb.ServerFrameResponse {
	frame := new(pb.ServerFrameResponse);

	frameGridUpdates(s, frame)

	for _, player := range s.remotePlayers {
		FramePlayerSpawns(s, frame, player)
	}

	// Note that position updates need to come after spawns, so we don't update
	// the position of an entity that has not yet spawned.
	frameEntityPositions(s, frame);

	log.WithFields(log.Fields{
		"serverFrame": frame,
	}).Debug("ServerFrame");

	return frame;
}

