package api

import (
	"github.com/deploifai/cli-go/api/generated"
	"github.com/deploifai/sdk-go/api"
)

// API is a type alias using the generated client
type API = api.API[generated.GQLClient]

func NewAPI(endpoint string, authToken string) API {

	return api.NewGenericAPI[generated.GQLClient](generated.NewClient, endpoint, authToken)

}
