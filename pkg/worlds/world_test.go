package worlds

import (
	"testing"

	"github.com/greyvar/datlib/gridfiles"
	"github.com/greyvar/datlib/worldfiles"
)

func TestAdjacentGridIdTMJ(t *testing.T) {
	world := &World{
		Grids: map[string]*gridfiles.Grid{
			"1.1.tmj": {},
			"1.2.tmj": {},
			"0.1.tmj": {},
		},
	}

	dest, ok := AdjacentGridId(world, "1.1.tmj", 0, 1)
	if !ok || dest != "1.2.tmj" {
		t.Fatalf("AdjacentGridId right = %q,%v want 1.2.tmj,true", dest, ok)
	}

	dest, ok = AdjacentGridId(world, "1.1.tmj", -1, 0)
	if !ok || dest != "0.1.tmj" {
		t.Fatalf("AdjacentGridId up = %q,%v want 0.1.tmj,true", dest, ok)
	}
}

func TestScrollDeltaBetweenWorld(t *testing.T) {
	world := &World{
		Definition: &worldfiles.Definition{
			Maps: []worldfiles.MapPlacement{
				{FileName: "1.1.tmj", X: 256, Y: 256},
				{FileName: "1.2.tmj", X: 512, Y: 256},
			},
		},
	}

	dx, dy, ok := ScrollDeltaBetween(world, "1.1.tmj", "1.2.tmj")
	if !ok || dx != 256 || dy != 0 {
		t.Fatalf("scroll delta = (%d,%d), want (256,0)", dx, dy)
	}
}
