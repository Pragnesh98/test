package azure

import (
	"testing"

	"bitbucket.org/yellowmessenger/asterisk-ari/bothelper"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
)

func TestSpeechEndpoint(t *testing.T) {
	t.Run("no_custom_endpoint", func(t *testing.T) {
		endpoint := getSTTEndpoint("", "test")
		if endpoint != "test" {
			t.Errorf("Expected endpoint = %q, got %q", "test", endpoint)
		}
	})

	t.Run("boost_phrases_only", func(t *testing.T) {
		botOptions := bothelper.BotOptions{
			MicrosoftSTTOptions: bothelper.MicrosoftSTTOptions{
				EndpointId: "endpoint",
			},
		}
		call.SetBotOptions("channelId", &botOptions)
		endpoint := getSTTEndpoint("channelId", "http://www.google.com")
		expected := "http://www.google.com?cid=endpoint"
		if endpoint != expected {
			t.Errorf("Expected %q, got %q", expected, endpoint)
		}
	})
}
