package app

import (
	"fmt"
	"github/GAtom22/missedblocks/client"
	"github/GAtom22/missedblocks/config"
	"github/GAtom22/missedblocks/metrics"
	"github/GAtom22/missedblocks/reporter"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Run(configPath string) {
	appConfig, err := config.LoadConfig(configPath)
	if err != nil {
		config.GetDefaultLogger().Fatal().Err(err).Msg("Could not load config")
		panic(fmt.Sprintf("Could not load config. Provided path: %s", configPath))
	}
	go start(appConfig)

	if appConfig.Metrics.Enabled {
		http.Handle("/metrics", promhttp.Handler())
		//nolint:gosec
		if err := http.ListenAndServe(fmt.Sprintf(":%d", appConfig.Metrics.Port), nil); err != nil {
			panic(err)
		}
	}
}

func start(conf *config.AppConfig) {
	conf.Validate()        // will exit if not valid
	conf.SetBechPrefixes() // will exit if not valid
	config.SetSdkConfigPrefixes(conf)

	log := config.GetLogger(conf.LogConfig)

	switch {
	case len(conf.IncludeValidators) == 0 && len(conf.ExcludeValidators) == 0:
		log.Info().Msg("Monitoring all validators")
	case len(conf.IncludeValidators) != 0:
		log.Info().
			Strs("validators", conf.IncludeValidators).
			Msg("Monitoring specific validators")
	default:
		log.Info().
			Strs("validators", conf.ExcludeValidators).
			Msg("Monitoring all validators except specific")
	}

	encCfg := simapp.MakeTestEncodingConfig()
	interfaceRegistry := encCfg.InterfaceRegistry

	http := client.NewTendermintHTTP(conf.NodeConfig.TendermintRPC, log)
	grpc := client.NewTendermintGRPC(conf.NodeConfig, interfaceRegistry, conf.QueryEachSigningInfo, log)
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

	conf.SetDefaultMissedBlocksGroups(params)
	if err := conf.MissedBlocksGroups.Validate(params.SignedBlocksWindow); err != nil {
		log.Fatal().Err(err).Msg("MissedBlockGroups config is invalid")
	}

	log.Info().
		Str("config", fmt.Sprintf("%+v", conf)).
		Msg("Started with following parameters")

	reporters := []reporter.Reporter{
		reporter.NewTelegramReporter(conf.ChainInfoConfig, conf.TelegramConfig, conf, &params, grpc, log),
		reporter.NewSlackReporter(conf.ChainInfoConfig, conf.SlackConfig, &params, log, conf.Metrics.Enabled),
	}

	for _, reporter := range reporters {
		reporter.Init()

		if reporter.Enabled() {
			log.Info().Str("name", reporter.Name()).Msg("Init reporter")
		}
	}

	reportGenerator := reporter.NewReportGenerator(params, grpc, conf, log, interfaceRegistry)

	for {
		report := reportGenerator.GenerateReport()
		if report == nil || len(report.Entries) == 0 {
			log.Info().Msg("Report is empty, not sending.")
			time.Sleep(time.Duration(conf.Interval) * time.Second)
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
		if len(reporters) == 0 && conf.Metrics.Enabled {
			for _, e := range report.Entries {
				metrics.UpdateMissedBlocks(e.ValidatorAddress, e.MissingBlocks)
			}
		}

		time.Sleep(time.Duration(conf.Interval) * time.Second)
	}
}
