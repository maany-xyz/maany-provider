package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	app "github.com/maany-xyz/maany-provider/app"
	"github.com/maany-xyz/maany-provider/cmd/maanypd/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
