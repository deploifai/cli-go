package project_config

type Dataset struct {
	ID             string `toml:"id"`
	LocalDirectory string `toml:"localDirectory"`
}

type Datasets map[string]Dataset
