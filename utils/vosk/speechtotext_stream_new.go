package vosk

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
	languageModel string,
	initialSilenceTimeout int32,
	finalSilenceTimeout int32,
	recoTypetHandler handlers.RecoTypeHandler,
) (string, error) {

	ymlogger.LogDebugf(callSID, "[VoskStreamSpeechToText] GoRoutines GoRoutine Started. [%#v]", runtime.NumGoroutine())
	var err error
	var text string
	var file *os.File
	var toContinue bool
	var resultTranscript string
	var messages chan *pb.RecognizeResponse

	ymlogger.LogInfo(callSID, "[VoskStreamSpeechToText] Creating STTHandler object")

	// transcriptHandler := handlers.GetTranscriptHandler(callSID, channelID, playbackId, interjectUtterances, playbackHandle, recognizeType)
	var sttservice model.SpeechToTextNew = &speechtotext.VoskSTT{
		CallSID:               callSID,
		ChannelID:             channelID,
		BoostPhrase:           boostPhrase,
		STTEngine:             sttEngine,
		InitialSilenceTimeout: initialSilenceTimeout,
		FinalSilenceTimeout:   finalSilenceTimeout,
		RecognizeType:         recoTypetHandler.GetMode(),
		RecoTypeHandler:       recoTypetHandler,
		Model:                 languageModel,
	}
	ymlogger.LogInfo(callSID, "[VoskStreamSpeechToText] Setting STTHandler object")

	call.SetSTTHandler(channelID, sttservice)

	startTime := time.Now()
	for {
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			if time.Since(startTime).Milliseconds() > 6000 {
				ymlogger.LogErrorf(callSID, "[VoskStreamSpeechToText] Error while opening the file. Error: [%#v]", err)
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
				ymlogger.LogErrorf(callSID, "[VoskStreamSpeechToText] Error while opening the file. Error: [%#v]", err)
				break
			}
			ymlogger.LogErrorf(callSID, "[VoskStreamSpeechToText] Error while opening the file. Error: [%#v]", err)
			continue
		}
		break
	}
	ymlogger.LogInfo(callSID, "[VoskStreamSpeechToText]Recording file successfully opened")

	defer file.Close()
	messages, err = sttservice.STTStreamingNew(ctx, file)
	if err != nil {
		ymlogger.LogErrorf(callSID, "[VoskStreamSpeechToText] Couldn't create responnse channel. Error [%v]", err)
		return text, nil
	}

	for response := range messages {
		ymlogger.LogInfof(callSID, "[VoskStreamSpeechToText] Response from STTHandler: [%#v]", response)
		toContinue, resultTranscript = recoTypetHandler.HandleTranscript(response)
		if !toContinue {
			ymlogger.LogInfo(callSID, "[VoskStreamSpeechToText] Discontinuing listening to messages")
			break
		}
	}
	ymlogger.LogInfo(callSID, "[VoskStreamSpeechToText] Channel messages Closed")

	// Send request and response for logging
	callStartTime := call.GetPickupTime(channelID)
	botID := call.GetCallerID(channelID).E164Format
	userID := call.GetDialedNumber(channelID).E164Format
	audioBuf := bytes.NewBuffer(nil)
	io.Copy(audioBuf, file)

	logsRequest := &speechlogging.LoggingRequest{}
	logsRequest.
		SetCallSID(callSID).
		SetSttService("Microsoft SDK").
		SetTranscript(resultTranscript).
		SetBotID(botID).
		SetUserID(userID).
		SetCallStartTime(callStartTime).
		SetStreamingRecognizeInfo(audioBuf.Bytes(), []int{}, []string{})

	go speechlogging.Send(logsRequest, speechlogging.URL)

	// Send accuracy to new relic
	go helper.SendAccuracyMetric("voice_azure_stt", call.GetCampaignID(channelID), resultTranscript)

	ymlogger.LogDebugf(callSID, "[VoskStreamSpeechToText] GoRoutines GoRoutine Ended. [%#v]", runtime.NumGoroutine())
	return resultTranscript, nil
}
