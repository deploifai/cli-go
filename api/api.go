package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/Yamashou/gqlgenc/clientv2"
	"github.com/deploifai/cli-go/api/generated"
	"net/http"
)

type API struct {
	Endpoint string

	AuthToken string

	Client generated.GQLClient
}

func New(endpoint string, authToken string) *API {

	return &API{
		Endpoint:  endpoint,
		AuthToken: authToken,
		Client: generated.NewClient(http.DefaultClient, endpoint, nil,
			func(ctx context.Context, req *http.Request, gqlInfo *clientv2.GQLRequestInfo, res interface{}, next clientv2.RequestInterceptorFunc) error {
				req.Header.Set("Authorization", authToken)
				//fmt.Println("intercepted request", req)
				//fmt.Println("intercepted res", res)

				return next(ctx, req, gqlInfo, res)
			},
		),
	}
}

func (api *API) GetClient() generated.GQLClient {
	return api.Client
}

func (api *API) ProcessError(err error) error {
	if handledError, ok := err.(*clientv2.ErrorResponse); ok {
		msg := "handled error: "
		if handledError.NetworkError != nil {
			msg = msg + fmt.Sprintf("network error: [status code = %d] %s\n", handledError.NetworkError.Code, handledError.NetworkError.Message)
		} else {
			msg = msg + fmt.Sprintf("graphql error: %v\n", handledError.GqlErrors)
		}
		return errors.New(msg)
	} else {
		return errors.New(fmt.Sprintf("unhandled error: %s\n", err.Error()))
	}

	return nil
}
