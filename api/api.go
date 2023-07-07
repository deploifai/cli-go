package api

import (
	"github.com/deploifai/cli-go/api/generated"
	"github.com/deploifai/sdk-go/api"
)

// API is a type alias using the generated client
type API = api.GenericAPI[generated.GQLClient]

func NewAPI(gqlEndpoint string, restEndpoint string, authToken string) API {

	return api.NewGenericAPI[generated.GQLClient](generated.NewClient, gqlEndpoint, restEndpoint, api.RequestHeaders{api.WithAuthHeader(authToken)})

}
