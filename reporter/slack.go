package reporter

import (
	"fmt"
	"github/GAtom22/missedblocks/config"
	"strings"

	"github.com/rs/zerolog"
	"github.com/slack-go/slack"
)

type SlackReporter struct {
	ChainInfoConfig config.ChainInfoConfig
	SlackConfig     config.SlackConfig
	Params          *config.Params
	Logger          zerolog.Logger

	SlackClient slack.Client
}

func NewSlackReporter(
	chainInfoConfig config.ChainInfoConfig,
	slackConfig config.SlackConfig,
	params *config.Params,
	logger *zerolog.Logger,
) *SlackReporter {
	return &SlackReporter{
		ChainInfoConfig: chainInfoConfig,
		SlackConfig:     slackConfig,
		Params:          params,
		Logger:          logger.With().Str("component", "slack_reporter").Logger(),
	}
}

func (r SlackReporter) Serialize(report Report) string {
	var sb strings.Builder

	for _, entry := range report.Entries {
		var (
			validatorLink string
			timeToJail    = ""
		)

		if entry.Direction == INCREASING {
			timeToJail = fmt.Sprintf(" (%s till jail)", entry.GetTimeToJail(r.Params))
		}

		validatorLink = r.ChainInfoConfig.GetValidatorPage(entry.ValidatorAddress, entry.ValidatorMoniker)
		sb.WriteString(fmt.Sprintf(
			"%s <strong>%s %s</strong>%s\n",
			entry.Emoji,
			validatorLink,
			entry.Description,
			timeToJail,
		))
	}

	return sb.String()
}

func (r *SlackReporter) Init() {
	if r.SlackConfig.Token == "" || r.SlackConfig.Chat == "" {
		r.Logger.Debug().Msg("Slack credentials not set, not creating Slack reporter.")
		return
	}

	client := slack.New(r.SlackConfig.Token)
	r.SlackClient = *client
}

func (r SlackReporter) Enabled() bool {
	return r.SlackConfig.Token != "" && r.SlackConfig.Chat != ""
}

func (r SlackReporter) SendReport(report Report) error {
	serializedReport := r.Serialize(report)
	_, _, err := r.SlackClient.PostMessage(
		r.SlackConfig.Chat,
		slack.MsgOptionText(serializedReport, false),
		slack.MsgOptionDisableLinkUnfurl(),
	)
	return err
}

func (r SlackReporter) Name() string {
	return "SlackReporter"
}
