package bothelper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/utils/ratelimit"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/newrelic"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

var botHTTPClient *http.Client

// BotRequest contains the bot request parameter for the bot API
type BotRequest struct {
	RequestID   string      `json:"traceId"`
	From        string      `json:"from"`
	To          string      `json:"to"`
	MessageData messageData `json:"messageData"`
}

type messageData struct {
	Message          string      `json:"message"`
	CallSID          string      `json:"call_sid"`
	ChannelID        string      `json:"channel_id"`
	RecordingURL     string      `json:"recording_url"`
	STTLanguage      string      `json:"stt_language"`
	To               string      `json:"to"`
	Direction        string      `json:"direction"`
	Language         string      `json:"language"`
	DetectedLanguage string      `json:"detected_language"`
	Interjected      bool        `json:"interjected"`
	InterjectedWords []string    `json:"interjected_words"`
	ChildLegStatus   string      `json:"child_leg_status"`
	ExtraParams      interface{} `json:"extra_params"`
}

// BotResponse is the response received from the bot
type BotResponse struct {
	Data       BotData `json:"data"`
	Disconnect bool    `json:"disconnect"`
	Message    string  `json:"message"`
	Success    bool    `json:"success"`
	BotID      string  `json:"botId"`
}

// BotData contains the parameters received from the bot
type BotData struct {
	Forward       bool       `json:"forward"`
	ForwardingNum string     `json:"forward_num"`
	Language      string     `json:"lang"`
	Message       string     `json:"message"`
	Speed         float64    `json:"speed"`
	TextType      string     `json:"text_type"`
	TTSEngine     string     `json:"tts_engine"`
	CaptureDTMF   bool       `json:"capture_dtmf"`
	Options       BotOptions `json:"options"`
}

type MicrosoftSTTOptions struct {
	EndpointId string `json:"endpoint_id"`
}

type VoskSTTOptions struct {
	LanguageModel string `json:"language_model"`
}

// TTSParams are params for TTS
type TTSParams struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	TextType string `json:"text_type"`
}

// TTSOptions are all TTS options
type TTSOptions struct {
	Message       string  `json:"message"`
	TextType      string  `json:"text_type"`
	TTSEngine     string  `json:"tts_engine"`
	VoiceLanguage string  `json:"voice_language"`
	VoiceID       string  `json:"voice_id"`
	Speed         float64 `json:"speed"`
	Pitch         float64 `json:"pitch"`
	Disconnect    bool    `json:"disconnect"`
}

// BotOptions are the custom bot parameters which we can receive from the bot response
type BotOptions struct {
	AuthenticateUser         bool                `json:"authenticate_user"`
	AuthProfileID            string              `json:"auth_profile_id"`
	TTS                      []TTSParams         `json:"tts"`
	TTSEngine                string              `json:"tts_engine"`
	TextType                 string              `json:"text_type"`
	Speed                    float64             `json:"speed"`
	Pitch                    float64             `json:"pitch"`
	DeviceProfiles           []string            `json:"device_profiles"`
	STTEngine                string              `json:"stt_engine"`
	STTMode                  string              `json:"stt_mode"`
	STTLanguage              string              `json:"stt_language"`
	VoiceID                  string              `json:"voice_id"`
	VoiceOTP                 bool                `json:"voice_otp"`
	CaptureVoice             bool                `json:"capture_voice"`
	RecordingBeep            bool                `json:"recording_beep"`
	RecordingSilenceDuration int                 `json:"recording_silence_duration"`
	FinalSilenceDurationMs   int                 `json:"final_silence_duration_millis"`
	InitialSilenceDurationMs int                 `json:"initial_silence_duration_millis"`
	RecordingMaxDuration     int                 `json:"recording_max_duration"`
	MaxBotFailureCount       int8                `json:"max_bot_failure_count"`
	HangupString             string              `json:"hangup_string"`
	BoostPhrases             []string            `json:"boost_phrases"`
	MicrosoftSTTOptions      MicrosoftSTTOptions `json:"microsoft_stt_options"`
	VoskSTTOptions           VoskSTTOptions      `json:"vosk_stt_options"`
	SecondarySTTEngine       string              `json:"secondary_stt_engine"`
	Interject                bool                `json:"interject"`
	InterjectionLanguage     string              `json:"interjection_language"`
	InterjectUtterances      []string            `json:"interject_utterances"`
	PrefetchTTS              []string            `json:"prefetch_tts"`
	UseNewMsSubscritpion     bool                `json:"use_new_ms_sub"`
	TTSQuality               string              `json:"tts_quality"`
	DetectLanguageCode       []string            `json:"detect_language_codes"`
	ForwardingCallerID       string              `json:"forward_caller_id"`
}

