package asterisk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/phonenumber"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
)

func CreateCall(
	ctx context.Context,
	number phonenumber.PhoneNumber,
	extension phonenumber.PhoneNumber,
	context string,
	callerID phonenumber.PhoneNumber,
	pipeType string,
) (ari.ChannelData, error) {
	// callRes holds the response from the http request
	var callRes ari.ChannelData

	// postData := []byte(`{"variables":{ "SIPADDHEADER01": "P-Preferred-Identity: <sip:` + configmanager.ConfStore.SIPPilotNumber + `@` + configmanager.ConfStore.SIPIP + `>"}}`)
	postData := []byte{}
	if len(configmanager.ConfStore.DialingNumberPrefix) <= 0 {
		postData = []byte(`{"variables":{ "SIPADDHEADER01": "P-Preferred-Identity: <sip:` + configmanager.ConfStore.SIPPilotNumber + `@` + configmanager.ConfStore.SIPIP + `>"}}`)
	}
	// Prepare the http request for creating the call
	callReq, err := http.NewRequest(http.MethodPost, configmanager.ConfStore.ARIURL+"/channels", bytes.NewBuffer(postData))
	if err != nil {
		ymlogger.LogErrorf("ListenCall", "Failed to form call create request. Error: [%#v]", err)
		return callRes, err
	}

	// Set Basic authentication for the request
	callReq.SetBasicAuth(configmanager.ConfStore.ARIUsername, configmanager.ConfStore.ARIPassword)

	dialingNumber := configmanager.ConfStore.DialingNumberPrefix + number.WithZeroNationalFormat
	if number.IsLandline {
		dialingNumber = configmanager.ConfStore.DialingNumberPrefix + number.WithZeroNationalFormat
	}
	// Set required query parameters
	q := callReq.URL.Query()
	if strings.ToLower(pipeType) == "sip" && !strings.HasPrefix(strings.ToLower(dialingNumber), "sip") {
		if callerID.IsInternational {
			if callerID.E164Format == "+639498837696" {
				ymlogger.LogInfof("createCall", "Setting sip trunk lamudi: [%v]:", callerID.E164Format)
				q.Add("endpoint", "SIP/115.84.246.86/"+number.E164Format)
			} else {
				q.Add("endpoint", "SIP/"+configmanager.ConfStore.SIPIP+"/"+number.E164Format)
			}

			// q.Add("endpoint", "SIP/"+configmanager.ConfStore.SIPIP+"/"+number.E164Format)
			q.Add("extension", extension.E164Format)
			q.Add("callerId", callerID.E164Format)

		} else {
			q.Add("endpoint", "SIP/"+configmanager.ConfStore.SIPIP+"/"+dialingNumber)
			q.Add("extension", extension.LocalFormat)
			q.Add("callerId", callerID.LocalFormat)

		}
	} else if strings.HasPrefix(strings.ToLower(dialingNumber), "sip") {
		q.Add("endpoint", dialingNumber)
		q.Add("extension", extension.LocalFormat)
		q.Add("callerId", callerID.LocalFormat)

	} else if strings.ToLower(pipeType) == "gsm" {
		q.Add("endpoint", "SIP/"+callerID.NationalFormat+"/"+number.WithZeroNationalFormat)
		q.Add("extension", extension.NationalFormat)
		q.Add("callerId", callerID.NationalFormat)

	} else {
		q.Add("endpoint", "dahdi/i1/"+dialingNumber)
		q.Add("extension", extension.NationalFormat)
		q.Add("callerId", callerID.NationalFormat)
	}
	q.Add("context", context)
	if configmanager.ConfStore.SIPCALLTIMEOUT != "" {
		ymlogger.LogInfof("ListenCall", "Setting sip cancel timeout: [%v]:", configmanager.ConfStore.SIPCALLTIMEOUT)
		q.Add("timeout", configmanager.ConfStore.SIPCALLTIMEOUT)
	}
	callReq.URL.RawQuery = q.Encode()

	// Add Content Type Header
	callReq.Header.Add("Content-type", "application/json")

	// Initlialize HTTP client
	client := &http.Client{
		Transport: &http.Transport{
			Dial:                (&net.Dialer{Timeout: 3 * time.Second}).Dial,
			TLSHandshakeTimeout: 3 * time.Second,
		},
		Timeout: time.Duration(30 * time.Second),
	}
	defer client.CloseIdleConnections()
	// Make the http request
	var response *http.Response
	for i := 0; i < configmanager.ConfStore.ARIAPIRetry; i++ {
		// Make the http request
		response, err = client.Do(callReq)
		if response != nil {
			defer response.Body.Close()
		}
		if err != nil || response == nil || response.StatusCode < 200 || response.StatusCode >= 300 {
			ymlogger.LogErrorf("CreateCall", "Error while creating the channel. Error: [%#v]. Retrying......", err)
			continue
		}
		break
	}
	if err != nil {
		ymlogger.LogErrorf("CreateCall", "Error while creating the channel. Error: [%#v]", err)
		return callRes, err
	}
	if response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
		ymlogger.LogErrorf("CreateCall", "Non 2xx response while creating the channel. StatusCode: [%#v].", response.StatusCode)
		return callRes, errors.New("Non 2xx response")
	}
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return callRes, err
	}
	err = json.Unmarshal(respBody, &callRes)
	if err != nil {
		ymlogger.LogErrorf("CreateCall", "Error while unmarshalling the response. Body: [%#v]", respBody)
		return callRes, err
	}
	return callRes, nil
}

