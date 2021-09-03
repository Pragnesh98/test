package asterisk

import (
	"context"
	"net"
	"net/http"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

// import (
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"io/ioutil"
// 	"net"
// 	"net/http"
// 	"path/filepath"
// 	"strings"
// 	"time"

// 	"bitbucket.org/yellowmessenger/asterisk-ari/call"
// 	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
// 	"bitbucket.org/yellowmessenger/asterisk-ari/phonenumber"
// 	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
// 	"github.com/CyCoreSystems/ari"
// )

// // GetChanVarResponse holds the response from Get channel variable asterisk API
// type GetChanVarResponse struct {
// 	Value string `json:"value"`
// }

// func Answer(
// 	ctx context.Context,
// 	h *ari.ChannelHandle,
// ) error {
// 	ymlogger.LogInfof(call.GetSID(h.ID()), "Going to answer the call. Channel ID: [%#v]", h.ID())
// 	if err := h.Answer(); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func HangupChannel(
// 	ctx context.Context,
// 	channelID string,
// 	hangupStr string,
// ) error {

// 	// Prepare the http request for destroying the channel
// 	chanDelReq, err := http.NewRequest(http.MethodDelete, configmanager.ConfStore.ARIURL+"/channels/"+channelID, nil)
// 	if err != nil {
// 		ymlogger.LogErrorf(call.GetSID(channelID), "Failed to form call create request. Error: [%#v]", err)
// 		return err
// 	}

// 	// Set Basic authentication for the request
// 	chanDelReq.SetBasicAuth(configmanager.ConfStore.ARIUsername, configmanager.ConfStore.ARIPassword)

// 	// Set required query parameters
// 	q := chanDelReq.URL.Query()
// 	q.Add("reason", hangupStr)
// 	chanDelReq.URL.RawQuery = q.Encode()
// 	chanDelReq.Header.Set("Connection", "close")

// 	// Initlialize HTTP client
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			Dial:                (&net.Dialer{Timeout: 3 * time.Second}).Dial,
// 			TLSHandshakeTimeout: 3 * time.Second,
// 		},
// 		Timeout: time.Duration(5 * time.Second),
// 	}
// 	defer client.CloseIdleConnections()
// 	// Make the http request
// 	response, err := client.Do(chanDelReq)
// 	if err != nil {
// 		return err
// 	}
// 	defer response.Body.Close()

// 	if response.StatusCode < 200 || response.StatusCode >= 300 {
// 		ymlogger.LogErrorf(call.GetSID(channelID), "Error while destroying the channel. StatusCode: [%#v]", response.StatusCode)
// 		return err
// 	}
// 	return nil
// }

// func RedirectChannel(
// 	ctx context.Context,
// 	channelID string,
// 	number phonenumber.PhoneNumber,
// 	spanType string,
// ) error {
// 	// Prepare the http request for destroying the channel
// 	chanRedReq, err := http.NewRequest(http.MethodPost, configmanager.ConfStore.ARIURL+"/channels/"+channelID+"/redirect", nil)
// 	if err != nil {
// 		ymlogger.LogErrorf(call.GetSID(channelID), "Failed to form channel redirect request. Error: [%#v]", err)
// 		return err
// 	}

// 	// Set Basic authentication for the request
// 	chanRedReq.SetBasicAuth(configmanager.ConfStore.ARIUsername, configmanager.ConfStore.ARIPassword)

// 	dialingNumber := number.NationalFormat
// 	if number.IsLandline {
// 		dialingNumber = number.WithZeroNationalFormat
// 	}
// 	// Set required query parameters
// 	q := chanRedReq.URL.Query()
// 	if spanType == call.PipeTypePRI.String() {
// 		q.Add("endpoint", "dahdi/i1/"+dialingNumber)
// 	} else {
// 		q.Add("endpoint", "SIP/"+dialingNumber)
// 	}
// 	chanRedReq.URL.RawQuery = q.Encode()
// 	chanRedReq.Header.Set("Connection", "close")

// 	// Initlialize HTTP client
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			Dial:                (&net.Dialer{Timeout: 3 * time.Second}).Dial,
// 			TLSHandshakeTimeout: 3 * time.Second,
// 		},
// 		Timeout: time.Duration(5 * time.Second),
// 	}
// 	defer client.CloseIdleConnections()
// 	// Make the http request
// 	response, err := client.Do(chanRedReq)
// 	if err != nil {
// 		return err
// 	}
// 	defer response.Body.Close()

// 	if response.StatusCode < 200 || response.StatusCode >= 300 {
// 		ymlogger.LogErrorf(call.GetSID(channelID), "Error while redirecting the channel. StatusCode: [%#v]. Error: [%#v]", response.StatusCode, err)
// 		return err
// 	}
// 	return nil
// }

// // SetChannelVariable sets the channel variable on a channel
// func SetChannelVariable(
// 	ctx context.Context,
// 	channelID string,
// 	variable string,
// 	value string,
// ) error {
// 	ymlogger.LogDebugf(call.GetSID(channelID), "Got request to set the channel variable. Variable: [%s] Value: [%s] ChannelID: [%s]", variable, value, channelID)
// 	chanVarReq, err := http.NewRequest(
// 		http.MethodPost,
// 		configmanager.ConfStore.ARIURL+"/channels/"+channelID+"/variable",
// 		nil,
// 	)
// 	if err != nil {
// 		ymlogger.LogErrorf(call.GetSID(channelID), "Error while preparing the SetChanVar request. Error: [%#v]", err)
// 		return err
// 	}

// 	// Set Basic authentication for the request
// 	chanVarReq.SetBasicAuth(configmanager.ConfStore.ARIUsername, configmanager.ConfStore.ARIPassword)

// 	// Set required query parameters
// 	q := chanVarReq.URL.Query()
// 	q.Add("variable", variable)
// 	q.Add("value", value)
// 	chanVarReq.URL.RawQuery = q.Encode()
// 	chanVarReq.Header.Set("Connection", "close")

// 	// Initlialize HTTP client
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			Dial:                (&net.Dialer{Timeout: 3 * time.Second}).Dial,
// 			TLSHandshakeTimeout: 3 * time.Second,
// 		},
// 		Timeout: time.Duration(5 * time.Second),
// 	}
// 	defer client.CloseIdleConnections()

// 	// Make the http request
// 	response, err := client.Do(chanVarReq)
// 	if err != nil {
// 		ymlogger.LogErrorf(call.GetSID(channelID), "Error while getting the response for ChanSetVar. Error: [%#v]", err)
// 		return err
// 	}
// 	defer response.Body.Close()

// 	if response.StatusCode < 200 || response.StatusCode >= 300 {
// 		ymlogger.LogErrorf(call.GetSID(channelID), "Error while setting the channel variable. StatusCode: [%#v]", response.StatusCode)
// 		return errors.New("Error while setting the channel variable")
// 	}
// 	return nil
// }

// // GetChannelVariable extracts the channel variable
// func GetChannelVariable(
// 	ctx context.Context,
// 	channelID string,
// 	variable string,
// ) (string, error) {
// 	ymlogger.LogDebugf(call.GetSID(channelID), "Got request to get the channel variable. Variable: [%s] ChannelID: [%s]", variable, channelID)
// 	var value string
// 	chanVarReq, err := http.NewRequest(
// 		http.MethodGet,
// 		configmanager.ConfStore.ARIURL+"/channels/"+channelID+"/variable",
// 		nil,
// 	)
// 	if err != nil {
// 		ymlogger.LogErrorf(call.GetSID(channelID), "Error while preparing the SetChanVar request. Error: [%#v]", err)
// 		return value, err
// 	}

// 	// Set Basic authentication for the request
// 	chanVarReq.SetBasicAuth(configmanager.ConfStore.ARIUsername, configmanager.ConfStore.ARIPassword)

// 	q := chanVarReq.URL.Query()
// 	q.Add("variable", variable)
// 	chanVarReq.URL.RawQuery = q.Encode()
// 	chanVarReq.Header.Set("Connection", "close")

// 	// Initlialize HTTP client
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			Dial:                (&net.Dialer{Timeout: 3 * time.Second}).Dial,
// 			TLSHandshakeTimeout: 3 * time.Second,
// 		},
// 		Timeout: time.Duration(5 * time.Second),
// 	}
// 	defer client.CloseIdleConnections()

// 	// Make the http request
// 	response, err := client.Do(chanVarReq)
// 	if err != nil {
// 		ymlogger.LogErrorf(call.GetSID(channelID), "Error while getting the response for ChanSetVar. Error: [%#v]", err)
// 		return value, err
// 	}
// 	defer response.Body.Close()

// 	if response.StatusCode < 200 || response.StatusCode >= 300 {
// 		ymlogger.LogErrorf(call.GetSID(channelID), "Error while getting the channel variable. StatusCode: [%#v] [%#v]", response.StatusCode, chanVarReq.URL)
// 		return value, nil
// 	}
// 	respBody, err := ioutil.ReadAll(response.Body)
// 	if err != nil {
// 		ymlogger.LogErrorf(call.GetSID(channelID), "Error while reading the response body for ChanSetVar response. Error: [%#v]", err)
// 		return value, err
// 	}
// 	var getChanVarRes GetChanVarResponse
// 	err = json.Unmarshal(respBody, &getChanVarRes)
// 	if err != nil {
// 		ymlogger.LogErrorf(call.GetSID(channelID), "Error while unmarshalling the response. Body: [%#v]", respBody)
// 		return value, nil
// 	}
// 	return getChanVarRes.Value, nil
// }

// func CreateCall(
// 	ctx context.Context,
// 	number phonenumber.PhoneNumber,
// 	extension phonenumber.PhoneNumber,
// 	context string,
// 	callerID phonenumber.PhoneNumber,
// 	pipeType string,
// ) (ari.ChannelData, error) {
// 	// callRes holds the response from the http request
// 	var callRes ari.ChannelData

// 	// Prepare the http request for creating the call
// 	callReq, err := http.NewRequest(http.MethodPost, configmanager.ConfStore.ARIURL+"/channels", nil)
// 	if err != nil {
// 		ymlogger.LogErrorf("CreateCall", "Failed to form call create request. Error: [%#v]", err)
// 		return callRes, err
// 	}

// 	// Set Basic authentication for the request
// 	callReq.SetBasicAuth(configmanager.ConfStore.ARIUsername, configmanager.ConfStore.ARIPassword)

// 	dialingNumber := number.NationalFormat
// 	if number.IsLandline {
// 		dialingNumber = number.WithZeroNationalFormat
// 	}
// 	// Set required query parameters
// 	q := callReq.URL.Query()
// 	if strings.ToLower(pipeType) == "sip" {
// 		q.Add("endpoint", "SIP/"+configmanager.ConfStore.SIPIP+"/"+dialingNumber)
// 		q.Add("extension", extension.LocalFormat)
// 		q.Add("callerId", callerID.LocalFormat)
// 	} else {
// 		q.Add("endpoint", "dahdi/i1/"+dialingNumber)
// 		q.Add("extension", extension.NationalFormat)
// 		q.Add("callerId", callerID.NationalFormat)
// 	}
// 	q.Add("context", context)
// 	callReq.URL.RawQuery = q.Encode()

// 	// Initlialize HTTP client
// 	client := &http.Client{
// 		Transport: &http.Transport{
// 			Dial:                (&net.Dialer{Timeout: 3 * time.Second}).Dial,
// 			TLSHandshakeTimeout: 3 * time.Second,
// 		},
// 		Timeout: time.Duration(5 * time.Second),
// 	}
// 	defer client.CloseIdleConnections()
// 	// // Make the http request
// 	// response, err := client.Do(callReq)
// 	// if response != nil {
// 	// 	defer response.Body.Close()
// 	// }
// 	// if err != nil {
// 	// 	return callRes, err
// 	// }

// 	// if response.StatusCode < 200 || response.StatusCode >= 300 {
// 	// 	ymlogger.LogErrorf("CreateCall", "Error while initiating the call. StatusCode: [%#v]", response.StatusCode)
// 	// 	return callRes, err
// 	// }
// 	var response *http.Response
// 	for i := 0; i < configmanager.ConfStore.ARIAPIRetry; i++ {
// 		// Make the http request
// 		response, err = client.Do(callReq)
// 		if response != nil {
// 			defer response.Body.Close()
// 		}
// 		if err != nil || response == nil || response.StatusCode < 200 || response.StatusCode >= 300 {
// 			ymlogger.LogErrorf("CreateCall", "Error while creating the channel. Error: [%#v]. Retrying......", err)
// 			continue
// 		}
// 		break
// 	}
// 	if err != nil {
// 		ymlogger.LogErrorf("CreateCall", "Error while creating the channel. Error: [%#v]", response.StatusCode, err)
// 		return callRes, err
// 	}
// 	if response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
// 		ymlogger.LogErrorf("CreateCall", "Non 2xx response while creating the channel. StatusCode: [%#v].", response.StatusCode)
// 		return callRes, errors.New("Non 2xx response")
// 	}
// 	respBody, err := ioutil.ReadAll(response.Body)
// 	if err != nil {
// 		return callRes, err
// 	}
// 	err = json.Unmarshal(respBody, &callRes)
// 	if err != nil {
// 		ymlogger.LogErrorf("CreateCall", "Error while unmarshalling the response. Body: [%#v]", respBody)
// 		return callRes, nil
// 	}
// 	return callRes, nil
// }

// func Play(
// 	ctx context.Context,
// 	h *ari.ChannelHandle,
// 	channelID string,
// 	fileName string,
// ) (*ari.PlaybackHandle, error) {

// 	ymlogger.LogInfof(call.GetSID(channelID), "Running Play on channel: [%#v] FileName: [%s]", channelID, fileName)
// 	fileWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))

// 	playbackHandle, err := h.Play(channelID, "sound:"+fileWithoutExt)
// 	if err != nil {
// 		ymlogger.LogErrorf(call.GetSID(channelID), "Failed to play sound. Error: [%#v]", err)
// 		return playbackHandle, err
// 	}

// 	ymlogger.LogInfo(call.GetSID(channelID), "Completed Playback")
// 	return playbackHandle, err
// }

// func Record(
// 	ctx context.Context,
// 	h *ari.ChannelHandle,
// ) (*ari.LiveRecordingHandle, error) {
// 	ymlogger.LogInfof(call.GetSID(h.ID()), "Going to record the call. Channel ID=[%#v]", h.ID())
// 	recordHandler, err := h.Record(h.ID(), &ari.RecordingOptions{
// 		Format:      configmanager.ConfStore.RecordingFormat,
// 		MaxDuration: call.GetRecordingMaxDuration(h.ID()),
// 		MaxSilence:  call.GetRecordingSilenceDuration(h.ID()),
// 		Exists:      "overwrite",
// 		Beep:        true,
// 		Terminate:   configmanager.ConfStore.RecordingTerminationKey,
// 	})
// 	if err != nil {
// 		return recordHandler, err
// 	}

// 	ymlogger.LogInfof(call.GetSID(h.ID()), "Recording is enqueued. Channel=[%#v]", h.ID())
// 	return recordHandler, nil
// }

// func RecordCall(
// 	ctx context.Context,
// 	fileName string,
// 	handler *ari.ChannelHandle,
// ) (*ari.LiveRecordingHandle, error) {
// 	ymlogger.LogInfof(call.GetSID(handler.ID()), "Going to record the complete call. Channel ID=[%#v]", handler.ID())
// 	callRecordHandle, err := handler.Record(fileName, &ari.RecordingOptions{
// 		Format:      configmanager.ConfStore.CallRecordingFormat,
// 		MaxDuration: 3600 * time.Second,
// 		Exists:      "overwrite",
// 		Terminate:   "none",
// 	})
// 	if err != nil {
// 		return callRecordHandle, err
// 	}

// 	ymlogger.LogInfof(call.GetSID(handler.ID()), "Call Recording is enqueued. Channel=[%#v]. CallSID: [%s]", handler.ID(), fileName)
// 	return callRecordHandle, nil
// }

func ChannelExists(
	ctx context.Context,
	channelID string,
) bool {
	chanExistsReq, err := http.NewRequest(
		http.MethodGet,
		configmanager.ConfStore.ARIURL+"/channels/"+channelID,
		nil,
	)
	if err != nil {
		ymlogger.LogErrorf(call.GetSID(channelID), "Error while preparing the ChanExists request. Error: [%#v]", err)
		return true
	}

	// Set Basic authentication for the request
	chanExistsReq.SetBasicAuth(configmanager.ConfStore.ARIUsername, configmanager.ConfStore.ARIPassword)

	chanExistsReq.Header.Set("Connection", "close")

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
	response, err := client.Do(chanExistsReq)
	if err != nil {
		ymlogger.LogErrorf(call.GetSID(channelID), "Error while getting the response for ChanExists. Error: [%#v]", err)
		return true
	}
	defer response.Body.Close()

	if response.StatusCode == 404 {
		ymlogger.LogErrorf(call.GetSID(channelID), "Channel does not exists. StatusCode: [%#v]", response.StatusCode)
		return false
	}
	return true
}
