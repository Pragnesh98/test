package eventhandler

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/callstore"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/azure"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/google"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/yellowmessenger"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/helper"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
)

// RecordingFinishedHandler is the handler when we receive RecordingFinished event
func (cH *CallHandlers) RecordingFinishedHandler(
	ctx context.Context,
	recordingName string,
) {
	var sttMode string
	var logSttType string
	sttLatencyInit := time.Now()
	callSID := call.GetSID(cH.ChannelHandler.ID())

	latencyStore := call.GetCallLatencyStore(cH.ChannelHandler.ID())
	latencyStore.AddNewStep(callSID, "unknown")

	ymlogger.LogInfof(callSID, "Recording Finished. Text: [%s]", call.GetTranscript(cH.ChannelHandler.ID()))
	//handler.MOH("default")

	end := cH.ChannelHandler.Subscribe(ari.Events.StasisEnd)
	defer end.Cancel()

	recordingName = call.GetUtteranceFilename(cH.ChannelHandler.ID())

	// This is for non streaming google speech to text API. Uncomment this if needed
	recordingFileName := configmanager.ConfStore.RecordingDirectory + recordingName + "." + configmanager.ConfStore.RecordingFormat

	if call.GetBotOptions(cH.ChannelHandler.ID()) != nil {
		sttMode = call.GetBotOptions(cH.ChannelHandler.ID()).STTMode
	}

	if strings.ToLower(configmanager.ConfStore.STTType) == "streaming" && strings.ToLower(sttMode) == "streaming" {
		sttHandler := call.GetSTTHandler(cH.ChannelHandler.ID())
		if sttHandler != nil {
			ymlogger.LogInfo(callSID, "Closing the Streaming STT Handler")
			sttHandler.Close()
		}
		logSttType = "streaming"
	}

	text := call.GetTranscript(cH.ChannelHandler.ID())
	// detectedLanguage := call.GetDetectedLanguages(cH.ChannelHandler.ID())

	if call.GetAuthenticateUser(cH.ChannelHandler.ID()) && len(call.GetAuthProfileID(cH.ChannelHandler.ID())) > 0 {
		vResp, err := azure.VerifySpeaker(ctx, cH.ChannelHandler.ID(), callSID, recordingFileName)
		if err != nil {
			ymlogger.LogErrorf(callSID, "Failed to get response from the verify user API. Error: [%#v]", err)
		}
		text = vResp.Result
		call.SetTranscript(cH.ChannelHandler.ID(), vResp.Result)
	} else if sttMode != "streaming" {
		logSttType = "static"
		var primarySTTEngine, secondarySTTEngine string
		if call.GetSTTEngine(cH.ChannelHandler.ID()) == "microsoft" {
			primarySTTEngine = "microsoft"
		} else {
			primarySTTEngine = "google"
		}

		if call.GetSTTEngine(cH.ChannelHandler.ID()) == "yellowmessenger" {
			primarySTTEngine = "yellowmessenger"
		}

		botOptions := call.GetBotOptions(cH.ChannelHandler.ID())
		if botOptions != nil && botOptions.SecondarySTTEngine != primarySTTEngine {
			secondarySTTEngine = botOptions.SecondarySTTEngine
		}

		sttLatencyInit := time.Now()
		text = cH.transcribeSpeech(ctx, callSID, recordingName, primarySTTEngine)
		if text == "" {
			ymlogger.LogInfof(callSID, "Failed to get text from primary %s, trying secondary %s", primarySTTEngine, secondarySTTEngine)
			text = cH.transcribeSpeech(ctx, callSID, recordingName, secondarySTTEngine)
		}
		call.SetTranscript(cH.ChannelHandler.ID(), text)

		sttResponseTime := time.Since(sttLatencyInit).Milliseconds()
		if !latencyStore.RecordLatency(callSID, "unknown", callstore.STTResponseTimeinMs, sttResponseTime) {
			ymlogger.LogErrorf(callSID, "Failed to record latency. [STTResponseTimeinMs: %d]", sttResponseTime)
		}
		ymlogger.LogInfof(callSID, "[LatencyParams] Recorded latency. [STTResponseTimeinMs: %d]", sttResponseTime)

	}
	ymlogger.LogInfof(callSID, "Got the response from Speech to text API. Text: [%s]", text)
	call.AddTranscript(cH.ChannelHandler.ID(), text)

	// Don't proceed if DTMF have been captured as DTMF received handler will take care of hitting the bot
	if call.GetDTMFCaptured(cH.ChannelHandler.ID()) {
		ymlogger.LogInfo(callSID, "Setting DTMF Captured to False")
		call.SetDTMFCaptured(cH.ChannelHandler.ID(), false)
		return
	}

	//latency between recordingfinished, and user text sent for processing
	sttLatency := time.Since(sttLatencyInit).Milliseconds()

	ymlogger.LogInfof("Speech to text engine: [%s]; mode: [%s]; Latency(Processtext - recordingFinished): [%v]", call.GetSTTEngine(cH.ChannelHandler.ID()), logSttType, sttLatency)
	helper.SendSTTLatency("STTLatency", call.GetCampaignID(cH.ChannelHandler.ID()), sttLatency, call.GetSTTEngine(strings.ToLower(cH.ChannelHandler.ID())), logSttType)

	recordingFileName = configmanager.ConfStore.RecordingDirectory + recordingName
	ymlogger.LogInfof(callSID, "Removing utterance files. Filename:[%s]", recordingFileName)
	err := os.Remove(recordingFileName + ".wav")
	if err != nil {
		ymlogger.LogErrorf(callSID, "Couldn't remove file: [%#v]", err)
	}
	err = os.Remove(recordingFileName + "." + configmanager.ConfStore.RecordingFormat)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Couldn't remove file: [%#v]", err)
	}
	cH.processUserText(ctx, cH.ChannelHandler.ID(), text, false)

	return
}

