package eventhandler

import (
	"context"
	"runtime"
	"strings"
	"time"

	//"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	//"bitbucket.org/yellowmessenger/asterisk-ari/utils/google"
	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/callstore"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/azure"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/google"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/handlers"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/vosk"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
)

func (cH *CallHandlers) RecordingStartedHandler(
	ctx context.Context,
	callSID string,
	recordingName string,
) {
	ymlogger.LogInfo(callSID, "Recording Started")
	if cH.ChannelHandler != nil {
		end := cH.ChannelHandler.Subscribe(ari.Events.StasisEnd)
		defer end.Cancel()
	}

	var sttMode string
	if call.GetBotOptions(cH.ChannelHandler.ID()) != nil {
		sttMode = call.GetBotOptions(cH.ChannelHandler.ID()).STTMode
	}

	if strings.ToLower(configmanager.ConfStore.STTType) == "streaming" && strings.ToLower(sttMode) == "streaming" {
		primarySTTEngine := strings.ToLower(call.GetSTTEngine(cH.ChannelHandler.ID()))

		streamCtx, cancel := context.WithCancel(ctx)
		call.SetStreamSTTCancel(cH.ChannelHandler.ID(), cancel)

		go cH.transcribeSpeechFromStream(streamCtx, callSID, primarySTTEngine)
	}
	return
}

func (cH *CallHandlers) transcribeSpeechFromStream(ctx context.Context, callSID, sttEngine string) string {
	var microsoftEndpoint string
	var boostPhrase []string
	var detectLanguageCode []string
	var initialSilenceTimeout, finalSilenceTimeout int32

	initialSilenceTimeout = int32(3 * call.GetBotOptions(cH.ChannelHandler.ID()).RecordingSilenceDuration)
	finalSilenceTimeout = int32(call.GetBotOptions(cH.ChannelHandler.ID()).RecordingSilenceDuration)
	boostPhrase = call.GetBotOptions(cH.ChannelHandler.ID()).BoostPhrases
	streamingType := configmanager.ConfStore.STTStepStreamingType
	latencyStore := call.GetCallLatencyStore(cH.ChannelHandler.ID())
	detectLanguageCode = call.GetBotOptions(cH.ChannelHandler.ID()).DetectLanguageCode

	cancel := call.GetStreamSTTCancel(cH.ChannelHandler.ID())
	recoTypetHandler := handlers.GetStepRecognizer(callSID, cH.ChannelHandler.ID(), cH.PlaybackHandler.ID(), cH.PlaybackHandler, streamingType, cH.RecordHandler, cancel)
	utterranceRecordingFile := call.GetUtteranceFilename(cH.ChannelHandler.ID())
	ymlogger.LogInfof(callSID, "[RecordingStartedHandler] Test [%d] [%d] [%v]", initialSilenceTimeout, finalSilenceTimeout, boostPhrase)

	ymlogger.LogDebugf("[RecordingStartedHandler] GoRoutines", "GoRoutine Started. [%#v]", runtime.NumGoroutine())

	ymlogger.LogInfof(callSID, "[RecordingStartedHandler] Starting transcribe speech with [%s]", sttEngine)

	sttLatencyInit := time.Now()
	switch sttEngine {

	case "microsoft":

		recordingFileName := configmanager.ConfStore.RecordingDirectory + utterranceRecordingFile + ".wav"
		microsoftEndpoint = call.GetBotOptions(cH.ChannelHandler.ID()).MicrosoftSTTOptions.EndpointId

		text, err := azure.GetStreamTextFromSpeechNew(
			ctx,
			cH.ChannelHandler.ID(),
			callSID,
			recordingFileName,
			[]string{},
			cH.PlaybackHandler,
			cH.PlaybackHandler.ID(),
			boostPhrase,
			"azure",
			microsoftEndpoint,
			initialSilenceTimeout,
			finalSilenceTimeout,
			recoTypetHandler,
			detectLanguageCode,
		)
		if err != nil {
			ymlogger.LogErrorf(callSID, "[RecordingStartedHandler ]Error while gettting text from speech. Error: [%#v]", err)
		}
		ymlogger.LogDebugf("[RecordingStartedHandler] GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())

		sttResponseTime := time.Since(sttLatencyInit).Milliseconds()
		if !latencyStore.RecordLatency(callSID, "unknown", callstore.STTResponseTimeinMs, sttResponseTime) {
			ymlogger.LogErrorf(callSID, "Failed to record latency. [STTResponseTimeinMs: %d]", sttResponseTime)
		}
		ymlogger.LogInfo(callSID, "Setting transcribe speech context to nil")
		call.SetStreamSTTCancel(cH.ChannelHandler.ID(), nil)
		return text
	case "vosk":

		// recordingFileName := configmanager.ConfStore.RecordingDirectory + utterranceRecordingFile + ".wav"
		recordingFileName := configmanager.ConfStore.RecordingDirectory + utterranceRecordingFile + "." + configmanager.ConfStore.RecordingFormat

		languageModel := call.GetBotOptions(cH.ChannelHandler.ID()).VoskSTTOptions.LanguageModel

		text, err := vosk.GetStreamTextFromSpeechNew(
			ctx,
			cH.ChannelHandler.ID(),
			callSID,
			recordingFileName,
			[]string{},
			cH.PlaybackHandler,
			cH.PlaybackHandler.ID(),
			boostPhrase,
			"vosk",
			languageModel,
			initialSilenceTimeout,
			finalSilenceTimeout,
			recoTypetHandler,
		)

		if err != nil {
			ymlogger.LogErrorf(callSID, "[RecordingStartedHandler ]Error while gettting text from speech. Error: [%#v]", err)
		}
		ymlogger.LogDebugf("[RecordingStartedHandler] GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
		ymlogger.LogInfof(callSID, "Setting transcribe speech context to nil")
		call.SetStreamSTTCancel(cH.ChannelHandler.ID(), nil)

		return text
	case "google":
		recordingFileName := configmanager.ConfStore.RecordingDirectory + utterranceRecordingFile + "." + configmanager.ConfStore.RecordingFormat
		text, err := google.GetStreamTextFromSpeech(
			ctx,
			cH.ChannelHandler.ID(),
			callSID,
			recordingFileName,
			[]string{},
			cH.PlaybackHandler,
			cH.PlaybackHandler.ID(),
			boostPhrase,
		)
		if err != nil {
			ymlogger.LogErrorf(callSID, "[RecordingStartedHandler] Error while gettting text from speech. Error: [%#v]", err)
		}
		sttResponseTime := time.Since(sttLatencyInit).Milliseconds()
		if !latencyStore.RecordLatency(callSID, "unknown", callstore.STTResponseTimeinMs, sttResponseTime) {
			ymlogger.LogErrorf(callSID, "Failed to record latency. [STTResponseTimeinMs: %d]", sttResponseTime)
		}
		ymlogger.LogInfo(callSID, "Setting transcribe speech context to nil")

		call.SetStreamSTTCancel(cH.ChannelHandler.ID(), nil)

		return text
	default:
		ymlogger.LogErrorf(callSID, "[RecordingStartedHandler]Error while gettting text from speech. Error: STT Engine [%s] not found", sttEngine)
		ymlogger.LogInfo(callSID, "Setting transcribe speech context to nil")
		call.SetStreamSTTCancel(cH.ChannelHandler.ID(), nil)

		return ""
	}
}
