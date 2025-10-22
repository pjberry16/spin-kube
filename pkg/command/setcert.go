package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/pjberry16/spin-kube/internal/util"
	cli "github.com/urfave/cli/v3"
)

var (
	setCertCmdName    = "set-cert"
	setCertCmdAliases = []string{"sc"}
	setCertCmdUsage   = "set the X.509 cert path"
)

func NewSetCertCmd() *cli.Command {
	return &cli.Command{
		Name:               setCertCmdName,
		Aliases:            setCertCmdAliases,
		Usage:              setCertCmdUsage,
		Action:             SetCertCmdAction,
		CustomHelpTemplate: "something",
	}
}

func SetCertCmdAction(ctx context.Context, c *cli.Command) error {
	if c.NArg() > 1 {
		fmt.Println("Too many arguments provided to command.")
		fmt.Println()

		return cli.ShowCommandHelp(ctx, c, c.Name)
	} else if c.NArg() < 1 {
		fmt.Println("Not enough arguments provided to command.")
		fmt.Println()

		return cli.ShowCommandHelp(ctx, c, c.Name)
	}

	path := strings.TrimSpace(c.Args().First())
	fmt.Printf("Setting X.509 cert path to %s...\n", Cyan(path))

	err := os.WriteFile(filepath.Join(userHomeDir(), ".spinkube", "cert"), []byte(path), 0600)
	if err != nil {
		return err
	}

	fmt.Println(Green("OK"))

	return nil
}
