package command_config

type Auth struct {
	Username string `toml:"username"`
	Token    string `toml:"token"`
}
