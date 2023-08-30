package root_config

type Auth struct {
	Username string `toml:"username"`
	Token    string `toml:"token"`
}
