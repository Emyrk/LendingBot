package bitfinexlender

import (
	"errors"
	"strings"
)

// StrategyConf ...
type StrategyConf struct {
	Active     string
	MarginBot  MarginBotConf
	CascadeBot CascadeBotConf
}

func ExecuteStrategy(conf BotConfig, dryRun bool) (err error) {
	// Sanity check
	if conf.API == nil {
		return errors.New("Please initialize the API instance first")
	}

	switch strings.ToLower(conf.Strategy.Active) {
	case "marginbot":
		return strategyMarginBot(conf, dryRun)
	case "cascadebot":
		return strategyCascadeBot(conf, dryRun)
	}

	return errors.New("Undefined strategy")
}
