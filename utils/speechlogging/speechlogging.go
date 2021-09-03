package speechlogging

import (
	"os"
	"time"

	"context"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"

	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

// URL for logging endpoint
const URL = "https://staging.yellowmessenger.com/api/sttlog/v2/speech_to_text/log/"

var voiceOTPPhrases = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

// LoggingRequest contains the google API request and response for logging
type LoggingRequest struct {
	CallSID                string                      `json:"call_sid"`
	RecognizeRequest       *speechpb.RecognizeRequest  `json:"recognize_request"`
	RecognizeResponse      *speechpb.RecognizeResponse `json:"recognize_response"`
	LatencyMillis          int64                       `json:"latency_millis"`
	SttService             string                      `json:"stt_service"`
	BotInfo                *botInfo                    `json:"bot_info,omitempty"`
	UserInfo               *userInfo                   `json:"user_info,omitempty"`
	TimeOffsetMillis       int                         `json:"time_offset_millis,omitempty"`
	StreamingRecognizeInfo *streamingRecognizeInfo     `json:"streaming_recognize_info"`
}

// Send sends the stt information to logs backend
func Send(l *LoggingRequest, logURL string) {

	if configmanager.ConfStore != nil && !configmanager.ConfStore.EnableSpeechLogging {
		return
	}

	accountName, accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT"), os.Getenv("AZURE_STORAGE_ACCESS_KEY")

	azureUploader, err := NewAzureUploader(accountName, accountKey, "logging", "logging/stt/recordings")

	if err != nil {
		ymlogger.LogError(l.CallSID, "[InternalLogAPI] Faield to initialize azureUploader")
		return
	}

	logger := NewAudioLogger("http://localhost:3001/structlog.stt", azureUploader)

	result, err := logger.LogAudio(context.Background(), l)

	ymlogger.LogInfof(l.CallSID, "[InternalLogAPI] response: %s, err: %s", result, err)

	return
}

// SetCallSID sets sid info
func (l *LoggingRequest) SetCallSID(callSid string) *LoggingRequest {
	if l == nil {
		return nil
	}

	l.CallSID = callSid
	return l
}

// SetLatencyMillis sets latency info
func (l *LoggingRequest) SetLatencyMillis(latencyMillis int64) *LoggingRequest {
	if l == nil {
		return nil
	}

	l.LatencyMillis = latencyMillis
	return l
}

// SetAudio sets latency audio info
func (l *LoggingRequest) SetAudio(captureOTP bool, languageCode string, data []byte) *LoggingRequest {
	if l == nil {
		return nil
	}

	var sampleRateHertz int32 = 8000
	if configmanager.ConfStore != nil {
		sampleRateHertz = configmanager.ConfStore.STTSampleRate
	}

	recoConfig := &speechpb.RecognitionConfig{
		Encoding:        speechpb.RecognitionConfig_LINEAR16,
		SampleRateHertz: sampleRateHertz,
		LanguageCode:    languageCode,
		Metadata: &speechpb.RecognitionMetadata{
			InteractionType: speechpb.RecognitionMetadata_VOICE_COMMAND,
		},
	}

	if captureOTP {
		speechCont := &speechpb.SpeechContext{
			Phrases: voiceOTPPhrases,
		}
		recoConfig.SpeechContexts = []*speechpb.SpeechContext{
			speechCont,
		}
	}

	l.RecognizeRequest = &speechpb.RecognizeRequest{
		Config: recoConfig,
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{Content: data},
		},
	}
	return l
}

// SetTranscript sets transcript info
func (l *LoggingRequest) SetTranscript(recognizedText string) *LoggingRequest {
	if l == nil {
		return nil
	}

	l.RecognizeResponse = &speechpb.RecognizeResponse{
		Results: []*speechpb.SpeechRecognitionResult{
			{
				Alternatives: []*speechpb.SpeechRecognitionAlternative{
					{
						Transcript: recognizedText,
					},
				},
			},
		},
	}

	return l
}

// SetSttService sets stt service info
func (l *LoggingRequest) SetSttService(service string) *LoggingRequest {
	if l == nil {
		return l
	}

	l.SttService = service
	return l
}

// SetUserID sets user id
func (l *LoggingRequest) SetUserID(phoneNumber string) *LoggingRequest {
	if l == nil {
		return l
	}

	if l.UserInfo == nil {
		l.UserInfo = &userInfo{}
	}

	l.UserInfo.PhoneNumber = phoneNumber

	return l
}

// SetBotID sets bot id
func (l *LoggingRequest) SetBotID(phoneNumber string) *LoggingRequest {
	if l == nil {
		return l
	}

	if l.BotInfo == nil {
		l.BotInfo = &botInfo{}
	}

	l.BotInfo.PhoneNumber = phoneNumber

	return l
}

// SetCallStartTime sets call start time.
func (l *LoggingRequest) SetCallStartTime(startTime time.Time) *LoggingRequest {
	if l == nil {
		return l
	}

	l.TimeOffsetMillis = int(time.Since(startTime).Milliseconds())

	return l
}

// SetStreamingRecognizeInfo sets streaming recognition details
func (l *LoggingRequest) SetStreamingRecognizeInfo(
	streamingAudio []byte,
	chunkIndices []int,
	transcripts []string,
) *LoggingRequest {
	if l == nil {
		return l
	}

	l.StreamingRecognizeInfo = &streamingRecognizeInfo{
		StreamingAudio:          streamingAudio,
		ChunkIndices:            chunkIndices,
		IntermediateTranscripts: transcripts,
	}

	return l
}
