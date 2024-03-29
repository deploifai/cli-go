package create

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/deploifai/cli-go/utils/spinner_utils"
	"github.com/deploifai/sdk-go/api/generated"
	"github.com/deploifai/sdk-go/api/utils"
	"os/exec"
)

type AzureCredentialsCreator struct {
	ctx context.Context
}

func NewAzureCredentialsCreator(ctx context.Context) AzureCredentialsCreator {
	return AzureCredentialsCreator{ctx: ctx}
}

type account struct {
	SubscriptionId string `json:"id"`
	Name           string `json:"name"`
	User           struct {
		Name string `json:"name"`
	} `json:"user"`
}

type servicePrincipal struct {
	TenantId     string `json:"tenant"`
	ClientId     string `json:"appId"`
	ClientSecret string `json:"password"`
}

func (r *AzureCredentialsCreator) getProfiles() (profiles []interface{}, err error) {
	output, err := exec.Command("az", "account", "list", "--output", "json").Output()

	if err != nil {
		return profiles, fmt.Errorf("Azure CLI error: %s.\nPlease make sure you have Azure CLI installed and configured", err.Error())
	}

	// unmarshal json
	var accounts []*account
	err = json.Unmarshal(output, &accounts)
	if err != nil {
		return profiles, fmt.Errorf("failed to parse Azure CLI output: %s", err.Error())
	}

	for _, account := range accounts {
		profiles = append(profiles, *account)
	}

	return profiles, nil
}

func (r *AzureCredentialsCreator) mapProfiles(profiles []interface{}) []string {
	var mappedProfiles []string
	for _, profile := range profiles {
		p := profile.(account)
		mappedProfiles = append(mappedProfiles, fmt.Sprintf("%s (%s)", p.Name, p.User.Name))
	}
	return mappedProfiles
}

func (r *AzureCredentialsCreator) getPromptMessage() string {
	return "Select subscription account"
}

func (r *AzureCredentialsCreator) createCredentials(profile interface{}, name string) (interface{}, error) {

	account := profile.(account)

	spinner := spinner_utils.NewAPICallSpinner()
	spinner.Suffix = " Creating credentials... "

	spinner.Start()
	defer spinner.Stop()

	// set azure cli context to the selected account
	combinedOut, err := exec.Command("az", "account", "set", "--subscription", account.SubscriptionId).CombinedOutput()

	if err != nil {
		return generated.AzureCredentials{}, fmt.Errorf("failed to set Azure CLI context: %s, %s", err.Error(), combinedOut)
	}

	// create service principal
	output, err := exec.Command("az", "ad", "sp", "create-for-rbac", "--name", name, "--role", "Contributor", "--scopes", fmt.Sprintf("/subscriptions/%s", account.SubscriptionId)).Output()

	if err != nil {
		return generated.AzureCredentials{}, fmt.Errorf("failed to create service principal: %s", err.Error())
	}

	// unmarshal json
	sp := &servicePrincipal{}
	err = json.Unmarshal(output, &sp)
	if err != nil {
		return generated.AzureCredentials{}, fmt.Errorf("failed to parse Azure CLI output: %s", err.Error())
	}

	// attempt to use credentials
	err = r.attemptToUseCredentials(sp)
	if err != nil {
		return generated.AzureCredentials{}, err
	}

	return generated.AzureCredentials{
		AzureSubscriptionID: account.SubscriptionId,
		AzureTenantID:       sp.TenantId,
		AzureClientID:       sp.ClientId,
		AzureClientSecret:   sp.ClientSecret,
	}, nil
}

func (r *AzureCredentialsCreator) attemptToUseCredentials(sp *servicePrincipal) error {
	credential, err := azidentity.NewClientSecretCredential(sp.TenantId, sp.ClientId, sp.ClientSecret, nil)

	if err != nil {
		return fmt.Errorf("failed to create Azure client secret credential: %s", err.Error())
	}

	retryCount := 20

	_, err = utils.CallWithRetries[azcore.AccessToken](
		func() (azcore.AccessToken, error) {
			return credential.GetToken(r.ctx, policy.TokenRequestOptions{
				Scopes:   []string{"https://management.azure.com/.default"},
				TenantID: sp.TenantId,
			})
		},
		&retryCount, nil,
	)
	if err != nil {
		return fmt.Errorf("failed to use credentials after trying for %d times", retryCount)
	}

	return nil
}
