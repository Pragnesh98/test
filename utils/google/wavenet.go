package google

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/newrelic"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/helper"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

const INPUTSSML = "ssml"
const VOICE_FEMALE = "female"

func GetSpeechFile(
	ctx context.Context,
	channelID string,
	callSID string,
	botID string,
	callerID string,
	text string,
	textType string,
	language string,
	speakRate float64,
	pitch float64,
	voiceID string,
	deviceProfiles []string,
) (string, error) {
	if len(voiceID) <= 0 {
		voiceID = VOICE_FEMALE
	}
	// Check if the file already exists
	fileName := fmt.Sprintf("%x", md5.Sum([]byte(text)))
	filePath := configmanager.ConfStore.TTSFilePath + fileName + "_" + textType + "_" + voiceID + "_" + fmt.Sprintf("%.2f", speakRate) + fmt.Sprintf("%.2f", pitch) + "_wavenet"
	mp3File := filePath + ".mp3"
	trimmedSLNFile := filePath + "_trimmed.sln"
	if _, err := os.Stat(trimmedSLNFile); !os.IsNotExist(err) {
		ymlogger.LogInfof(callSID, "File already exists: [%#v]", trimmedSLNFile)
		return trimmedSLNFile, nil
	}

	botResponseText := []string{text}

	if strings.ToLower(textType) == "ssml" {
		// Check for multiple ssml texts
		SSMLList := helper.GetSSMLList(text)
		if len(SSMLList) == 0 {
			ymlogger.LogInfof(callSID, "No valid SSML text found for TTS")
			return mp3File, errors.New("Invalid SSML text")
		}
		botResponseText = SSMLList
	}

	var finalResponse []byte

	//Get audio bytes for each SSML section
	for _, val := range botResponseText {
		response, err := GetSpeech(ctx, channelID, callSID, botID, callerID, val, textType, language, speakRate, pitch, voiceID, deviceProfiles)
		if err != nil {
			ymlogger.LogErrorf(callSID, "Failed to write the content to the file. Error: [%#v]", err)
			continue
		}
		finalResponse = append(finalResponse, response...)
	}

	err := ioutil.WriteFile(mp3File, finalResponse, 0644)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to write the content to the file. Error: [%#v]", err)
		return trimmedSLNFile, err
	}

	audioFile, err := helper.ConvertAudioFile(ctx, mp3File)
	if err != nil {
		ymlogger.LogInfof(callSID, "Error while converting file. mp3File:[%s] audioFile:[%s], Error: [%#v]", mp3File, audioFile, err)
		return trimmedSLNFile, err
	}
	trimmedSLNFile, err = helper.TrimSilence(ctx, audioFile)
	if err != nil {
		ymlogger.LogInfof(callSID, "Error while trimming the file. mp3File:[%s] TrimmedFile:[%s], Error: [%#v]", mp3File, trimmedSLNFile, err)
		return trimmedSLNFile, err
	}
	return trimmedSLNFile, nil
}

func GetSpeech(
	ctx context.Context,
	channelID string,
	callSID string,
	botID string,
	callerID string,
	text string,
	textType string,
	language string,
	speakRate float64,
	pitch float64,
	voiceID string,
	deviceProfiles []string) ([]byte, error) {

	if speakRate < 0.25 {
		speakRate = 1.0
	}
	// Initialize the client
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to initialize the Google text to speech. Error: [%#v]", err)
		return nil, err
	}
	defer client.Close()

	// Initialize Input
	input := &texttospeechpb.SynthesisInput{
		InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
	}
	if strings.ToLower(textType) == INPUTSSML {
		input = &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Ssml{Ssml: text},
		}
	}

	voiceSelectParams := &texttospeechpb.VoiceSelectionParams{
		LanguageCode: language,
	}
	if voiceID == "male" || voiceID == "female" {
		voiceSelectParams.SsmlGender = texttospeechpb.SsmlVoiceGender_FEMALE
		if strings.ToLower(voiceID) != VOICE_FEMALE {
			voiceSelectParams.SsmlGender = texttospeechpb.SsmlVoiceGender_MALE
		}
	} else {
		voiceSelectParams.Name = voiceID
	}

	// Perform the text-to-speech request on the text input with the selected
	// voice parameters and audio file type.
	req := texttospeechpb.SynthesizeSpeechRequest{
		// Set the text input to be synthesized.
		Input: input,
		// Build the voice request, select the language code ("en-IN") and the SSML
		// voice gender ("female").
		Voice: voiceSelectParams,
		// Select the type of audio file you want returned.
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding:    texttospeechpb.AudioEncoding_MP3,
			SpeakingRate:     speakRate,
			Pitch:            pitch,
			EffectsProfileId: deviceProfiles,
		},
	}

	// Capture API response time
	sTime := time.Now()
	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to Synthesize Speech through Google TTS. Error: [%#v]", err)
		return nil, err
	}
	if err := newrelic.SendCustomEvent("voice_google_tts", map[string]interface{}{
		"campaign_id":   call.GetCampaignID(channelID),
		"response_time": time.Since(sTime).Milliseconds(),
	}); err != nil {
		ymlogger.LogErrorf("NewRelicMetric", "Failed to send voice_google_stt_internal metric to newrelic. Error: [%#v]", err)
	}
	// Send TTS Character metric to newrelic
	go helper.SendTTSCharactersMetric(callSID, channelID, "voice_google_tts", text, botID, callerID)
	return resp.AudioContent, nil
}
