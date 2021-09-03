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
)

// GetChanVarResponse holds the response from Get channel variable asterisk API
type GetChanVarResponse struct {
	Value string `json:"value"`
}

// SetChannelVariable sets the channel variable on a channel
func SetChannelVariable(
	ctx context.Context,
	channelID string,
	callSID string,
	variable string,
	value string,
) error {
	ymlogger.LogDebugf(callSID, "Got request to set the channel variable. Variable: [%s] Value: [%s] ChannelID: [%s]", variable, value, channelID)
	chanVarReq, err := http.NewRequest(
		http.MethodPost,
		configmanager.ConfStore.ARIURL+"/channels/"+channelID+"/variable",
		nil,
	)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while preparing the SetChanVar request. Error: [%#v]", err)
		return err
	}

	// Set Basic authentication for the request
	chanVarReq.SetBasicAuth(configmanager.ConfStore.ARIUsername, configmanager.ConfStore.ARIPassword)

	// Set required query parameters
	q := chanVarReq.URL.Query()
	q.Add("variable", variable)
	q.Add("value", value)
	chanVarReq.URL.RawQuery = q.Encode()
	chanVarReq.Header.Set("Connection", "close")

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
	response, err := client.Do(chanVarReq)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while getting the response for ChanSetVar. Error: [%#v]", err)
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		ymlogger.LogErrorf(callSID, "Error while setting the channel variable. StatusCode: [%#v]", response.StatusCode)
		return errors.New("Error while setting the channel variable")
	}
	return nil
}

// GetChannelVariable extracts the channel variable
func GetChannelVariable(
	ctx context.Context,
	channelID string,
	callSID string,
	variable string,
) (string, error) {
	ymlogger.LogDebugf(callSID, "Got request to get the channel variable. Variable: [%s] ChannelID: [%s]", variable, channelID)
	var value string
	chanVarReq, err := http.NewRequest(
		http.MethodGet,
		configmanager.ConfStore.ARIURL+"/channels/"+channelID+"/variable",
		nil,
	)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while preparing the SetChanVar request. Error: [%#v]", err)
		return value, err
	}

	// Set Basic authentication for the request
	chanVarReq.SetBasicAuth(configmanager.ConfStore.ARIUsername, configmanager.ConfStore.ARIPassword)

	q := chanVarReq.URL.Query()
	q.Add("variable", variable)
	chanVarReq.URL.RawQuery = q.Encode()
	chanVarReq.Header.Set("Connection", "close")

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
	response, err := client.Do(chanVarReq)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while getting the response for ChanSetVar. Error: [%#v]", err)
		return value, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		ymlogger.LogErrorf(callSID, "Error while getting the channel variable. StatusCode: [%#v] [%#v]", response.StatusCode, chanVarReq.URL)
		return value, nil
	}
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while reading the response body for ChanSetVar response. Error: [%#v]", err)
		return value, err
	}
	var getChanVarRes GetChanVarResponse
	err = json.Unmarshal(respBody, &getChanVarRes)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while unmarshalling the response. Body: [%#v]", respBody)
		return value, nil
	}
	return getChanVarRes.Value, nil
}
