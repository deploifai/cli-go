package create

import "github.com/deploifai/cli-go/api/generated"

func processAWS() (generated.AWSCredentials, error) {

	return generated.AWSCredentials{
		AwsAccessKey:       "",
		AwsSecretAccessKey: "",
	}, nil
}