func exists(filePath string) (bool, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}
	return fileInfo.Size() > 0, nil
}

func ensureWavFile(dirName, fileName string) (string, error) {
	wavFilePath := path.Join(dirName, fileName) + ".wav"

	fileExists, _ := exists(wavFilePath)
	if fileExists {
		return wavFilePath, nil
	}

	rawFilePath := path.Join(dirName, fileName) + "." + configmanager.ConfStore.RecordingFormat

	fileExists, err := exists(rawFilePath)
	if err != nil {
		return "", err
	}
	if !fileExists {
		return "", fmt.Errorf("Raw file %q not found", rawFilePath)
	}
	return helper.ConvertToWAV8000(rawFilePath)
}

func (cH *CallHandlers) transcribeSpeech(ctx context.Context, callSID, recordingName, sttEngine string) string {
	switch sttEngine {
	case "microsoft":
		recordingFileName, err := ensureWavFile(configmanager.ConfStore.RecordingDirectory, recordingName)
		if err != nil {
			ymlogger.LogErrorf(callSID, "[MicrosoftAPI] Erorr converting to wave file [%#v]", err)
			return ""
		}
		text, err := azure.GetTextFromSpeech(ctx, cH.ChannelHandler.ID(), callSID, recordingFileName)
		if err != nil {
			ymlogger.LogErrorf(callSID, "[MicrosoftAPI]Error while gettting text from speech. Error: [%#v]", err)
		}
		return text
	case "google":
		recordingFileName := configmanager.ConfStore.RecordingDirectory + recordingName + "." + configmanager.ConfStore.RecordingFormat
		text, err := google.GetTextFromSpeech(ctx, cH.ChannelHandler.ID(), callSID, call.GetSTTLanguage(cH.ChannelHandler.ID()), recordingFileName)
		if err != nil {
			ymlogger.LogErrorf(callSID, "Error while gettting text from speech. Error: [%#v]", err)
		}
		return text
	case "yellowmessenger":
		recordingFileName, err := ensureWavFile(configmanager.ConfStore.RecordingDirectory, recordingName)
		if err != nil {
			ymlogger.LogErrorf(callSID, "[YMAPI] Erorr converting to wave file [%#v]", err)
			return ""
		}
		text, err := yellowmessenger.GetTextFromSpeech(ctx, cH.ChannelHandler.ID(), callSID, recordingFileName)
		if err != nil {
			ymlogger.LogErrorf(callSID, "[YMAPI]Error while gettting text from speech. Error: [%#v]", err)
		}
		return text
	default:
		return ""
	}
}
