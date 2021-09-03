package newrelic

import (
	"errors"
	"os"
)

// SendCustomEvent sends custom event to newrelic
func SendCustomEvent(metricName string, metric map[string]interface{}) error {
	hostName, err := os.Hostname()
	if err != nil {
		return errors.New("Failed sending the metric. Hostname not found")
	}
	metric["host"] = hostName
	App.RecordCustomEvent(metricName, metric)
	return nil
}
