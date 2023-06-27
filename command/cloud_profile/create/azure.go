package create

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/deploifai/cli-go/api/generated"
	"github.com/deploifai/cli-go/utils/spinner_utils"
	"os/exec"
	"time"
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
		return profiles, fmt.Errorf("Azure CLI error: %s. Please make sure you have Azure CLI installed and configured", err.Error())
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

func (r *AzureCredentialsCreator) getPromptLabel() string {
	return "Select account"
}

func (r *AzureCredentialsCreator) createCredentials(profile interface{}, name string) (interface{}, error) {

	account := profile.(account)

	// set azure cli context to the selected account
	err := exec.Command("az", "account", "set", "--subscription", account.SubscriptionId).Run()

	if err != nil {
		return generated.AzureCredentials{}, fmt.Errorf("failed to set Azure CLI context: %s", err.Error())
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
	spinner := spinner_utils.NewSleepSpinner()
	spinner.Suffix = " Attempting to use service principal credentials... "

	currentSleepDuration := 3 * time.Second
	totalSleepDuration := 0 * time.Second

	credential, err := azidentity.NewClientSecretCredential(sp.TenantId, sp.ClientId, sp.ClientSecret, nil)

	if err != nil {
		return fmt.Errorf("failed to create Azure client secret credential: %s", err.Error())
	}

	spinner.Start()
	defer spinner.Stop()

	for true {
		if totalSleepDuration > 60*time.Second {
			return fmt.Errorf("failed to use credentials after trying for %s", totalSleepDuration.String())
		}
		_, err = credential.GetToken(r.ctx, policy.TokenRequestOptions{
			Scopes:   []string{"https://management.azure.com/.default"},
			TenantID: sp.TenantId,
		})
		if err == nil {
			break
		} else {
			time.Sleep(currentSleepDuration)
			totalSleepDuration += currentSleepDuration
		}
	}

	return err
}
