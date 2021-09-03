package eventhandler

import (
	"context"
	"runtime"
	"strings"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/azure"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/google"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/handlers"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
	guuid "github.com/google/uuid"
)

// PlaybackStartedHandler is the handler when we receive PlaybackStarted event
func (cH *CallHandlers) PlaybackStartedHandler(ctx context.Context) {
	callSID := call.GetSID(cH.ChannelHandler.ID())
	ymlogger.LogInfo(callSID, "Playback Started")
	call.SetPlaybackID(cH.ChannelHandler.ID(), cH.PlaybackHandler.ID())
	if cH.ChannelHandler != nil {
		end := cH.ChannelHandler.Subscribe(ari.Events.StasisEnd)
		defer end.Cancel()
	}
	if call.GetBotOptions(cH.ChannelHandler.ID()) != nil && call.GetBotOptions(cH.ChannelHandler.ID()).Interject {
		// Set Parent Unique ID for Snoop Channel
		call.SetParentUniqueID("inter"+callSID, cH.ChannelHandler.ID())
		// Set Recording File Name for user interjections.
		fileName := guuid.New().String()
		call.SetInterRecordingFilename(cH.ChannelHandler.ID(), fileName)

		// Start snooping the channel for user interjections.
		if cH.ChannelHandler != nil {
			var err error
			for i := 0; i < configmanager.ConfStore.ARIAPIRetry; i++ {
				cH.InterSnoopHandler, err = cH.ChannelHandler.Snoop("inter"+callSID, &ari.SnoopOptions{
					Spy: "in",
					App: "hello-world",
				})
				if err != nil {
					ymlogger.LogErrorf(callSID, "Error while snooping the call. Error: [%#v]. Retrying.....", err)
					continue
				}
				break
			}
			if err != nil {
				ymlogger.LogErrorf(callSID, "Error while snooping the call. Error: [%#v]", err)
				cH.ChannelHandler.Hangup()
				return
			}
		}
		sttEngine := strings.ToLower(call.GetSTTEngine(cH.ChannelHandler.ID()))
		streamCtx, cancel := context.WithCancel(ctx)
		call.SetStreamSTTCancel(cH.ChannelHandler.ID(), cancel)

		go cH.transcribeInterjection(streamCtx, callSID, fileName, sttEngine)
	}
	return
}

func (cH *CallHandlers) transcribeInterjection(ctx context.Context, callSID, recordingName, sttEngine string) string {
	var microsoftEndpoint string
	var boostPhrase []string
	var initialSilenceTimeout, finalSilenceTimeout int32

	finalSilenceTimeout = 0
	initialSilenceTimeout = 0
	streamingType := configmanager.ConfStore.STTInterjectStreamingType
	boostPhrase = call.GetBotOptions(cH.ChannelHandler.ID()).BoostPhrases
	interjectUtterances := call.GetBotOptions(cH.ChannelHandler.ID()).InterjectUtterances
	microsoftEndpoint = call.GetBotOptions(cH.ChannelHandler.ID()).MicrosoftSTTOptions.EndpointId
	recoTypetHandler := handlers.GetInterjectionRecognizer(callSID, cH.ChannelHandler.ID(), cH.PlaybackHandler.ID(), interjectUtterances, cH.PlaybackHandler, streamingType)
	detectLanguageCode := call.GetBotOptions(cH.ChannelHandler.ID()).DetectLanguageCode

	ymlogger.LogDebugf("[PlaybackStartedHandler] GoRoutines", "GoRoutine Started. [%#v]", runtime.NumGoroutine())
	ymlogger.LogInfof(callSID, "[PlaybackStartedHandler] Starting transcribe speech with [%s]", sttEngine)

	switch sttEngine {

	case "microsoft":

		recordingFileName := configmanager.ConfStore.RecordingDirectory + recordingName + ".wav"

		text, err := azure.GetStreamTextFromSpeechNew(
			ctx,
			cH.ChannelHandler.ID(),
			callSID,
			recordingFileName,
			interjectUtterances,
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
			ymlogger.LogErrorf(callSID, "[PlaybackStartedHandler ]Error while gettting text from speech. Error: [%#v]", err)
		}
		ymlogger.LogDebugf("[PlaybackStartedHandler] GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())

		return text
	case "vosk":

		recordingFileName := configmanager.ConfStore.RecordingDirectory + recordingName + ".wav"

		text, err := azure.GetStreamTextFromSpeechNew(
			ctx,
			cH.ChannelHandler.ID(),
			callSID,
			recordingFileName,
			interjectUtterances,
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
			ymlogger.LogErrorf(callSID, "[PlaybackStartedHandler ]Error while gettting text from speech. Error: [%#v]", err)
		}
		ymlogger.LogDebugf("[PlaybackStartedHandler] GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())

		return text

	case "google":
		recordingFileName := configmanager.ConfStore.RecordingDirectory + recordingName + "." + configmanager.ConfStore.RecordingFormat
		text, err := google.GetStreamTextFromSpeech(
			ctx,
			cH.ChannelHandler.ID(),
			callSID,
			recordingFileName,
			interjectUtterances,
			cH.PlaybackHandler,
			cH.PlaybackHandler.ID(),
			boostPhrase,
		)
		if err != nil {
			ymlogger.LogErrorf(callSID, "[PlaybackStartedHandler] Error while gettting text from speech. Error: [%#v]", err)
		}
		return text
	default:
		ymlogger.LogErrorf(callSID, "[PlaybackStartedHandler]Error while gettting text from speech. Error: STT Engine [%s] not found", sttEngine)
		return ""
	}
}
