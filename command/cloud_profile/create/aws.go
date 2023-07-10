package create

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/deploifai/cli-go/utils/spinner_utils"
	"github.com/deploifai/sdk-go/api/generated"
	"os/exec"
	"strings"
)

type AWSCredentialsCreator struct {
	ctx context.Context
}

func NewAWSCredentialsCreator(ctx context.Context) AWSCredentialsCreator {
	return AWSCredentialsCreator{
		ctx: ctx,
	}
}

func (r *AWSCredentialsCreator) getProfiles() ([]interface{}, error) {
	combinedOut, err := exec.Command("aws", "configure", "list-profiles").CombinedOutput()

	if err != nil {
		return []interface{}{}, fmt.Errorf("AWS CLI error: %s, %s.\nPlease make sure you have AWS CLI installed and configured", err.Error(), combinedOut)
	}

	// split by \n, remove \r, remove empty strings
	splitStrings := strings.Split(strings.ReplaceAll(string(combinedOut), "\r\n", "\n"), "\n")
	var profiles []interface{}
	for _, s := range splitStrings {
		if s != "" {
			profiles = append(profiles, s)
		}
	}

	return profiles, nil
}

func (r *AWSCredentialsCreator) mapProfiles(profiles []interface{}) []string {
	var mappedProfiles []string
	for _, profile := range profiles {
		mappedProfiles = append(mappedProfiles, profile.(string))
	}
	return mappedProfiles
}

func (r *AWSCredentialsCreator) getPromptMessage() string {
	return "Select profile"
}

func (r *AWSCredentialsCreator) createCredentials(profile interface{}, name string) (interface{}, error) {
	spinner := spinner_utils.NewAPICallSpinner()
	spinner.Suffix = " Creating credentials... "

	spinner.Start()
	defer spinner.Stop()

	cfg, err := config.LoadDefaultConfig(r.ctx, config.WithSharedConfigProfile(profile.(string)))

	if err != nil {
		return generated.AWSCredentials{}, err
	}

	client := iam.NewFromConfig(cfg)

	// check credentials - make an api call
	_, err = client.ListUsers(r.ctx, &iam.ListUsersInput{})
	if err != nil {
		return generated.AWSCredentials{}, fmt.Errorf("AWS credentials failed: %s", err.Error())
	}

	// create iam user
	_, err = client.CreateUser(r.ctx, &iam.CreateUserInput{
		UserName: &name,
	})
	if err != nil {
		return generated.AWSCredentials{}, fmt.Errorf("failed to create IAM user: %s", err.Error())
	}

	// attach user policy
	policyArns := []string{"arn:aws:iam::aws:policy/PowerUserAccess", "arn:aws:iam::aws:policy/IAMFullAccess"}
	for _, policyArn := range policyArns {
		_, err = client.AttachUserPolicy(r.ctx, &iam.AttachUserPolicyInput{
			UserName:  &name,
			PolicyArn: &policyArn,
		})
	}

	// create access key
	output, err := client.CreateAccessKey(r.ctx, &iam.CreateAccessKeyInput{
		UserName: &name,
	})

	return generated.AWSCredentials{
		AwsAccessKey:       *output.AccessKey.AccessKeyId,
		AwsSecretAccessKey: *output.AccessKey.SecretAccessKey,
	}, nil
}
