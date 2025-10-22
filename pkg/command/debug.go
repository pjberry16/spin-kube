package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/pjberry16/spin-kube/internal/gcloud"
	"github.com/pjberry16/spin-kube/internal/k8sgpt"
	"github.com/pjberry16/spin-kube/internal/spinnaker"
	. "github.com/pjberry16/spin-kube/internal/util"

	"github.com/spinnaker/spin/config"
	"github.com/spinnaker/spin/config/auth"
	spinX509 "github.com/spinnaker/spin/config/auth/x509"

	"github.com/urfave/cli/v3"
)

const (
	executionIdFlag = "execution-id"
	stageFlag       = "stage"
)

var (
	gateEndpoint = os.Getenv("GATE_ENDPOINT")
)

var (
	debugCmdName    = "debug"
	debugCmdAliases = []string{"d"}
	debugCmdUsage   = "todo"
	debugCmdFlags   = []cli.Flag{
		&cli.StringFlag{
			Name:     executionIdFlag,
			Aliases:  []string{"e"},
			Usage:    "Spinnaker execution-id",
			Required: true,
		},
		&cli.StringFlag{
			Name:     stageFlag,
			Aliases:  []string{"s"},
			Usage:    "pipeline stage name",
			Required: false,
		},
	}

	failStatuses = map[string]struct{}{
		"TERMINAL": {},
		"STOPPED":  {},
	}
)

func NewDebugCommand() *cli.Command {
	return &cli.Command{
		Name:    debugCmdName,
		Aliases: debugCmdAliases,
		Usage:   debugCmdUsage,
		Flags:   debugCmdFlags,
		Action:  DebugCmdAction,
	}
}

func DebugCmdAction(ctx context.Context, c *cli.Command) error {
	if c.NArg() > 2 {
		fmt.Println("Too many arguments provided to debug command.")
		fmt.Println()

		// todo: fix missing help for each command
		return cli.ShowCommandHelp(ctx, c, c.Name)
	}

	executionId := c.String(executionIdFlag)

	fmt.Println(Cyan("Debugging execution ID:", executionId))

	// todo: create a spin client from set cert and key
	// pull execution from spinnaker and find the Deploy (Manifest) stage(s)
	// find status failed
	// print message
	// once that's done work on k8s messages

	certPath, err := os.ReadFile(filepath.Join(userHomeDir(), ".spinkube", "cert"))
	if err != nil {
		return errors.New("x.509 cert file path not set. Please run the 'set-cert' command to set it")
	}

	cert, err := os.ReadFile(string(certPath))
	if err != nil {
		return fmt.Errorf("failed to read x.509 cert file: %w", err)
	}

	keyPath, err := os.ReadFile(filepath.Join(userHomeDir(), ".spinkube", "key"))
	if err != nil {
		return errors.New("x.509 key file path not set. Please run the 'set-key' command to set it")
	}

	key, err := os.ReadFile(string(keyPath))
	if err != nil {
		return fmt.Errorf("failed to read x.509 key file: %w", err)
	}

	authConfig := &auth.Config{
		Enabled: true,
		X509: &spinX509.Config{
			Cert: string(cert),
			Key:  string(key),
		},
	}
	cfg := config.Config{}

	// todo: make gate endpoint configurable with a command. current default is prod central
	cfg.Gate.Endpoint = gateEndpoint
	cfg.Auth = authConfig

	client, err := spinnaker.NewClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create spinnaker client: %w", err)
	}

	e, err := client.GetPipelineExecution(executionId)
	if err != nil {
		return fmt.Errorf("failed to get pipeline execution %s: %w", executionId, err)
	}

	fmt.Println(Yellow("Pipeline Execution Details:"))
	fmt.Printf("Execution Name: %s\n", e.Name)

	deployManifestStages, err := getDeployManifestStages(e)
	if err != nil {
		return fmt.Errorf("failed to get deploy manifest stages for execution id %s: %w", executionId, err)
	}

	failedStages, err := getFailedStages(deployManifestStages)
	if err != nil {
		return fmt.Errorf("failed to get failed stages for execution id %s: %w", executionId, err)
	}

	var stageErrors []string
	for _, stage := range failedStages {
		fmt.Println("Failed Stage Name: ", stage.Name)
		fmt.Println()
		fmt.Println(Red("Spinnaker Stage Errors: ", Red(stage.Context.Exception.Details.Errors)))

		stageErrors = append(stageErrors, stage.Context.Exception.Details.Errors...)
	}

	if len(stageErrors) == 0 {
		return errors.New("no errors found in failed stages")
	}

	fmt.Println()
	fmt.Println(Yellow("Parsing error information..."))
	fmt.Println()

	for _, errStr := range stageErrors {
		fmt.Println(Yellow("Analyzing Error Message:"), Red(errStr), "\n")

		kind, name, namespace, account, err := parseInfo(errStr)
		if err != nil {
			fmt.Printf(Red("Failed to parse error string: %s\n", errStr))
			continue
		}

		fmt.Println(Yellow("Successfully parsed the following information:"))
		fmt.Printf("Kind: %s\n", kind)
		fmt.Printf("Name: %s\n", name)
		fmt.Printf("Namespace: %s\n", namespace)
		fmt.Printf("Account: %s\n", account)
		fmt.Println()

		// Set the gcloud context to the correct account
		fmt.Println(Yellow(fmt.Sprintf("Setting gcloud context from account: %s\n", account)))

		err = gcloud.SetContext(account)
		if err != nil {
			fmt.Printf(Red(fmt.Sprintf("Failed to set gcloud context: %s\n", err)))
			fmt.Println(Red("Skipping k8sgpt analysis for this error message.\n"))
			continue
		}

		fmt.Println(Green("Successfully set gcloud context.\n"))

		fmt.Println(Yellow("Running k8sgpt for further analysis on the error message..."))
		fmt.Println()

		err = k8sgpt.Analyze(kind, name, namespace)
		if err != nil {
			fmt.Printf(Red("k8sgpt failed: %v\n", err))
		}
	}

	fmt.Println()
	fmt.Println(Green("Debugging complete."))

	return nil
}

