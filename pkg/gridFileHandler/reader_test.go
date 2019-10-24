package gridFileHandler;

import (
	"testing"
)

type GridTest struct {
	description string
	in string
	wantGrid *GridFile
	wantError error
}

func TestGetWidthHeightAndTile(t *testing.T) {
	gf, err := ReadGridFile("../../dat/worlds/isleOfStarting_dev/grids/1.2.grid");

	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if gf.Width != 16 || gf.Height != 16 {
		t.Errorf("Error: Grid width and height are not 16")
	}

	for _, tile := range gf.Tiles {
		if (tile.X == 1 && tile.Y == 10) {
			if !tile.FlipV {
				t.Errorf("Expected a flipped tile: %v", tile);
			}
		}
	}
}

func TestReadGridFile(t *testing.T) {
	_, readError := ReadGridFile("404.yml")

	if readError == nil {
		t.Errorf("Should not be able to find this grid");
	}
}
