package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	app "github.com/maany-xyz/maany-provider/app"
	appconfig "github.com/maany-xyz/maany-provider/app/config"
	"github.com/maany-xyz/maany-provider/cmd/maanypd/cmd"
)

func main() {
	cfg := appconfig.GetDefaultConfig()
	cfg.Seal()  
	rootCmd := cmd.NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
