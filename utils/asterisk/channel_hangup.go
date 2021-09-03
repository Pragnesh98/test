package asterisk

import (
	"context"
	"net"
	"net/http"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

func HangupChannel(
	ctx context.Context,
	channelID string,
	callSID string,
	hangupStr string,
) error {

	ymlogger.LogDebugf(callSID, "Trying to hangup the channel. ChannelID: [%#v] Hangup: [%#v]", channelID, hangupStr)
	// Prepare the http request for destroying the channel
	chanDelReq, err := http.NewRequest(http.MethodDelete, configmanager.ConfStore.ARIURL+"/channels/"+channelID, nil)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to form call create request. Error: [%#v]", err)
		return err
	}

	// Set Basic authentication for the request
	chanDelReq.SetBasicAuth(configmanager.ConfStore.ARIUsername, configmanager.ConfStore.ARIPassword)

	// Set required query parameters
	q := chanDelReq.URL.Query()
	q.Add("reason", hangupStr)
	chanDelReq.URL.RawQuery = q.Encode()
	chanDelReq.Header.Set("Connection", "close")

	// Initlialize HTTP client
	client := &http.Client{
		Transport: &http.Transport{
			Dial:                (&net.Dialer{Timeout: 3 * time.Second}).Dial,
			TLSHandshakeTimeout: 3 * time.Second,
		},
		Timeout: time.Duration(5 * time.Second),
	}
	defer client.CloseIdleConnections()
	// Make the http request
	response, err := client.Do(chanDelReq)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		ymlogger.LogErrorf(callSID, "Error while destroying the channel. StatusCode: [%#v]", response.StatusCode)
		return err
	}
	return nil
}
