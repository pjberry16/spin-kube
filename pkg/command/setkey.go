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
	setKeyCmdName    = "set-key"
	setKeyCmdAliases = []string{"sk"}
	setKeyCmdUsage   = "set the X.509 key path"
)

func NewSetKeyCmd() *cli.Command {
	return &cli.Command{
		Name:    setKeyCmdName,
		Aliases: setKeyCmdAliases,
		Usage:   setKeyCmdUsage,
		Action:  SetKeyCmdAction,
	}
}

func SetKeyCmdAction(ctx context.Context, c *cli.Command) error {
	if c.NArg() > 1 {
		fmt.Println("Too many arguments provided to command.")
		fmt.Println()

		return cli.ShowCommandHelp(ctx, c, c.Name)
	} else if c.NArg() < 1 {
		fmt.Println("Not enough arguments provided to command.")
		fmt.Println()

		commands := c.Commands
		for _, cmd := range commands {
			fmt.Println(cmd.Name)
		}

		return cli.ShowCommandHelp(ctx, c, c.Name)
	}

	path := strings.TrimSpace(c.Args().First())
	fmt.Printf("Setting X.509 key path to %s...\n", Cyan(path))

	err := os.WriteFile(filepath.Join(userHomeDir(), ".spinkube", "key"), []byte(path), 0600)
	if err != nil {
		return err
	}

	fmt.Println(Green("OK"))

	return nil
}
