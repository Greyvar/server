package gridFileHandler

type Tile struct {
	Row uint32
	Col uint32
	Rot int32
	FlipH bool `yaml:"flipH"`
	FlipV bool `yaml:"flipV"`
	Traversable bool
	Texture string
}

type GridFileEntityInstance struct {
	Row int32
	Col int32
	Definition string
	GridID string `yaml:"id"` // Ignored from map file for now. Overwritten by server. Will need to change this.

	Spawned bool `yaml:"-"`
	State string `yaml:"-"`
}

type GridFile struct {
	Filename string
	ColCount uint32
	RowCount uint32
	Tiles []Tile
	Entities []GridFileEntityInstance
	LastEntityId string `yaml:"lastEntityId"`
}