func getDeployManifestStages(e spinnaker.Execution) ([]spinnaker.Stage, error) {
	var deployManifestStages []spinnaker.Stage

	for _, stage := range e.Stages {
		if stage.Type == "deployManifest" {
			deployManifestStages = append(deployManifestStages, stage)
		}
	}

	if len(deployManifestStages) == 0 {
		return nil, errors.New("no deploy manifest stages found in execution")
	}

	return deployManifestStages, nil
}

func getFailedStages(stages []spinnaker.Stage) ([]spinnaker.Stage, error) {
	var failedStages []spinnaker.Stage
	for _, stage := range stages {
		if _, ok := failStatuses[stage.Status]; ok {
			failedStages = append(failedStages, stage)
		}
	}

	if len(failedStages) == 0 {
		return nil, errors.New("no failed stages found")
	}

	return failedStages, nil
}

func getFailureReason(stage spinnaker.Stage) string {
	return ""
}

// parseInfo parses error strings of the format:
// 'Deployment captain-hook' in 'cd' for account gke_np-te-cd-tools_us-central1_dev-us-central1-np: ReplicaSet "captain-hook-8566d4fd78" has timed out progressing.
// and returns kind, name, namespace, account.
func parseInfo(s string) (string, string, string, string, error) {
	re := regexp.MustCompile(`(\w+)\s+([^\s]+)'\s+in\s+'([^\']+)'\s+for\s+account\s+([^\s:]+)`)
	m := re.FindStringSubmatch(s)
	if len(m) == 5 {
		return m[1], m[2], m[3], m[4], nil
	}
	return "", "", "", "", fmt.Errorf("failed to parse info from string: %s", s)
}
