package greyvarserver

import (
	"testing"

	"github.com/greyvar/datlib/gridfiles"
	"github.com/greyvar/server/pkg/worlds"
)

func testServerWithGrid(grid *gridfiles.Grid) (*serverInterface, *RemotePlayer) {
	world := &worlds.World{
		ID:    "testWorld",
		Grids: map[string]*gridfiles.Grid{"testGrid": grid},
	}

	s := &serverInterface{
		loadedWorlds:    map[string]*worlds.World{"testWorld": world},
		entityInstances: make(map[int64]*Entity),
	}

	rp := &RemotePlayer{
		CurrentWorldId: "testWorld",
		CurrentGridId:  "testGrid",
	}

	return s, rp
}

func TestIsTileTraversable(t *testing.T) {
	blockingTileTextures = map[string]bool{
		"water.png":   true,
		"barrier.png": true,
	}

	tests := []struct {
		name string
		tile *gridfiles.Tile
		want bool
	}{
		{
			name: "nil tile",
			tile: nil,
			want: false,
		},
		{
			name: "water is blocked",
			tile: &gridfiles.Tile{Texture: "water.png"},
			want: false,
		},
		{
			name: "barrier is blocked",
			tile: &gridfiles.Tile{Texture: "barrier.png"},
			want: false,
		},
		{
			name: "sand is walkable",
			tile: &gridfiles.Tile{Texture: "sand.png"},
			want: true,
		},
		{
			name: "shoreline blend is walkable",
			tile: &gridfiles.Tile{Texture: "sandWater.png"},
			want: true,
		},
		{
			name: "sand water corner is walkable",
			tile: &gridfiles.Tile{Texture: "sandWaterConcaveCornerAndSide.png"},
			want: true,
		},
		{
			name: "explicit traversable true",
			tile: &gridfiles.Tile{Texture: "sand.png", Traversable: true},
			want: true,
		},
		{
			name: "dev wall texture without tiledef is walkable for now",
			tile: &gridfiles.Tile{Texture: "dirtHillsideDarkGrass.png", Traversable: false},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTileTraversable(tt.tile); got != tt.want {
				t.Fatalf("isTileTraversable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanMoveTo(t *testing.T) {
	blockingTileTextures = map[string]bool{
		"water.png": true,
	}

	grid := &gridfiles.Grid{
		RowCount: 2,
		ColCount: 2,
	}
	grid.Build()
	grid.Tiles[0][0] = &gridfiles.Tile{Row: 0, Col: 0, Texture: "sand.png", Traversable: true}
	grid.Tiles[0][1] = &gridfiles.Tile{Row: 0, Col: 1, Texture: "water.png"}
	grid.Tiles[1][0] = &gridfiles.Tile{Row: 1, Col: 0, Texture: "sand.png", Traversable: true}
	grid.Tiles[1][1] = &gridfiles.Tile{Row: 1, Col: 1, Texture: "sand.png", Traversable: true}

	s, rp := testServerWithGrid(grid)

	player := &Entity{
		ServerId:   1,
		Definition: "player",
		X:          0,
		Y:          0,
		WorldId:    "testWorld",
		GridId:     "testGrid",
	}
	// Current convention: x = col * 16, y = row * 16.
	// Water sits at row 0, col 1 => pixel (16, 0).
	otherPlayer := &Entity{
		ServerId:   2,
		Definition: "player",
		X:          16,
		Y:          16,
		WorldId:    "testWorld",
		GridId:     "testGrid",
	}

	s.entityInstances[player.ServerId] = player
	s.entityInstances[otherPlayer.ServerId] = otherPlayer

	if !s.canMoveTo(rp, player, 0, 16) {
		t.Fatal("expected move onto open sand tile to succeed")
	}

	if s.canMoveTo(rp, player, 16, 0) {
		t.Fatal("expected move into water tile to be blocked")
	}

	if s.canMoveTo(rp, player, 16, 16) {
		t.Fatal("expected move into another player to be blocked")
	}

	if s.canMoveTo(rp, player, 32, 0) {
		t.Fatal("expected move outside grid to be blocked")
	}

	ghostPlayer := &Entity{
		ServerId:   3,
		Definition: "player",
		X:          0,
		Y:          0,
		WorldId:    "testWorld",
		GridId:     "testGrid",
	}
	s.entityInstances[ghostPlayer.ServerId] = ghostPlayer

	if !s.canMoveTo(rp, player, 0, 4) {
		t.Fatal("expected move away from stacked ghost player to succeed")
	}
}

func TestCanMoveToFromShorelineSpawn(t *testing.T) {
	blockingTileTextures = map[string]bool{
		"water.png": true,
	}

	grid := &gridfiles.Grid{
		RowCount: 12,
		ColCount: 12,
	}
	grid.Build()
	grid.Tiles[9][8] = &gridfiles.Tile{Row: 9, Col: 8, Texture: "sandWaterConcaveCornerAndSide.png"}
	grid.Tiles[8][8] = &gridfiles.Tile{Row: 8, Col: 8, Texture: "sandWaterConvexCorner.png"}
	grid.Tiles[10][8] = &gridfiles.Tile{Row: 10, Col: 8, Texture: "water.png"}

	s, rp := testServerWithGrid(grid)

	// Player on the shoreline tile at row 9, col 8 => pixel (128, 144).
	player := &Entity{
		ServerId:   1,
		Definition: "player",
		X:          128,
		Y:          144,
		WorldId:    "testWorld",
		GridId:     "testGrid",
	}
	s.entityInstances[player.ServerId] = player

	if !s.canMoveTo(rp, player, 128, 128) {
		t.Fatal("expected move onto shoreline tile to succeed")
	}

	if s.canMoveTo(rp, player, 128, 160) {
		t.Fatal("expected move into water tile to be blocked")
	}
}
