package gridFileHandler

type Tile struct {
	X uint32
	Y uint32
	Rot int32
	FlipH bool `yaml:"flipH"`
	FlipV bool `yaml:"flipV"`
	Traversable bool
	Texture string
}

type GridFileEntityInstance struct {
	X int32
	Y int32
	Definition string
	GridID string `yaml:"id"` // Ignored from map file for now. Overwritten by server. Will need to change this.

	Spawned bool `yaml:"-"`
	State string `yaml:"-"`
}

type GridFile struct {
	Width int
	Height int
	Tiles []Tile
	Entities []GridFileEntityInstance
	LastEntityId string `yaml:"lastEntityId"`
}
