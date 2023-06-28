package create

import (
	"cloud.google.com/go/iam/apiv1/iampb"
	"context"
	"encoding/json"
	"fmt"
	"github.com/deploifai/cli-go/api/generated"
	"github.com/deploifai/cli-go/utils/spinner_utils"
	"os"
	"os/exec"
)

type GCPCredentialsCreator struct {
	ctx context.Context
}

func NewGCPCredentialsCreator(ctx context.Context) GCPCredentialsCreator {
	return GCPCredentialsCreator{ctx: ctx}
}

type project struct {
	Name           string `json:"name"`
	ProjectId      string `json:"projectId"`
	LifecycleState string `json:"lifecycleState"`
}

type serviceAccount struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

type simplePolicy struct {
	Bindings []*iampb.Binding `json:"bindings"`
	Version  int32            `json:"version"`
	Etag     string           `json:"etag"`
}

func (r *GCPCredentialsCreator) getProfiles() ([]interface{}, error) {

	combinedOut, err := exec.Command("gcloud", "projects", "list", "--format=json").CombinedOutput()

	if err != nil {
		return []interface{}{}, fmt.Errorf("GCP CLI error: %s, %s.\nPlease make sure you have GCP CLI installed and configured", err.Error(), combinedOut)
	}

	// unmarshal json
	var projects []*project
	if err = json.Unmarshal(combinedOut, &projects); err != nil {
		return []interface{}{}, fmt.Errorf("failed to parse GCP CLI output: %s", err.Error())
	}

	// filter out projects with lifecycleState != ACTIVE
	// convert to interface{}
	var profiles []interface{}
	for _, project := range projects {
		if project.LifecycleState == "ACTIVE" {
			profiles = append(profiles, *project)
		}
	}

	if len(profiles) == 0 {
		return []interface{}{}, fmt.Errorf("no active projects found. Are you logged in? Try running `gcloud auth login` or `gcloud auth list` to check your accounts")
	}

	return profiles, nil
}

func (r *GCPCredentialsCreator) mapProfiles(profiles []interface{}) []string {
	var mappedProfiles []string
	for _, profile := range profiles {
		p := profile.(project)
		mappedProfiles = append(mappedProfiles, fmt.Sprintf("%s (%s)", p.Name, p.ProjectId))
	}
	return mappedProfiles
}

func (r *GCPCredentialsCreator) getPromptLabel() string {
	return "Select project"
}

func (r *GCPCredentialsCreator) createCredentials(profile interface{}, name string) (interface{}, error) {
	project := profile.(project)

	spinner := spinner_utils.NewAPICallSpinner()
	spinner.Suffix = " Creating credentials... "

	spinner.Start()
	defer spinner.Stop()

	// set project
	combinedOut, err := exec.Command("gcloud", "config", "set", "project", project.ProjectId).CombinedOutput()
	if err != nil {
		return generated.GCPCredentials{}, fmt.Errorf("failed to set project: %s, %s", err.Error(), combinedOut)
	}

	// enable services
	combinedOut, err = exec.Command("gcloud", "services", "enable",
		"artifactregistry.googleapis.com",
		"compute.googleapis.com",
		"iam.googleapis.com",
		"iamcredentials.googleapis.com",
		"storage-api.googleapis.com").CombinedOutput()
	if err != nil {
		return generated.GCPCredentials{}, fmt.Errorf("failed to enable services: %s, %s", err.Error(), combinedOut)
	}

	combinedOut, err = exec.Command("gcloud", "iam", "service-accounts", "create", name, "--display-name", name, "--description", "Service account for Deploifai").CombinedOutput()
	if err != nil {
		return generated.GCPCredentials{}, fmt.Errorf("failed to create service account: %s, %s", err.Error(), combinedOut)
	}

	// get service account
	combinedOut, err = exec.Command("gcloud", "iam", "service-accounts", "list", "--format=json").CombinedOutput()
	if err != nil {
		return generated.GCPCredentials{}, fmt.Errorf("failed to get service account: %s", err.Error())
	}

	// unmarshal json
	var serviceAccounts []*serviceAccount
	if err = json.Unmarshal(combinedOut, &serviceAccounts); err != nil {
		return generated.GCPCredentials{}, fmt.Errorf("failed to parse GCP CLI output: %s", err.Error())
	}
	var serviceAccount serviceAccount
	for _, s := range serviceAccounts {
		if s.DisplayName == name {
			serviceAccount = *s
			break
		}
	}

	// get iam policy
	combinedOut, err = exec.Command("gcloud", "projects", "get-iam-policy", project.ProjectId, "--format=json").CombinedOutput()
	if err != nil {
		return generated.GCPCredentials{}, fmt.Errorf("failed to get iam policy: %s, %s", err.Error(), combinedOut)
	}

	// unmarshal json
	currentPolicy := &simplePolicy{}
	if err = json.Unmarshal(combinedOut, &currentPolicy); err != nil {
		return generated.GCPCredentials{}, fmt.Errorf("failed to parse GCP CLI output: %s", err.Error())
	}

	// create new policy
	roles := []string{
		"roles/editor", "roles/resourcemanager.projectIamAdmin", "roles/storage.admin", "roles/run.admin",
	}
	for _, role := range roles {
		binding := iampb.Binding{
			Role:    role,
			Members: []string{fmt.Sprintf("serviceAccount:%s", serviceAccount.Email)},
		}
		currentPolicy.Bindings = append(currentPolicy.Bindings, &binding)
	}

	// write policy to json file
	policyFilename := ".policy.json"
	newPolicyContent, err := json.Marshal(currentPolicy)
	if err != nil {
		return generated.GCPCredentials{}, fmt.Errorf("failed to marshal policy: %s", err.Error())
	}
	if err = os.WriteFile(policyFilename, newPolicyContent, 0644); err != nil {
		return generated.GCPCredentials{}, fmt.Errorf("failed to write policy file: %s", err.Error())
	}

	// add iam policy binding
	combinedOut, err = exec.Command("gcloud", "projects", "set-iam-policy", project.ProjectId, policyFilename).CombinedOutput()
	if err != nil {
		return generated.GCPCredentials{}, fmt.Errorf("failed to add iam policy binding: %s, %s", err.Error(), combinedOut)
	}

	// delete policy file
	if err = os.Remove(policyFilename); err != nil {
		return generated.GCPCredentials{}, fmt.Errorf("failed to delete policy file: %s", err.Error())
	}

	// create key file
	keyFilename := ".key.json"
	combinedOut, err = exec.Command("gcloud", "iam", "service-accounts", "keys", "create", keyFilename, fmt.Sprintf("--iam-account=%s", serviceAccount.Email), "--key-file-type=json").CombinedOutput()
	if err != nil {
		return generated.GCPCredentials{}, fmt.Errorf("failed to create key file: %s, %s", err.Error(), combinedOut)
	}

	// read key file
	fileContent, err := os.ReadFile(keyFilename)
	if err != nil {
		return generated.GCPCredentials{}, fmt.Errorf("failed to read key file: %s", err.Error())
	}

	// delete key file
	if err = os.Remove(keyFilename); err != nil {
		return generated.GCPCredentials{}, fmt.Errorf("failed to delete key file: %s", err.Error())
	}

	return generated.GCPCredentials{
		GcpProjectID:         project.ProjectId,
		GcpServiceAccountKey: string(fileContent),
	}, nil
}
