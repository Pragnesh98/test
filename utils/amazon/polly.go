package amazon

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/newrelic"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/helper"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"
)

func GetSpeechFile(
	ctx context.Context,
	channelID string,
	callSID string,
	botID string,
	callerID string,
	text string,
	textType string,
	voiceID string,
) (string, error) {
	fileName := fmt.Sprintf("%x", md5.Sum([]byte(text)))
	if len(voiceID) <= 0 {
		voiceID = configmanager.ConfStore.TTSVoiceID
	}

	mp3File := configmanager.ConfStore.TTSFilePath + fileName + "_" + textType + "_polly_" + voiceID + ".mp3"
	// Check if the file already exists
	if _, err := os.Stat(mp3File); !os.IsNotExist(err) {
		ymlogger.LogInfof(callSID, "File already exists: [%#v]", mp3File)
		return mp3File, err
	}

	outFile, err := os.Create(mp3File)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while creating file. mp3File:[%s], Error: [%#v]", mp3File, err)
		return mp3File, err
	}
	defer outFile.Close()

	// Check for multiple ssml texts
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

	for _, ssml := range botResponseText {
		result, err := GetSpeech(ctx, channelID, callSID, botID, callerID, ssml, textType, voiceID)
		if err != nil {
			ymlogger.LogInfof(callSID, "Failed to recieve output from Amazon polly: [%#v]", err)
			continue
		}

		_, err = io.Copy(outFile, result.AudioStream)
		if err != nil {
			ymlogger.LogErrorf(callSID, "Error while copying input. mp3File:[%s], Error: [%#v]", mp3File, err)
			return mp3File, err
		}
	}

	audioFile, err := helper.ConvertAudioFile(ctx, mp3File)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while converting file. mp3File:[%s] audioFile:[%s] Error: [%#v]", mp3File, audioFile, err)
		return audioFile, err
	}
	return audioFile, nil
}

func GetSpeech(
	ctx context.Context,
	channelID string,
	callSID string,
	botID string,
	callerID string,
	text string,
	textType string,
	voiceID string) (*polly.SynthesizeSpeechOutput, error) {

	// Initialize a session that the SDK uses to load
	// credentials from the shared credentials file. (~/.aws/credentials).
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create Polly client
	svc := polly.New(sess)

	// Outpu to MP3
	input := &polly.SynthesizeSpeechInput{
		OutputFormat: aws.String(configmanager.ConfStore.TTSFileOutputFormat),
		Text:         aws.String(text),
		VoiceId:      aws.String(voiceID),
		SampleRate:   aws.String(strconv.Itoa(configmanager.ConfStore.TTSFrequency)),
	}
	if strings.ToLower(textType) == "ssml" {
		ymlogger.LogDebugf(callSID, "Putting the text type as SSML. TextType:[%s]", textType)
		input.TextType = aws.String(polly.TextTypeSsml)
	}
	// Capture API response time
	sTime := time.Now()
	output, err := svc.SynthesizeSpeech(input)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while synthesizing input. mp3File:[%s], Error: [%#v]", "", err)
		return nil, err
	}
	// Send API response time to newrelic
	if err := newrelic.SendCustomEvent("voice_amazon_tts", map[string]interface{}{
		"campaign_id":   call.GetCampaignID(channelID),
		"response_time": time.Since(sTime).Milliseconds(),
	}); err != nil {
		ymlogger.LogErrorf("NewRelicMetric", "Failed to send voice_amazon_tts metric to newrelic. Error: [%#v]", err)
	}
	// Send TTS Character metric to newrelic
	go helper.SendTTSCharactersMetric(callSID, channelID, "voice_amazon_tts", text, botID, callerID)
	return output, nil
}
