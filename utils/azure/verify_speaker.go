package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/helper"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

type VerificationResponse struct {
	Result     string `json:"result"`
	Confidence string `json:"confidence"`
	Phrase     string `json:"phrase"`
}

func VerifySpeaker(
	ctx context.Context,
	channelID string,
	callSID string,
	audioFilePath string,
) (VerificationResponse, error) {
	var vsResp VerificationResponse
	wavFilePath, err := helper.ConvertToWAV(audioFilePath + "16")
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to convert SLN to WAV. Error: [%#v]", err)
		return vsResp, err
	}
	dat, err := ioutil.ReadFile(wavFilePath)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while reading the file to verify user. Error: [%#v]", err)
		return vsResp, err
	}
	vsReq, err := http.NewRequest(http.MethodPost, configmanager.ConfStore.AzureSpeakerAPIEndpoint+"verify", bytes.NewBuffer(dat))
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while preparing the request for Speaker Verification. Error: [%#v]", err)
		return vsResp, err
	}
	// Set required query parameters
	q := vsReq.URL.Query()
	q.Add("verificationProfileId", call.GetAuthProfileID(channelID))
	vsReq.URL.RawQuery = q.Encode()
	vsReq.Header.Set("Content-Type", "application/octet-stream")
	vsReq.Header.Set("Ocp-Apim-Subscription-Key", configmanager.ConfStore.AzureSpeakerAPIKey)

	client := &http.Client{
		Transport: &http.Transport{
			Dial:                (&net.Dialer{Timeout: 5 * time.Second}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
		Timeout: time.Duration(10 * time.Second),
	}
	defer client.CloseIdleConnections()

	var response *http.Response
	for i := 0; i < configmanager.ConfStore.ARIAPIRetry; i++ {
		// Make the http request
		response, err = client.Do(vsReq)
		if response != nil {
			defer response.Body.Close()
		}
		if err != nil || response == nil || response.StatusCode < 200 || response.StatusCode >= 300 {
			ymlogger.LogErrorf(callSID, "Error while getting the response for verifying the user. Error: [%#v]. Retrying......", err)
			continue
		}
		break
	}
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while getting the response for verifying the user. Error: [%#v]", err)
		return vsResp, err
	}
	if response != nil && (response.StatusCode < 200 || response.StatusCode >= 300) {
		ymlogger.LogErrorf(callSID, "Non 2xx response while verifying the user. StatusCode: [%#v].", response.StatusCode)
		return vsResp, errors.New("Non 2xx response")
	}
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return vsResp, err
	}
	err = json.Unmarshal(respBody, &vsResp)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while unmarshalling the response of User Verification. Body: [%#v]", respBody)
		return vsResp, err
	}
	return vsResp, nil
}
