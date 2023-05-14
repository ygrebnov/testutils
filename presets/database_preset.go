package presets

type databasePreset struct {
	defaultPreset `yaml:",inline"`
	Database      presetDatabase `yaml:"database"`
}

// presetDatabase holds databasePreset database data.
type presetDatabase struct {
	Name string `yaml:"name"`
}
