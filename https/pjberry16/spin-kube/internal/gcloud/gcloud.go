package gcloud

import (
	"errors"
	"os/exec"
	"regexp"
)

// SetContext sets the gcloud context using the provided Spinnaker account.
// gcloud container clusters get-credentials {cluster} --region {region} --project {project}
func SetContext(account string) error {
	// example Spinnaker account:
	// gke_np-te-cd-tools_us-central1_dev-us-central1-np
	re := regexp.MustCompile(`^gke_([^_]+)_([^_]+)_(.+)-[^-]+$`)
	matches := re.FindStringSubmatch(account)
	if len(matches) == 0 {
		return errors.New("account string does not match expected format")
	}
	project := matches[1]
	region := matches[2]
	cluster := matches[3]

	cmd := exec.Command("gcloud", "container", "clusters", "get-credentials", cluster, "--region", region, "--project", project)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	return nil
}
