package k8sgpt

import (
	"fmt"
	"os/exec"
	"strings"

	. "github.com/pjberry16/spin-kube/internal/util"
)

func Analyze(kind, name, namespace string) error {
	output, err := run(namespace, "analyze", "--selector", fmt.Sprintf("app.kubernetes.io/name=%s", name), "--namespace", namespace, "--filter", kind, "-e")
	if err != nil {
		fmt.Println(Red("k8sgpt analyze failed:"))
		fmt.Println(Red(err.Error()))
		return err
	}

	if len(output) == 1 && strings.Contains(output[0], "No problems detected") && kind == "Deployment" {
		fmt.Println(Green("k8sgpt found no problems with Deployment. Checking related Pods...\n"))
		output, err = run(namespace, "analyze", "--selector", fmt.Sprintf("app.kubernetes.io/name=%s", name), "--namespace", namespace, "--filter", "Pod", "-e")
	}

	fmt.Println(Cyan("k8sgpt output:"))
	colorPrintOutput(output)
	fmt.Println()

	return nil
}

// run runs the k8sgpt command with the provided arguments and prints the output.
func run(namespace string, args ...string) ([]string, error) {
	cmd := exec.Command("k8sgpt", args...)
	out, err := cmd.CombinedOutput()

	outputs := formatOutput(string(out), namespace)

	return outputs, err
}

// formatK8sGPTOutput removes irrelevant parts from the k8sgpt output and splits the relevant parts by line.
func formatOutput(output, namespace string) []string {
	var relevantParts []string

	parts := strings.SplitN(output, "\n", -1)

	for _, part := range parts {
		strings.TrimSpace(part)
		if strings.Contains(part, "AI Provider:") || (strings.Contains(part, "100%") && strings.Contains(part, "it/s")) {
			continue
		} else if len(part) == 0 {
			continue
		}

		relevantParts = append(relevantParts, part)

		// k8sgpt sometimes outputs kubectl commands without the namespace flag
		if strings.Contains(part, "kubectl") {
			relevantParts = append(relevantParts, fmt.Sprintf("Info: if running a kubectl command, ensure to include the `-n %s` flag to target the correct namespace.", namespace))
		}
	}

	return relevantParts
}

func colorPrintOutput(output []string) {
	for _, line := range output {
		if strings.Contains(line, "No problems detected") {
			fmt.Println(Green(line))
		} else if strings.Contains(line, "Error:") {
			fmt.Println(Red(line))
		} else if strings.Contains(line, "Solution:") {
			fmt.Println(Green(line))
		} else if strings.Contains(line, "Info:") {
			fmt.Println(Yellow(line))
		} else {
			fmt.Println(line)
		}
	}
}
