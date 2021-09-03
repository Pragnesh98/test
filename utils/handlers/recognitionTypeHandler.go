package handlers

import (
	"context"
	"runtime"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	pb "bitbucket.org/yellowmessenger/asterisk-ari/utils/grpc/proto"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
)

type RecoTypeHandler interface {
	HandleTranscript(transcript *pb.RecognizeResponse) (bool, string)
	GetMode() pb.RecognitionConfig_RecognitionType
	Close() bool
}

func GetInterjectionRecognizer(
	callSID string,
	channelID string,
	playbackID string,
	interjectUtterances []string,
	playbackHandle *ari.PlaybackHandle,
	streamingType string,
) *RecognizeInterject {

	var recogMode = pb.RecognitionConfig_CONTINUOUS
	if strings.ToLower(streamingType) == "once" {
		recogMode = pb.RecognitionConfig_ONCE
	}

	return &RecognizeInterject{
		interjectUtterances: interjectUtterances,
		playbackHandle:      playbackHandle,
		playbackID:          playbackID,
		channelID:           channelID,
		mode:                recogMode,
		callSID:             callSID,
		transcriptProcessed: make(chan bool, 1),
	}
}

func GetStepRecognizer(
	callSID string,
	channelID string,
	playbackID string,
	playbackHandle *ari.PlaybackHandle,
	streamingType string,
	recordHandle *ari.LiveRecordingHandle,
	cancel context.CancelFunc,
) *RecognizeStep {

	var recogMode = pb.RecognitionConfig_CONTINUOUS
	if strings.ToLower(streamingType) == "once" {
		recogMode = pb.RecognitionConfig_ONCE
	}

	return &RecognizeStep{
		playbackHandle:      playbackHandle,
		playbackID:          playbackID,
		channelID:           channelID,
		mode:                recogMode,
		callSID:             callSID,
		transcriptProcessed: make(chan bool, 1),
		recordHandle:        recordHandle,
		cancel:              cancel,
	}

}

// RecognizeStep for normal stt
type RecognizeStep struct {
	playbackHandle      *ari.PlaybackHandle
	recordHandle        *ari.LiveRecordingHandle
	playbackID          string
	channelID           string
	mode                pb.RecognitionConfig_RecognitionType
	callSID             string
	transcriptProcessed chan bool
	cancel              context.CancelFunc
}

// RecognizeInterject mainly for interjection
type RecognizeInterject struct {
	interjectUtterances []string
	playbackHandle      *ari.PlaybackHandle
	playbackID          string
	channelID           string
	mode                pb.RecognitionConfig_RecognitionType
	callSID             string
	transcriptProcessed chan bool
}

