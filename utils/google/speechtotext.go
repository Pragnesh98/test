package google

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/helper"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/speechlogging"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"

	speech "cloud.google.com/go/speech/apiv1"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

var voiceOTPPhrases = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

func GetTextFromSpeech(
	ctx context.Context,
	channelID string,
	callSID string,
	languageCode string,
	fileName string,
) (string, error) {
	sttOutcome := helper.STTOutcome{
		MetricName  :   "STTOutcome", 
		Stt_engine  :   "Google",
		Stt_type    :   "REST",
		Reason      :   "FAILED",
	}

	var text string
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return text, err
	}

	if fileInfo.Size() == 0 {
		return text, errors.New("File size is zero")
	}

	// Creates a client.
	client, err := speech.NewClient(ctx)
	if err != nil {
		return text, err
	}
	defer client.Close()
	sttOutcome.Reason="ClIENT_ERR"
	// Reads the audio file into memory.
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return text, err
	}
	ymlogger.LogDebugf(callSID, "STT Language: [%s]", call.GetSTTLanguage(channelID))

	recoConfig := newRecognitionConfig(channelID, callSID, languageCode)
	recognizeRequest := &speechpb.RecognizeRequest{
		Config: recoConfig,
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{Content: data},
		},
	}
	// Capture API response time
	sTime := time.Now()
	// Detects speech in the audio file.
	resp, err := client.Recognize(ctx, recognizeRequest)
	if err != nil {
		sttOutcome.Reason="REQ_ERROR"
		go helper.SendSTTOutcome(sttOutcome.MetricName, channelID, callSID, sttOutcome.Status, sttOutcome.Reason, sttOutcome.Stt_engine, sttOutcome.Stt_type, 0)
		return text, err
	}
	sttOutcome.Status = "REQ_SUCCESS"
	// Send API response time to newrelic
	go helper.SendResponseTimeMetric("voice_google_stt", call.GetCampaignID(channelID), time.Since(sTime).Milliseconds())

	// Send STT Duration metric to newrelic
	go helper.SendSTTDurationMetric(callSID, channelID, "voice_google_stt", fileName, call.GetBotID(channelID), call.GetCallerID(channelID).NationalFormat)

	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			ymlogger.LogInfof(callSID, "[%v] (confidence=%v)", alt.Transcript, alt.Confidence)
			text = text + alt.Transcript
		}
	}
	
	sttOutcome.Status = "RECOGNIZED"
	if text == "" {
		sttOutcome.Status = "NO_MATCH"
	}

	go helper.SendSTTOutcome(sttOutcome.MetricName, channelID, callSID, sttOutcome.Status, sttOutcome.Reason, sttOutcome.Stt_engine, sttOutcome.Stt_type, 0)
	
	// Send STT Duration metric to newrelic
	go helper.SendSTTDurationMetric(callSID, channelID, "voice_google_stt", fileName, call.GetBotID(channelID), call.GetCallerID(channelID).NationalFormat)

	client.Close()

	// Send request and response for logging
	callStartTime := call.GetPickupTime(channelID)
	botID := call.GetCallerID(channelID).E164Format
	userID := call.GetDialedNumber(channelID).E164Format

	wavBytes, err := helper.ConvertToWAV8000Bytes(data)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to convert to wav. err=%s", err)
	}

	logsRequest := &speechlogging.LoggingRequest{}
	logsRequest.
		SetCallSID(callSID).
		SetAudio(call.GetCaptureVoiceOTP(channelID), call.GetSTTLanguage(channelID), wavBytes).
		SetLatencyMillis(time.Since(sTime).Milliseconds()).
		SetSttService("google").
		SetTranscript(text).
		SetBotID(botID).
		SetUserID(userID).
		SetCallStartTime(callStartTime)

	go speechlogging.Send(logsRequest, speechlogging.URL)
	go helper.SendAccuracyMetric("voice_google_stt", call.GetCampaignID(channelID), text)

	return text, nil
}

func newRecognitionConfig(channelID, callSID, languageCode string) *speechpb.RecognitionConfig {

	var sampleRateHz int32 = 8000
	if configmanager.ConfStore != nil {
		sampleRateHz = configmanager.ConfStore.STTSampleRate
	}
	//Initialize Recognition config
	recoConfig := &speechpb.RecognitionConfig{
		Encoding:        speechpb.RecognitionConfig_LINEAR16,
		SampleRateHertz: sampleRateHz,
		LanguageCode:    languageCode,
		//LanguageCode:    configmanager.ConfStore.STTLanguage,
		Metadata: &speechpb.RecognitionMetadata{
			InteractionType: speechpb.RecognitionMetadata_VOICE_COMMAND,
		},
	}

	botOptions := call.GetBotOptions(channelID)

	var phrases []string

	if botOptions != nil {
		phrases = append(phrases, botOptions.BoostPhrases...)
	}
	if call.GetCaptureVoiceOTP(channelID) {
		phrases = append(phrases, voiceOTPPhrases...)
	}

	if len(phrases) > 0 {
		ymlogger.LogDebugf(callSID, "Setting up the context as Voice OTP. ChannelID: [%s], %s", channelID, phrases)
		speechCont := &speechpb.SpeechContext{
			Phrases: phrases,
		}
		recoConfig.SpeechContexts = []*speechpb.SpeechContext{
			speechCont,
		}
	}

	return recoConfig
}
