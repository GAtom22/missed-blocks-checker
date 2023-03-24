package cli

import (
	"github/GAtom22/missedblocks/app"
	"github/GAtom22/missedblocks/config"

	"github.com/spf13/cobra"
)

func Run() {
	var ConfigPath string
	rootCmd := &cobra.Command{
		Use:  "tendermint-exporter",
		Long: "Scrape the data on Tendermint node.",
		Run: func(cmd *cobra.Command, args []string) {
			app.Run(ConfigPath)
		},
	}

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Config file path")
	if err := rootCmd.MarkPersistentFlagRequired("config"); err != nil {
		config.GetDefaultLogger().Fatal().Err(err).Msg("Could not set flags")
	}

	if err := rootCmd.Execute(); err != nil {
		config.GetDefaultLogger().Fatal().Err(err).Msg("Could not start application")
	}
}
