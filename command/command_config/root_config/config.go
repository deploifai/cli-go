package root_config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Auth Auth `toml:"auth"`

	Workspace Workspace `toml:"workspace"`
}

func SetDefaultConfig(v *viper.Viper) {
	v.SetDefault("auth", Auth{
		Username: "",
		Token:    "",
	})
	v.SetDefault("workspace", Workspace{
		Username: "",
	})
}

func (c *Config) WriteStructIntoViper(v *viper.Viper) {
	v.Set("auth", c.Auth)
	v.Set("workspace", c.Workspace)
}