// InitBotHTTPclient initializes the bot's HTTP client
func InitBotHTTPclient() {
	botHTTPClient = &http.Client{
		Transport: &http.Transport{
			// Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConnsPerHost:   100,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   3 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: time.Duration(configmanager.ConfStore.BotTimeoutPeriod) * time.Millisecond,
	}
}

// GetBotResponse gets the bot response based on the given parameters
func GetBotResponse(
	ctx context.Context,
	traceID string,
	channelID string,
	callSID string,
	botID string,
	campaignID string,
	text string,
	from string,
	to string,
	language string,
	detectedLanguage string,
	direction string,
	recordingFileName string,
	sttLanguage string,
	interjected bool,
	interjectedWords []string,
	childLegStatus string,
	extraParams interface{},
	botRateLimiter *ratelimit.AdaptiveRateLimiter,
) (BotResponse, error) {
	// from = "+919413745250"
	// if to == "+918068983600" {
	// 	to = "+918068402356"
	// }
	// if to == "+918068983601" {
	// 	to = "+918068402349"
	// }
	var botResponse BotResponse
	botBody := formBotBody(ctx, traceID, channelID, callSID, botID, text, from, to, language, detectedLanguage, direction, recordingFileName, sttLanguage, interjected, interjectedWords, childLegStatus, extraParams)
	botBodyJSON, err := json.Marshal(botBody)
	if err != nil {
		return botResponse, err
	}
	ymlogger.LogDebugf(callSID, "Hitting Bot API with request Body: [%v], ChannelID: [%s]", string(botBodyJSON), channelID)

	botEndPoint := configmanager.ConfStore.BotEndPoint
	if to == "+918068402304" {
		botEndPoint = "https://staging.yellowmessenger.com/integrations/voice/execute"
	}

	botReq, err := http.NewRequest(http.MethodPost, botEndPoint, bytes.NewBuffer(botBodyJSON))
	if err != nil {
		return botResponse, err
	}
	if to != "+918068402304" {
		botReq.Host = "app.yellowmessenger.com"
		botReq.Header.Set("Host", "app.yellowmessenger.com")
	}
	botReq.Header.Set("Content-Type", "application/json")
	botReq.Header.Set("Authorization", configmanager.ConfStore.GoogleAccessToken)

	var response *http.Response

	for i := 0; i < configmanager.ConfStore.ARIAPIRetry; i++ {
		// Capture Bot response time
		sTime := time.Now()
		// Make the http request
		response, err = botHTTPClient.Do(botReq)

		if botRateLimiter != nil {
			botRateLimiter.RecordLatency(time.Since(sTime))
		}
		callerID := ""
		if len(to) > 1 {
			callerID = to[1:]
		}
		if err := newrelic.SendCustomEvent("voice_botapi", map[string]interface{}{
			"caller_id":     callerID,
			"campaign_id":   campaignID,
			"response_time": time.Since(sTime).Milliseconds(),
		}); err != nil {
			ymlogger.LogErrorf("NewRelicMetric", "Failed to send botAPI metric to newrelic. Error: [%#v]", err)
		}

		if response != nil {
			if response.StatusCode == 500 {
				ymlogger.LogErrorf(callSID, "Got server error response from bot. Response: [%#v]. ChannelID: [%s]. Retrying.....", response, channelID)
				continue
			}
			defer response.Body.Close()
		}

		if err != nil || response == nil {
			ymlogger.LogErrorf(callSID, "Failed to get response from bot. Error: [%#v]. ChannelID: [%s]. Retrying.....", err.Error(), channelID)
			if err := newrelic.SendCustomEvent("voice_botapi", map[string]interface{}{
				"caller_id":  callerID,
				"bot_failed": "true",
				"count":      1,
			}); err != nil {
				ymlogger.LogErrorf("NewRelicMetric", "Failed to send botAPI metric to newrelic. Error: [%#v]", err)
			}
			continue
		}
		break
	}
	if err != nil || response == nil {
		ymlogger.LogErrorf(callSID, "Failed to get response from bot. Error: [%#v]. ChannelID: [%s].", err.Error(), channelID)
		return botResponse, err
	}
	ymlogger.LogErrorf(callSID, "Failed to get response from bot. Error: [%#v]. ChannelID: [%s]. [%d]", err, channelID, response.StatusCode)
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to read the response body. Error: [%#v]. ChannelID: [%s]. Retrying.....", err.Error(), channelID)
		return botResponse, err
	}

	ymlogger.LogInfof(callSID, "Response body payload: [%#v]", string(respBody) )

	err = json.Unmarshal(respBody, &botResponse)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to unmarshal the response body. Error: [%#v]. ChannelID: [%s].", err.Error(), channelID)
		return botResponse, nil
	}
	// botResponse.Data.Options.Interject = true
	// botResponse.Data.Options.InterjectUtterances = []string{"English", "retailer", "change"}
	return botResponse, nil
}

// CheckIfBotUp checks if the bot is giving response
func CheckIfBotUp(
	ctx context.Context,
	channelID string,
	callSID string,
	text string,
	from string,
	to string,
	language string,
	direction string,
	recordingFileName string,
	sttLanguage string,
) (bool, error) {
	botBody := formBotBody(ctx, "", channelID, callSID, "", text, from, to, language, sttLanguage, direction, recordingFileName, sttLanguage, false, []string{}, "", nil)
	botBodyJSON, err := json.Marshal(botBody)
	if err != nil {
		return false, errors.New("Bot did not respond")
	}
	ymlogger.LogDebugf(callSID, "Hitting Bot API with request Body: [%v], ChannelID: [%s]", string(botBodyJSON), channelID)

	botEndPoint := configmanager.ConfStore.BotEndPoint
	if to == "+918068402304" {
		botEndPoint = "https://staging.yellowmessenger.com/integrations/voice/execute"
	}
	botReq, err := http.NewRequest(http.MethodPost, botEndPoint, bytes.NewBuffer(botBodyJSON))
	if err != nil {
		return false, errors.New("Bot did not respond")
	}
	botReq.Header.Set("Content-Type", "application/json")

	var response *http.Response
	for i := 0; i < 1; i++ {
		// Make the http request
		response, err = botHTTPClient.Do(botReq)
		if response != nil {
			defer response.Body.Close()
		}
		if err != nil || response == nil {
			ymlogger.LogErrorf(callSID, "Failed to get response from bot. Error: [%#v]. ChannelID: [%s]. Retrying.....", err.Error(), channelID)
			continue
		}
		break
	}
	if err != nil || response == nil {
		ymlogger.LogErrorf(callSID, "Failed to get response from bot. Error: [%#v]. ChannelID: [%s].", err.Error(), channelID)
		return false, errors.New("Bot did not respond")
	}
	return true, nil
}

func formBotBody(
	ctx context.Context,
	traceID string,
	channelID string,
	callSID string,
	botID string,
	text string,
	from string,
	to string,
	language string,
	detectedLanguage string,
	direction string,
	recordingFileName string,
	sttLanguage string,
	interjected bool,
	interjectedWords []string,
	childLegStatus string,
	extraParams interface{},
) BotRequest {
	recordingURL := "https://yellowmessenger.blob.core.windows.net/recordings/" +
		botID + "/" +
		time.Now().Format("2006-01-02") + "/" +
		recordingFileName + ".wav"
	botBody := BotRequest{
		RequestID: traceID,
		From:      from,
		To:        to,
		MessageData: messageData{
			Message:          text,
			CallSID:          callSID,
			ChannelID:        channelID,
			RecordingURL:     recordingURL,
			STTLanguage:      sttLanguage,
			To:               to,
			Language:         language,
			DetectedLanguage: detectedLanguage,
			Direction:        direction,
			Interjected:      interjected,
			InterjectedWords: interjectedWords,
			ChildLegStatus:   childLegStatus,
			ExtraParams:      extraParams,
		},
	}
	return botBody
}

// UnmarshalJSON Custom JSON Unmarshaller for Bot Response
func (br *BotResponse) UnmarshalJSON(data []byte) error {
	// Set Default value of Capture Voice to true
	br.Data.Options.CaptureVoice = true
	br.Data.Options.RecordingBeep = true
	// create alias to prevent endless loop
	type Alias BotResponse
	tmp := (*Alias)(br)
	return json.Unmarshal(data, tmp)
}