func (rc *RecognizeInterject) HandleTranscript(response *pb.RecognizeResponse) (bool, string) {
	if response == nil || len(response.Results) == 0 {
		ymlogger.LogInfof(rc.callSID, "[RecoTypeHandler] Response nil or result empty")
		if rc.playbackHandle != nil {
			ymlogger.LogInfo(rc.callSID, "[RecoTypeHandler] Stopping Playback handler")
			rc.playbackHandle.Stop()
		}
		rc.transcriptProcessed <- true
		return false, ""
	}
	transcripts := response.Results[0].Alternatives
	var text string
	ymlogger.LogInfof(rc.callSID, "[RecoTypeHandler] Started ")

	for _, result := range transcripts {
		if call.GetPlaybackID(rc.channelID) == rc.playbackID {
			ymlogger.LogInfo(rc.callSID, "[RecoTypeHandler] Playback is same. Will break from here")
		}
		text = result.Transcript
		call.SetInterjectedWords(rc.channelID, result.Transcript)
		ymlogger.LogInfof(rc.callSID, "[RecoTypeHandler] Interjected worrds: [%s]", text)
		ymlogger.LogInfof(rc.callSID, "[RecoTypeHandler] InterjectUtterances: [%s]", rc.interjectUtterances)

		if utteranceExists(rc.interjectUtterances, text) {
			ymlogger.LogDebug(rc.callSID, "[RecoTypeHandler] Stream Send: Found user utterance, closing the stream")
			if call.GetPlaybackID(rc.channelID) == rc.playbackID {
				if rc.playbackHandle != nil {
					ymlogger.LogDebug(rc.callSID, "[AzureStreamSpeechToText] Playback stop")
					rc.playbackHandle.Stop()
				}
				call.SetTranscript(rc.channelID, text)
			}
			rc.transcriptProcessed <- true
			ymlogger.LogDebugf("GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
			return false, text
		}
	}
	return true, text
}

func (rc *RecognizeInterject) Close() bool {
	select {
	case <-rc.transcriptProcessed:
		ymlogger.LogInfo(rc.callSID, "[RecoTypeHandler] InterjectTranscript set")
		return true
	default:
		ymlogger.LogInfo(rc.callSID, "[RecoTypeHandler] InterjectTranscript not found")
		return false
	}
}

func (ri *RecognizeInterject) GetMode() pb.RecognitionConfig_RecognitionType {
	return ri.mode
}

func (rc *RecognizeStep) HandleTranscript(response *pb.RecognizeResponse) (bool, string) {
	// languageDetected := response.Detected_Language
	if response == nil || len(response.Results) == 0 {
		ymlogger.LogInfof(rc.callSID, "[RecoTypeHandler] Response nil or result empty")
		if rc.recordHandle != nil {
			ymlogger.LogInfo(rc.callSID, "[RecoTypeHandler] Stopping Record handler")
			rc.recordHandle.Stop()
		}
		rc.transcriptProcessed <- true
		return false, ""
	}
	ymlogger.LogInfof(rc.callSID, "[RecoTypeHandler] Setting detected language : [%s]",response.GetDetected_Language())
	call.SetDetectedLanguage(rc.channelID, response.GetDetected_Language())

	transcripts := response.Results[0].Alternatives
	if transcripts == nil || len(transcripts) == 0 {
		ymlogger.LogInfof(rc.callSID, "[RecoTypeHandler] Transcripts not found or empty")
		if rc.recordHandle != nil {
			ymlogger.LogInfo(rc.callSID, "[RecoTypeHandler] Stopping Record handler")
			rc.recordHandle.Stop()
		}
		rc.transcriptProcessed <- true
		return false, ""
	}
	finalResult := transcripts[0]

	if call.GetPlaybackID(rc.channelID) == rc.playbackID {
		ymlogger.LogInfo(rc.callSID, "[RecoTypeHandler] Playback is same. Will break from here")
	}

	for _, result := range transcripts {
		ymlogger.LogInfof(rc.callSID, "[RecoTypeHandler] [%v] (confidence=%v)", result.Transcript, result.Confidence)

		if result.Confidence > finalResult.Confidence {
			finalResult = result
		}
	}

	ymlogger.LogInfof(rc.callSID, "[RecoTypeHandler] Result words: [%s]. Setting transcript to call.", finalResult.Transcript)
	call.SetTranscript(rc.channelID, finalResult.Transcript)
	if rc.recordHandle != nil {
		ymlogger.LogInfo(rc.callSID, "[RecoTypeHandler] Stopping Record handler")
		rc.recordHandle.Stop()
	}
	rc.transcriptProcessed <- true
	return false, finalResult.Transcript
}

func (rs *RecognizeStep) Close() bool {
	select {
	case <-rs.transcriptProcessed:
		ymlogger.LogInfo(rs.callSID, "[RecoTypeHandler] Transcript set")
		return true
	case <-time.After(time.Duration(configmanager.ConfStore.STTCancelTimeout) * time.Second):
		ymlogger.LogInfo(rs.callSID, "[RecoTypeHandler] Timed out")
		return false
	}
}

func (rs *RecognizeStep) GetMode() pb.RecognitionConfig_RecognitionType {
	return rs.mode
}

func utteranceExists(utterances []string, text string) bool {
	for _, utterance := range utterances {
		if strings.Contains(strings.ToLower(text), strings.ToLower(utterance)) {
			return true
		}
	}
	return false
}
