package asterisk

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
)

func SnoopChannel(
	ctx context.Context,
	channelID string,
	callSID string,
	snoopDirection string,
	snoopID string,
) (ari.ChannelData, error) {
	// snoopRes holds the response from the http request
	var snoopRes ari.ChannelData
	url := configmanager.ConfStore.ARIURL + "/channels/" + channelID + "/snoop"
	if len(snoopID) > 0 {
		url = url + "/" + snoopID
	}
	// Prepare the http request for creating the call
	snoopReq, err := http.NewRequest(
		http.MethodPost,
		url,
		nil,
	)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to form snoop request. Error: [%#v]", err)
		return snoopRes, err
	}

	// Set Basic authentication for the request
	snoopReq.SetBasicAuth(configmanager.ConfStore.ARIUsername, configmanager.ConfStore.ARIPassword)

	// Set required query parameters
	q := snoopReq.URL.Query()
	q.Add("spy", snoopDirection)
	q.Add("app", configmanager.ConfStore.ARIApplication)
	snoopReq.URL.RawQuery = q.Encode()
	snoopReq.Header.Set("Connection", "close")

	// Initlialize HTTP client
	client := &http.Client{
		Transport: &http.Transport{
			Dial:                (&net.Dialer{Timeout: 3 * time.Second}).Dial,
			TLSHandshakeTimeout: 3 * time.Second,
		},
		Timeout: time.Duration(5 * time.Second),
	}

	defer client.CloseIdleConnections()
	var response *http.Response
	for i := 0; i < configmanager.ConfStore.ARIAPIRetry; i++ {
		// Make the http request
		response, err = client.Do(snoopReq)
		if response != nil {
			defer response.Body.Close()
		}
		if err != nil || response == nil || response.StatusCode < 200 || response.StatusCode >= 300 {
			ymlogger.LogErrorf(callSID, "Error while getting the response for SnoopChannel. Error: [%#v]. Retrying......", err)
			continue
		}
		break
	}
	if err != nil || response == nil || response.StatusCode < 200 || response.StatusCode >= 300 {
		ymlogger.LogErrorf(callSID, "Error while snooping the channel. StatusCode: [%#v]. Error: [%#v]", response.StatusCode, err)
		return snoopRes, errors.New("Error while snooping the channel")
	}
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return snoopRes, err
	}
	err = json.Unmarshal(respBody, &snoopRes)
	if err != nil {
		ymlogger.LogErrorf("CreateCall", "Error while unmarshalling the response. Body: [%#v]", respBody)
		return snoopRes, err
	}
	return snoopRes, nil
}
