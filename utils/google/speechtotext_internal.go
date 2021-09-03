package google

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/newrelic"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/helper"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

// STTRequest is the complete request internal STT
type STTRequest struct {
	CallSID          string           `json:"callSid"`
	RecognizeRequest RecognizeRequest `json:"recognizeRequest"`
}

// RecognizeRequest is the request for speech recognition for google format
type RecognizeRequest struct {
	Audio  RecognitionAudioSource `json:"audio"`
	Config RecognitionConfig      `json:"config"`
}

// RecognitionAudioSource contains the source of audio data
type RecognitionAudioSource struct {
	AudioSource RecognitionAudio `json:"AudioSource"`
}

// RecognitionAudio contains the audio data
type RecognitionAudio struct {
	Content string `json:"content"`
}

// RecognitionConfig contains the config for speech recognition request
type RecognitionConfig struct {
	Encoding        int                 `json:"encoding"`
	LanguageCode    string              `json:"language_code"`
	SampleHertzRate int32               `json:"sample_rate_hertz"`
	MetaData        RecognitionMetaData `json:"metadata"`
}

// RecognitionMetaData is metadata for speech recognition request
type RecognitionMetaData struct {
	InteractionType int `json:"interaction_type"`
}

// SpeechToTextResponse is the format for speech recognition response
type SpeechToTextResponse struct {
	Response RecognizeResponse `json:"recognizeResponse"`
}

// RecognizeResponse contains the result of the speech
type RecognizeResponse struct {
	Results []SpeechRecognitionResult `json:"results"`
}

// SpeechRecognitionResult contains the results from speech to text API response
type SpeechRecognitionResult struct {
	Alternatives []SpeechRecognitionAlternative `json:"alternatives"`
	ChannelTag   int                            `json:"channelTag"`
	LanguageCode string                         `json:"languageCode"`
}

// SpeechRecognitionAlternative contains transript and other info
type SpeechRecognitionAlternative struct {
	Transcript string     `json:"transcript"`
	Confidence float64    `json:"confidence"`
	Words      []WordInfo `json:"words,omitempty"`
}

// WordInfo contains each words info from speech recogntion response
type WordInfo struct {
	StartTime  string  `json:"startTime"`
	EndTime    string  `json:"endTime"`
	Word       string  `json:"word"`
	Confidence float64 `json:"confidence"`
	SpeakerTag int     `json:"speakerTag"`
}

// StartSpeechToTextInternal hits the internal API for recognizing the speech
func StartSpeechToTextInternal(
	ctx context.Context,
	channelID string,
	callSID string,
	languageCode string,
	fileName string,
) {
	// Open file on disk.
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		ymlogger.LogErrorf(callSID, "[InternalSpeechToText] Errow while getting the stat for file. Error: [%#v]", err)
	}

	if fileInfo == nil || fileInfo.Size() == 0 {
		ymlogger.LogInfof(callSID, "[InternalSpeechToText] File size is zero. Exiting")
		return
	}

	filePath, err := helper.ConvertToWAV8000(fileName)
	if err != nil {
		ymlogger.LogErrorf(callSID, "[InternalSpeechToText] Failed to convert SLN to wav. Error: [%#v]", err)
		return
	}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		ymlogger.LogErrorf(callSID, "[InternalSpeechToText] Errow while reading the file. Error: [%#v]", err)
		return
	}

	// Encode as base64.
	base64Encoded := base64.StdEncoding.EncodeToString(content)

	recReqBody := prepareSpeechRecoginitionRequest(callSID, channelID, languageCode, base64Encoded)
	recReqBodyJSON, err := json.Marshal(recReqBody)
	if err != nil {
		log.Printf(callSID, "[InternalSpeechToText] Errow while marshalling speech recognition request. Error: [%#v]", err)
		return
	}

	// Prepare the request
	recReq, err := http.NewRequest(http.MethodPost, "http://52.188.35.33/v1/speech_to_text/cloud/google", bytes.NewBuffer(recReqBodyJSON))
	if err != nil {
		ymlogger.LogErrorf(callSID, "[InternalSpeechToText] Error while forming request for speech recognition. Error: [%#v]", err)
		return
	}
	recReq.Host = "speechtotext.yellowmessenger.com"
	// Set the Headers
	recReq.Header.Set("Content-type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			Dial:                (&net.Dialer{Timeout: 10 * time.Second}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		},
		Timeout: time.Duration(10 * time.Second),
	}
	defer client.CloseIdleConnections()
	// Capture API response time
	sTime := time.Now()
	// Make the http request
	response, err := client.Do(recReq)
	if response != nil {
		defer response.Body.Close()
	}
	// Send API response time to newrelic
	if err := newrelic.SendCustomEvent("voice_google_stt_internal", map[string]interface{}{
		"response_time": time.Since(sTime).Milliseconds(),
	}); err != nil {
		ymlogger.LogErrorf("NewRelicMetric", "Failed to send voice_google_stt_internal metric to newrelic. Error: [%#v]", err)
	}

	if err != nil || response == nil {
		ymlogger.LogErrorf(callSID, "[InternalSpeechToText] Failed to get response from bot. Error: [%#v]. ChannelID: [%s]. ", err.Error(), channelID)
		return
	}

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ymlogger.LogErrorf(callSID, "[InternalSpeechToText] Failed to read the response body. Error: [%#v]. ChannelID: [%s].", err.Error(), channelID)
		return
	}
	var speechResponse SpeechToTextResponse
	json.Unmarshal(respBody, &speechResponse)
	ymlogger.LogInfof(callSID, "[InternalSpeechToText] Response from the API: [%#v] Raw Body: [%#v]", speechResponse, string(respBody))

	eventData := map[string]interface{}{
		"event_type": "accuracy",
	}
	if len(speechResponse.Response.Results) > 0 && len(speechResponse.Response.Results[0].Alternatives) > 0 && len(speechResponse.Response.Results[0].Alternatives[0].Transcript) > 0 {
		eventData["status"] = "success"
	} else {
		eventData["status"] = "failure"
	}
	// Send accuracy to new relic
	if err := newrelic.SendCustomEvent("voice_google_stt_internal", eventData); err != nil {
		ymlogger.LogErrorf("NewRelicMetric", "Failed to send voice_google_stt_internal accuracy metric to newrelic. Error: [%#v]", err)
	}
	return
}

func prepareSpeechRecoginitionRequest(
	callSID string,
	channelID string,
	languageCode string,
	base64Encoded string,
) STTRequest {
	recReq := STTRequest{
		CallSID: callSID,
		RecognizeRequest: RecognizeRequest{
			Audio: RecognitionAudioSource{
				AudioSource: RecognitionAudio{
					Content: base64Encoded,
				},
			},
			Config: RecognitionConfig{
				Encoding:        1, // "LINEAR16"
				LanguageCode:    call.GetSTTLanguage(channelID),
				SampleHertzRate: configmanager.ConfStore.STTSampleRate,
				MetaData: RecognitionMetaData{
					InteractionType: 7, // "VOICE_COMMAND"
				},
			},
		},
	}
	return recReq
}
