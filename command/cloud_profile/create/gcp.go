package create

import "github.com/deploifai/cli-go/api/generated"

func processGCP() (generated.GCPCredentials, error) {

	return generated.GCPCredentials{
		GcpProjectID:         "",
		GcpServiceAccountKey: "",
	}, nil
}
