package project_config

import "github.com/spf13/viper"

const ConfigFilename = "deploifai.toml"

type Config struct {
	Project Project `toml:"project"`

	Datasets Datasets `toml:"datasets"`

	ConfigFile string
}

func SetDefaultConfig(v *viper.Viper) {
	v.SetDefault("project", Project{
		ID: "",
	})
	v.SetDefault("datasets", Datasets{})
}

func (c *Config) SetConfigFile(f string) {
	c.ConfigFile = f
}

func (c *Config) WriteStructIntoViper(v *viper.Viper) {
	v.Set("project", c.Project)
	v.Set("datasets", c.Datasets)
}
