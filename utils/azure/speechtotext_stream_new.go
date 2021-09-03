package azure

import (
	"bytes"
	"context"
	"io"
	"os"
	"runtime"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/grpc/model"
	pb "bitbucket.org/yellowmessenger/asterisk-ari/utils/grpc/proto"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/grpc/speechtotext"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/handlers"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/helper"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/speechlogging"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
)

func GetStreamTextFromSpeechNew(
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
	recoTypetHandler handlers.RecoTypeHandler,
	detectLanguageCode []string,
) (string, error) {

	ymlogger.LogDebugf(callSID, "[AzureStreamSpeechToText] GoRoutines GoRoutine Started. [%#v]", runtime.NumGoroutine())
	var err error
	var text string
	var file *os.File
	var toContinue bool
	var resultTranscript string
	var messages chan *pb.RecognizeResponse
	var logAudioBuffer bytes.Buffer

	ymlogger.LogInfof(callSID, "[AzureStreamSpeechToText] Creating STTHandler object")

	// transcriptHandler := handlers.GetTranscriptHandler(callSID, channelID, playbackId, interjectUtterances, playbackHandle, recognizeType)
	var sttservice model.SpeechToTextNew = &speechtotext.AzureNew{
		CallSID:               callSID,
		ChannelID:             channelID,
		BoostPhrase:           boostPhrase,
		STTEngine:             sttEngine,
		MsEndpoint:            microsoftEndpoint,
		InitialSilenceTimeout: initialSilenceTimeout,
		FinalSilenceTimeout:   finalSilenceTimeout,
		RecognizeType:         recoTypetHandler.GetMode(),
		RecoTypeHandler:       recoTypetHandler,
		DetectLanguage:        detectLanguageCode,
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
	
	ymlogger.LogInfo(callSID, "[AzureStreamSpeechToText]Recording file successfully opened")
	fstat, err := os.Stat(fileName)
	if err == nil {
		ymlogger.LogInfof(callSID, "fstat start %s - %d - %v", fileName, fstat.Size(), fstat)
	}

	defer file.Close()

	messages, err = sttservice.STTStreamingNew(ctx, io.TeeReader(file, &logAudioBuffer))
	if err != nil {
		ymlogger.LogErrorf(callSID, "[AzureStreamSpeechToText] Couldn't create responnse channel. Error [%v]", err)
		return text, nil
	}
	call.SetSTTHandler(channelID, sttservice)

	for response := range messages {
		ymlogger.LogInfof(callSID, "[AzureStreamSpeechToText] Response from STTHandler: [%#v]", response)
		toContinue, resultTranscript = recoTypetHandler.HandleTranscript(response)
		if !toContinue {
			ymlogger.LogInfo(callSID, "[AzureStreamSpeechToText] Discontinuing listening to messages")
			break
		}
	}

	fstatEnd, err := os.Stat(fileName)
	if err == nil {
		ymlogger.LogInfof(callSID, "fstat end %s - %d - %d - %v", fileName, fstat.Size(), fstatEnd.Size(), fstatEnd)
	}

	ymlogger.LogInfo(callSID, "[AzureStreamSpeechToText] Channel messages Closed")

	// Send request and response for logging
	callStartTime := call.GetPickupTime(channelID)
	botID := call.GetCallerID(channelID).E164Format
	userID := call.GetDialedNumber(channelID).E164Format
	logsRequest := &speechlogging.LoggingRequest{}
	logsRequest.
		SetCallSID(callSID).
		SetAudio(call.GetCaptureVoiceOTP(channelID), call.GetSTTLanguage(channelID), logAudioBuffer.Bytes()).
		SetSttService("Azure").
		SetTranscript(resultTranscript).
		SetBotID(botID).
		SetUserID(userID).
		SetCallStartTime(callStartTime)

	go speechlogging.Send(logsRequest, speechlogging.URL)

	// Send accuracy to new relic
	go helper.SendAccuracyMetric("voice_azure_stt", call.GetCampaignID(channelID), resultTranscript)
	// Send STT Duration metric
	go helper.SendSTTDurationMetric(callSID, channelID, "voice_azure_stt_stream", fileName, call.GetBotID(channelID), call.GetCallerID(channelID).NationalFormat)
	ymlogger.LogDebugf(callSID, "[AzureStreamSpeechToText] GoRoutines GoRoutine Ended. [%#v]", runtime.NumGoroutine())
	return resultTranscript, nil
}
