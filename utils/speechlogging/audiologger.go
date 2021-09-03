package speechlogging

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

type request struct {
	CallSid          string           `json:"callSid"`
	RecognizeRequest recognizeRequest `json:"recognizeRequest"`
}

type response struct {
	RecognizeResponse *speechpb.RecognizeResponse `json:"recognizeResponse"`
}

type recognizeRequest struct {
	Audio struct {
		AudioSource struct {
			Content []byte `json:"Content"`
		} `json:"AudioSource"`
	} `json:"audio"`
	Config *speechpb.RecognitionConfig `json:"config"`
}

type sttRequest struct {
	Encoding        string `json:"encoding"`
	SampleRateHertz int    `json:"sample_rate_hertz"`
	LanguageCode    string `json:"language_code"`
	InteractionType string `json:"interaction_type"`
}

type sttResponse struct {
	Transcript            string  `json:"transcript"`
	Confidence            float32 `json:"confidence"`
	DecodingLatencyMillis int     `json:"decoding_latency_millis"`
	ResultCount           int     `json:"result_count"`
}

type userInfo struct {
	// Phone number of the user in this conversation
	PhoneNumber string `json:"phone_number,omitempty"`
}

type botInfo struct {
	// Phone number of the bot handling this conversation
	PhoneNumber string `json:"phone_number,omitempty"`
}

type sttLog struct {
	CallSid         string `json:"call_sid"`
	TimestampMicros int64  `json:"timestamp_micros"`

	// Offset of this clip in the session
	TimeOffsetMillis           int          `json:"time_offset_milliss"`
	AudioMd5Hex                string       `json:"audio_md5_hex"`
	AudioSizeBytes             int32        `json:"audio_size_bytes"`
	AudioURI                   string       `json:"audio_uri"`
	AudioStoreMetadata         string       `json:"audio_store_metadata"`
	SttRequest                 *sttRequest  `json:"stt_request,omitempty"`
	SttResponse                *sttResponse `json:"stt_response,omitempty"`
	SttService                 string       `json:"stt_service"`
	UserInfo                   *userInfo    `json:"user_info,omitempty"`
	BotInfo                    *botInfo     `json:"bot_info,omitempty"`
	SttMetadata                string       `json:"stt_metadata,omitempty"`
	StreamingAudioChunkIndices []int        `json:"streaming_audio_chunk_indices,omitempty"`
	StreamingAudioTranscripts  []string     `json:"streaming_audio_transcripts,omitempty"`
	StreamingAudioBlobName     string       `json:"streaming_audio_blob_name,omitempty"`
}

// logUploader uploads the blobs to blobstore and logs details to logger
type logUploader struct {
	logger   logger
	uploader uploader
}

type uploader interface {
	Upload(ctx context.Context, name string, data []byte) (string, error)
	GetInfo() string
}

// NewAudioLogger returns an implementation that uploads audio to a store and logs a metadata to logger.
func NewAudioLogger(URL string, uploader uploader) *logUploader {
	return &logUploader{
		logger: &HTTPLogger{
			URL: URL,
		},
		uploader: uploader,
	}
}

type logger interface {
	Log(ctx context.Context, data []byte) error
}

// HTTPLogger stores http endpoint for log uploads
type HTTPLogger struct {
	URL string
}

// Log uploads requests to an http server
func (l *HTTPLogger) Log(ctx context.Context, data []byte) error {
	client := &http.Client{}
	_, err := client.Post(l.URL, "application/json", bytes.NewReader(data))
	return err
}

type streamingRecognizeInfo struct {
	StreamingAudio          []byte   `json:"streaming_audio"`
	ChunkIndices            []int    `json:"chunk_indices"`
	IntermediateTranscripts []string `json:"intermediate_transcript"`
}

type logData struct {
	CallSid                string                      `json:"call_sid"`
	RecognizeRequest       *recognizeRequest           `json:"recognize_request"`
	RecognizeResponse      *speechpb.RecognizeResponse `json:"recognize_response"`
	LatencyMillis          int                         `json:"latency_millis"`
	TimeOffsetMillis       int                         `json:"time_offset_millis"`
	UserInfo               *userInfo                   `json:"user_info,omitempty"`
	BotInfo                *botInfo                    `json:"bot_info,omitempty"`
	SttService             string                      `json:"stt_service,omitempty"`
	SttMetadata            string                      `json:"stt_metadata,omitempty"`
	StreamingRecognizeInfo *streamingRecognizeInfo     `json:"streaming_recognize_info"`
}

