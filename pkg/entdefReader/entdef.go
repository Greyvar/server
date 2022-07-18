package entdefReader

type EntityDefinition struct {
	Title string
	InitialState string `yaml:"initialState"`
	States map[string]EntityState
	Texture string
}

type EntityState struct {
	Name string
	Frames []int32
}
