package main

import (
	"fmt"
	"time"

	"github/GAtom22/missedblocks/client"
	"github/GAtom22/missedblocks/config"
	"github/GAtom22/missedblocks/reporter"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/spf13/cobra"
)

func Execute(configPath string) {
	appConfig, err := config.LoadConfig(configPath)
	if err != nil {
		config.GetDefaultLogger().Fatal().Err(err).Msg("Could not load config")
	}

	appConfig.Validate()        // will exit if not valid
	appConfig.SetBechPrefixes() // will exit if not valid
	config.SetSdkConfigPrefixes(appConfig)

	log := config.GetLogger(appConfig.LogConfig)

	if len(appConfig.IncludeValidators) == 0 && len(appConfig.ExcludeValidators) == 0 {
		log.Info().Msg("Monitoring all validators")
	} else if len(appConfig.IncludeValidators) != 0 {
		log.Info().
			Strs("validators", appConfig.IncludeValidators).
			Msg("Monitoring specific validators")
	} else {
		log.Info().
			Strs("validators", appConfig.ExcludeValidators).
			Msg("Monitoring all validators except specific")
	}

	encCfg := simapp.MakeTestEncodingConfig()
	interfaceRegistry := encCfg.InterfaceRegistry

	http := client.NewTendermintHTTP(appConfig.NodeConfig, log)
	grpc := client.NewTendermintGRPC(appConfig.NodeConfig, interfaceRegistry, appConfig.QueryEachSigningInfo, log)
	slashingParams := grpc.GetSlashingParams()

	params := config.Params{
		AvgBlockTime:       http.GetAvgBlockTime(),
		SignedBlocksWindow: slashingParams.SignedBlocksWindow,
		MissedBlocksToJail: slashingParams.MissedBlocksToJail,
	}

	log.Info().
		Int64("missedBlocksToJail", params.MissedBlocksToJail).
		Float64("avgBlockTime", params.AvgBlockTime).
		Msg("Chain params calculated")

	appConfig.SetDefaultMissedBlocksGroups(params)
	if err := appConfig.MissedBlocksGroups.Validate(params.SignedBlocksWindow); err != nil {
		log.Fatal().Err(err).Msg("MissedBlockGroups config is invalid")
	}

	log.Info().
		Str("config", fmt.Sprintf("%+v", appConfig)).
		Msg("Started with following parameters")

	reporters := []reporter.Reporter{
		reporter.NewTelegramReporter(appConfig.ChainInfoConfig, appConfig.TelegramConfig, appConfig, &params, grpc, log),
		reporter.NewSlackReporter(appConfig.ChainInfoConfig, appConfig.SlackConfig, &params, log),
	}

	for _, reporter := range reporters {
		reporter.Init()

		if reporter.Enabled() {
			log.Info().Str("name", reporter.Name()).Msg("Init reporter")
		}
	}

	reportGenerator := reporter.NewReportGenerator(params, grpc, appConfig, log, interfaceRegistry)

	for {
		report := reportGenerator.GenerateReport()
		if report == nil || len(report.Entries) == 0 {
			log.Info().Msg("Report is empty, not sending.")
			time.Sleep(time.Duration(appConfig.Interval) * time.Second)
			continue
		}

		for _, reporter := range reporters {
			if !reporter.Enabled() {
				log.Debug().Str("name", reporter.Name()).Msg("Reporter is disabled.")
				continue
			}

			log.Info().Str("name", reporter.Name()).Msg("Sending a report to reporter...")
			if err := reporter.SendReport(*report); err != nil {
				log.Error().Err(err).Str("name", reporter.Name()).Msg("Could not send message")
			}
		}

		// if no reporters but metrics enabled, update the metrics
		if len(reporters) == 0 {
			// TODO update metrics with report data
		}

		time.Sleep(time.Duration(appConfig.Interval) * time.Second)
	}
}

func main() {
	var ConfigPath string

	rootCmd := &cobra.Command{
		Use:  "tendermint-exporter",
		Long: "Scrape the data on Tendermint node.",
		Run: func(cmd *cobra.Command, args []string) {
			Execute(ConfigPath)
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