func (l *logUploader) uploadBlob(ctx context.Context, audio []byte) (name string, URL string) {

	var encoded bytes.Buffer
	e := base64.NewEncoder(base64.StdEncoding, &encoded)
	e.Write(audio)

	h := md5.New()
	h.Write(encoded.Bytes())
	name = hex.EncodeToString(h.Sum(nil))
	URL, err := l.uploader.Upload(ctx, name, audio)

	if err != nil {
		ymlogger.LogErrorf("", "Failed to upload audio %s\n", err)
	} else {
		ymlogger.LogInfof("", "Failed to upload audio %s\n", err)

	}
	return
}

// LogAudio uploads the audio to blobstore and logs request/response metadata to a logger
func (l *logUploader) LogAudio(ctx context.Context, logData *LoggingRequest) ([]byte, error) {
	if logData.RecognizeRequest == nil {
		log.Println("recognize request is nil")
		return nil, fmt.Errorf("recognizeRequest must not be nil")
	}

	audio := logData.RecognizeRequest.Audio.GetContent()
	var name, URL, streamingAudioName string

	if len(audio) > 0 {
		name, URL = l.uploadBlob(ctx, audio)
	}

	var chunkIndices []int
	var streamingTranscripts []string
	if logData.StreamingRecognizeInfo != nil {
		streamingAudioName, _ = l.uploadBlob(ctx, logData.StreamingRecognizeInfo.StreamingAudio)
		chunkIndices = logData.StreamingRecognizeInfo.ChunkIndices
		streamingTranscripts = logData.StreamingRecognizeInfo.IntermediateTranscripts
	}

	audioLog := sttLog{
		CallSid:                    logData.CallSID,
		TimestampMicros:            time.Now().Unix() * time.Second.Microseconds(),
		TimeOffsetMillis:           logData.TimeOffsetMillis,
		AudioMd5Hex:                name,
		AudioSizeBytes:             int32(len(audio)),
		AudioURI:                   URL,
		AudioStoreMetadata:         l.uploader.GetInfo(),
		UserInfo:                   logData.UserInfo,
		BotInfo:                    logData.BotInfo,
		SttService:                 logData.SttService,
		StreamingAudioBlobName:     streamingAudioName,
		StreamingAudioChunkIndices: chunkIndices,
		StreamingAudioTranscripts:  streamingTranscripts,
	}

	if config := logData.RecognizeRequest.Config; config != nil {
		audioLog.SttRequest = &sttRequest{
			Encoding:        config.Encoding.String(),
			LanguageCode:    config.LanguageCode,
			SampleRateHertz: int(config.SampleRateHertz),
		}

		if config.Metadata != nil {
			audioLog.SttRequest.InteractionType = config.Metadata.InteractionType.String()
		}
	}

	audioLog.SttResponse = &sttResponse{
		DecodingLatencyMillis: int(logData.LatencyMillis),
	}
	if logData.RecognizeResponse != nil {
		gRes := logData.RecognizeResponse
		if len(gRes.Results) > 0 && len(gRes.Results[0].Alternatives) > 0 {
			audioLog.SttResponse.ResultCount = len(gRes.Results)
			audioLog.SttResponse.Transcript = gRes.Results[0].Alternatives[0].Transcript
			audioLog.SttResponse.Confidence = gRes.Results[0].Alternatives[0].Confidence
		}
	}

	jsonLog, err := json.Marshal(audioLog)

	if err != nil {
		ymlogger.LogErrorf(logData.CallSID, "Unable to serialize log %s, %v", err, audioLog)
		return nil, err
	}

	err = l.logger.Log(ctx, jsonLog)

	if err != nil {
		ymlogger.LogErrorf(logData.CallSID, "Failed to log data err=%s, data=%s", err, jsonLog)
	}
	return jsonLog, err
}
