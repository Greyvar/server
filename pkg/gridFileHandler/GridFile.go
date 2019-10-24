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

type Entity struct {
	X uint32
	Y uint32
	Id string
	Definition string
}

type GridFile struct {
	Width int
	Height int
	Tiles []Tile
	Entities []Entity
	LastEntityId string `yaml:"lastEntityId"`
}
