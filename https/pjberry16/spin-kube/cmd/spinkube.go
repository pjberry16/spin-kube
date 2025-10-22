package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"
	"github.com/pjberry16/spin-kube/pkg/command"
	cli "github.com/urfave/cli/v3"
)

var (
	spinnakerURL = GetEnv("SPINNAKER_URL", "")
	app          *cli.Command
	red          = color.New(color.FgRed, color.Bold).SprintFunc()
)

func main() {
	_ = os.MkdirAll(filepath.Join(userHomeDir(), ".clipper"), os.ModePerm)

	ctx := context.Background()

	err := app.Run(ctx, os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	_ = os.MkdirAll(filepath.Join(userHomeDir(), ".spinkube"), os.ModePerm)

	app = &cli.Command{}
	app.EnableShellCompletion = true
	app.Usage = "A CLI tool for diagnosing Kubernetes deployment failures triggered by Spinnaker"
	app.ExitErrHandler = func(context context.Context, app *cli.Command, err error) {
		if err != nil {
			fmt.Println(red("FAILED"))
			fmt.Println(err.Error())
		}

		os.Exit(1)
	}

	app.Commands = []*cli.Command{
		command.NewDebugCommand(),
		command.NewSetCertCmd(),
		command.NewSetKeyCmd(),
	}
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}

		return home
	} else if runtime.GOOS == "linux" {
		home := os.Getenv("XDG_CONFIG_HOME")
		if home != "" {
			return home
		}
	}

	return os.Getenv("HOME")
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
