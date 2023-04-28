package command_config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Auth Auth `toml:"auth"`

	//Workspace struct {
	//	Workspace `mapstructure:",squash"`
	//} `toml:"workspace"`

	Workspace Workspace `toml:"workspace"`
}

func SetDefaultConfig(v *viper.Viper) {
	v.SetDefault("auth", map[string]interface{}{
		"username": "",
		"token":    "",
	})
	v.SetDefault("workspace", map[string]interface{}{
		"username": "",
	})
}

func (c *Config) WriteStructIntoConfig(v *viper.Viper) {
	v.Set("auth", c.Auth)
	v.Set("workspace", c.Workspace)
}
