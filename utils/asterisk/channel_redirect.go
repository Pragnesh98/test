package asterisk

import (
	"context"
	"net"
	"net/http"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/phonenumber"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

func RedirectChannel(
	ctx context.Context,
	channelID string,
	callSID string,
	number phonenumber.PhoneNumber,
	spanType string,
) error {
	// Prepare the http request for destroying the channel
	chanRedReq, err := http.NewRequest(http.MethodPost, configmanager.ConfStore.ARIURL+"/channels/"+channelID+"/redirect", nil)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to form channel redirect request. Error: [%#v]", err)
		return err
	}

	// Set Basic authentication for the request
	chanRedReq.SetBasicAuth(configmanager.ConfStore.ARIUsername, configmanager.ConfStore.ARIPassword)

	dialingNumber := number.NationalFormat
	if number.IsLandline {
		dialingNumber = number.WithZeroNationalFormat
	}
	// Set required query parameters
	q := chanRedReq.URL.Query()
	if spanType == call.PipeTypePRI.String() {
		q.Add("endpoint", "dahdi/i1/"+dialingNumber)
	} else {
		q.Add("endpoint", "SIP/"+dialingNumber)
	}
	chanRedReq.URL.RawQuery = q.Encode()
	chanRedReq.Header.Set("Connection", "close")

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
	response, err := client.Do(chanRedReq)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		ymlogger.LogErrorf(callSID, "Error while redirecting the channel. StatusCode: [%#v]. Error: [%#v]", response.StatusCode, err)
		return err
	}
	return nil
}
