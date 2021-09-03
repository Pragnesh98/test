package azure

import (
	"context"
	"os"
	"runtime"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	pb "bitbucket.org/yellowmessenger/asterisk-ari/utils/grpc/proto"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/grpc/speechtotext"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/helper"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/speechlogging"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
)

func GetStreamTextFromSpeech(
	ctx context.Context,
	channelID string,
	callSID string,
	fileName string,
	interjectUtterances []string,
	playbackHandle *ari.PlaybackHandle,
	playbackId string,
	boostPhrase []string,
	sttEngine string,
	microsoftEndpoint string,
	initialSilenceTimeout int32,
	finalSilenceTimeout int32,
	recognizeType pb.RecognitionConfig_RecognitionType,
) (string, error) {
	ymlogger.LogDebugf("GoRoutines", "GoRoutine Started. [%#v]", runtime.NumGoroutine())
	var text string
	var file *os.File
	var err error
	var messages chan *pb.RecognizeResponse
	var sttservice speechtotext.SpeechToText = speechtotext.Azure{
		CallSID:               callSID,
		ChannelID:             channelID,
		BoostPhrase:           boostPhrase,
		STTEngine:             sttEngine,
		MsEndpoint:            microsoftEndpoint,
		InitialSilenceTimeout: initialSilenceTimeout,
		FinalSilenceTimeout:   finalSilenceTimeout,
		RecognizeType:         recognizeType,
	}

	startTime := time.Now()
	for {
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			if time.Since(startTime).Milliseconds() > 6000 {
				ymlogger.LogErrorf(callSID, "[AzureStreamSpeechToText] Error while opening the file. Error: [%#v]", err)
				break
			}
			continue
		}
		break
	}

	for {
		file, err = os.Open(fileName)
		if err != nil {
			if time.Since(startTime).Milliseconds() > 6000 {
				ymlogger.LogErrorf(callSID, "[AzureStreamSpeechToText] Error while opening the file. Error: [%#v]", err)
				break
			}
			ymlogger.LogErrorf(callSID, "[AzureStreamSpeechToText] Error while opening the file. Error: [%#v]", err)
			continue
		}
		break
	}
	defer file.Close()

	// Connection to stt service
	messages, done, err := sttservice.STTStreaming(ctx, file)
	if err != nil {
		ymlogger.LogErrorf(callSID, "[AzureStreamSpeechToText] Couldn't create responnse channel. Error [%v]", err)
		return text, nil
	}

	var resultTranscript string
	var words []string
	for response := range messages {
		if response == nil {
			continue
		}
		for _, result := range response.Results[0].Alternatives {
			if call.GetPlaybackID(channelID) == playbackId {
				ymlogger.LogInfo(callSID, "[AzureStreamSpeechToText] Playback is same. Will break from here")
			}
			text = result.Transcript
			call.SetInterjectedWords(channelID, result.Transcript)
			words = append(words, result.Transcript)
			if utteranceExists(interjectUtterances, result.Transcript) {
				ymlogger.LogDebug(callSID, "[AzureStreamSpeechToText] Stream Send: Found user utterance, closing the stream")
				if call.GetPlaybackID(channelID) == playbackId {
					if playbackHandle != nil {
						// ymlogger.LogDebug(callSID, "[AzureStreamSpeechToText] Mocking the playback stop")
						playbackHandle.Stop()
					}
					call.SetTranscript(channelID, text)
				}
				done <- true
				ymlogger.LogDebugf("GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
				return text, nil
			}
		}
		//Non-interjection case
		//resultTranscript, err = getResultTextFromSTTResponse(callSID, response)
	}

	// Send request and response for logging
	callStartTime := call.GetPickupTime(channelID)
	botID := call.GetCallerID(channelID).E164Format
	userID := call.GetDialedNumber(channelID).E164Format

	logsRequest := &speechlogging.LoggingRequest{}
	logsRequest.
		SetCallSID(callSID).
		SetSttService("Microsoft SDK").
		SetTranscript(resultTranscript).
		SetBotID(botID).
		SetUserID(userID).
		SetCallStartTime(callStartTime)

	go speechlogging.Send(logsRequest, speechlogging.URL)

	// Send accuracy to new relic
	go helper.SendAccuracyMetric("voice_azure_stt", call.GetCampaignID(channelID), resultTranscript)
	// Send STT Duration metric
	go helper.SendSTTDurationMetric(callSID, channelID, "voice_azure_stt_stream", fileName, call.GetBotID(channelID), call.GetCallerID(channelID).NationalFormat)

	ymlogger.LogDebugf("GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
	return resultTranscript, nil
}

func getResultTextFromSTTResponse(callSID string, respData *pb.RecognizeResponse) (string, error) {
	var text string
	var err error
	sttResponse := respData
	if len(sttResponse.Results) == 0 || len(sttResponse.Results[0].Alternatives) == 0 {
		ymlogger.LogInfof(callSID, "[AzureStreamSpeechToText] Empty transcript from results from Azure SDK")
		return text, err
	}

	finalResult := sttResponse.Results[0].Alternatives[0]
	results := sttResponse.Results[0].Alternatives

	// Obtain result with maximum confidence
	for _, result := range results {
		ymlogger.LogInfof(callSID, "[%v] (confidence=%v)", result.Transcript, result.Confidence)

		if result.Confidence > finalResult.Confidence {
			finalResult = result
		}
	}
	return finalResult.Transcript, nil
}

func utteranceExists(utterances []string, text string) bool {
	for _, utterance := range utterances {
		if strings.Contains(strings.ToLower(text), strings.ToLower(utterance)) {
			return true
		}
	}
	return false
}
