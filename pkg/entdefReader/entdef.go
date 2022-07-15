package entdefReader

type EntityDefinition struct {
	Title string
	InitialState string `yaml:"initialState"`
	States map[string]EntityState
}

type EntityState struct {
	Tex string
	Frames int
}