func CreateCallWithID(
	ctx context.Context,
	channelID string,
	number phonenumber.PhoneNumber,
	extension phonenumber.PhoneNumber,
	context string,
	callerID phonenumber.PhoneNumber,
	pipeType string,
) (ari.ChannelData, error) {
	// callRes holds the response from the http request
	var callRes ari.ChannelData

	// Prepare the http request for creating the call
	callReq, err := http.NewRequest(http.MethodPost, configmanager.ConfStore.ARIURL+"/channels/"+channelID, nil)
	if err != nil {
		ymlogger.LogErrorf("ListenCall", "Failed to form call create request. Error: [%#v]", err)
		return callRes, err
	}

	// Set Basic authentication for the request
	callReq.SetBasicAuth(configmanager.ConfStore.ARIUsername, configmanager.ConfStore.ARIPassword)

	dialingNumber := number.NationalFormat
	if number.IsLandline {
		dialingNumber = number.WithZeroNationalFormat
	}
	// Set required query parameters
	q := callReq.URL.Query()
	if strings.ToLower(pipeType) == "sip" {
		q.Add("endpoint", "SIP/"+configmanager.ConfStore.SIPIP+"/"+dialingNumber)
		q.Add("extension", extension.LocalFormat)
		q.Add("callerId", callerID.LocalFormat)
	} else {
		q.Add("endpoint", "dahdi/i1/"+dialingNumber)
		q.Add("extension", extension.NationalFormat)
		q.Add("callerId", callerID.NationalFormat)
	}
	q.Add("context", context)
	callReq.URL.RawQuery = q.Encode()

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
	var response *http.Response
	for i := 0; i < configmanager.ConfStore.ARIAPIRetry; i++ {
		// Make the http request
		response, err = client.Do(callReq)
		if response != nil {
			defer response.Body.Close()
		}
		if err != nil || response == nil || response.StatusCode < 200 || response.StatusCode >= 300 {
			ymlogger.LogErrorf("CreateCall", "Error while creating the channel. Error: [%#v]. Retrying......", err)
			continue
		}
		break
	}
	if err != nil {
		ymlogger.LogErrorf("CreateCall", "Error while creating the channel. Error: [%#v]", response.StatusCode, err)
		return callRes, err
	}
	if response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
		ymlogger.LogErrorf("CreateCall", "Non 2xx response while creating the channel. StatusCode: [%#v].", response.StatusCode)
		return callRes, errors.New("Non 2xx response")
	}
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return callRes, err
	}
	err = json.Unmarshal(respBody, &callRes)
	if err != nil {
		ymlogger.LogErrorf("CreateCall", "Error while unmarshalling the response. Body: [%#v]", respBody)
		return callRes, err
	}
	return callRes, nil
}
