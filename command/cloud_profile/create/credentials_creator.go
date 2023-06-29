package create

import (
	"context"
	"fmt"
	"github.com/deploifai/cli-go/api/generated"
	"github.com/manifoldco/promptui"
)

type CredentialsCreator interface {
	getProfiles() ([]interface{}, error)
	mapProfiles([]interface{}) []string
	getPromptLabel() string
	createCredentials(profile interface{}, name string) (interface{}, error)
}

type CredentialsCreatorWrapper struct {
	credentialsCreator CredentialsCreator
	provider           generated.CloudProvider
}

func NewCredentialsCreatorWrapper(ctx context.Context, provider generated.CloudProvider) (wrapper CredentialsCreatorWrapper) {
	switch provider {
	case generated.CloudProviderAws:
		creator := NewAWSCredentialsCreator(ctx)
		wrapper.credentialsCreator = &creator
	case generated.CloudProviderAzure:
		creator := NewAzureCredentialsCreator(ctx)
		wrapper.credentialsCreator = &creator

	case generated.CloudProviderGcp:
		creator := NewGCPCredentialsCreator(ctx)
		wrapper.credentialsCreator = &creator
	}

	wrapper.provider = provider

	return wrapper
}

func (r *CredentialsCreatorWrapper) populateInput(input *generated.CloudProfileCreateInput, credentials interface{}) {

	switch r.provider {
	case generated.CloudProviderAws:
		c := credentials.(generated.AWSCredentials)
		input.AwsCredentials = &c
		break
	case generated.CloudProviderAzure:
		c := credentials.(generated.AzureCredentials)
		input.AzureCredentials = &c
		break
	case generated.CloudProviderGcp:
		c := credentials.(generated.GCPCredentials)
		input.GcpCredentials = &c
		break
	}
}

func (r *CredentialsCreatorWrapper) promptProfile(profiles []interface{}) (profile interface{}, err error) {
	if len(profiles) == 0 {
		return "", fmt.Errorf("no profiles provided")
	}

	if len(profiles) == 1 {
		return profiles[0], nil
	}

	items := r.credentialsCreator.mapProfiles(profiles)

	prompt := promptui.Select{
		Label: r.credentialsCreator.getPromptLabel(),
		Items: items,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return profiles[i], nil
}
