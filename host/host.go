package host

type endpoint struct {
	Base string

	GraphQL string

	Auth auth
}

type auth struct {
	Login string
}

const base = "https://api.deploif.ai"

var Endpoint = endpoint{
	Base: base,

	GraphQL: base + "/graphql",

	Auth: auth{
		Login: base + "/auth/login/cli",
	},
}
