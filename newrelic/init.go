package newrelic

import (
	"os"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// App contains the newrelic application
var App *newrelic.Application

// InitNewRelicApp initializes the New Relic app
func InitNewRelicApp() error {
	var err error
	App, err = newrelic.NewApplication(
		newrelic.ConfigAppName("Telephony APIs"),
		newrelic.ConfigLicense(os.Getenv("NEW_RELIC_LICENSE_KEY")),
		// newrelic.ConfigDebugLogger(os.Stdout),
	)
	if err != nil {
		return err
	}
	return nil
}
