package speechlogging

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	speech "google.golang.org/genproto/googleapis/cloud/speech/v1"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

type fakeUploader struct {
	name string
	data []byte
	err  error
	URL  string
}

func (u *fakeUploader) Upload(ctx context.Context, name string, data []byte) (string, error) {
	u.name = name
	u.data = data
	return u.URL, u.err
}

func (u *fakeUploader) GetInfo() string {
	audioStoreMetadata := audioStoreMetadata{
		azureBlobStorage{
			ContainerName:  "test-container",
			StorageAccount: "test-account",
		},
	}
	result, _ := json.Marshal(audioStoreMetadata)
	return string(result)
}

func TestAudioLogger(t *testing.T) {
	var body []byte
	var contentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = ioutil.ReadAll(r.Body)
		r.Body.Close()
		contentType = r.Header.Get("Content-type")

		w.Write([]byte(""))
	}))

	logger := HTTPLogger{
		URL: server.URL,
	}

	mockUploader := &fakeUploader{}

	audioLogger := logUploader{
		logger:   &logger,
		uploader: mockUploader,
	}

	t.Run("nil_request", func(t *testing.T) {
		logData := LoggingRequest{
			LatencyMillis: 0,
		}
		_, err := audioLogger.LogAudio(context.Background(), &logData)
		if err == nil {
			t.Fatal("Expected err with nil request")
		}
	})

	t.Run("nil_response", func(t *testing.T) {
		body = []byte{}
		contentType = ""
		logData := &LoggingRequest{
			CallSID: "sid",
			RecognizeRequest: &speech.RecognizeRequest{
				Config: &speech.RecognitionConfig{
					Encoding:        speech.RecognitionConfig_LINEAR16,
					LanguageCode:    "en-IN",
					SampleRateHertz: 2500,
					Metadata: &speech.RecognitionMetadata{
						InteractionType: speech.RecognitionMetadata_VOICE_COMMAND,
					},
				},
				Audio: &speechpb.RecognitionAudio{
					AudioSource: &speechpb.RecognitionAudio_Content{
						Content: []byte("FooBar"),
					},
				},
			},
			LatencyMillis: 0,
			SttService:    "google",
		}
		_, err := audioLogger.LogAudio(context.Background(), logData)
		if err != nil {
			t.Fatal(err)
		}

		if contentType != "application/json" {
			t.Fatal("application/json content-type header expected")
		}
		var sttLog sttLog
		err = json.Unmarshal(body, &sttLog)
		if err != nil {
			t.Fatalf("Failed to parse json log %s", err)
		}

		assertStringEquals(sttLog.CallSid, "sid", t)
		assertStringEquals(sttLog.AudioURI, mockUploader.URL, t)
		assertIntEquals(int(sttLog.AudioSizeBytes), len("FooBar"), t)
		assertStringEquals(sttLog.SttRequest.Encoding, "LINEAR16", t)
		assertStringEquals(sttLog.SttRequest.InteractionType, "VOICE_COMMAND", t)
		assertStringEquals(sttLog.SttRequest.LanguageCode, "en-IN", t)
		assertIntEquals(int(sttLog.SttRequest.SampleRateHertz), 2500, t)
	})

	t.Run("audio_uploaded", func(t *testing.T) {
		body = []byte{}
		contentType = ""

		logData := LoggingRequest{
			RecognizeRequest: &speech.RecognizeRequest{
				Audio: &speechpb.RecognitionAudio{
					AudioSource: &speechpb.RecognitionAudio_Content{
						Content: []byte("TestAudio"),
					},
				},
			},
			LatencyMillis: 0,
			SttService:    "google",
		}
		audioLogger.LogAudio(context.Background(), &logData)
		assertStringEquals(string(mockUploader.data), string([]byte("TestAudio")), t)
		md5Hex := md5.New()
		md5Hex.Write(b64Encoded([]byte("TestAudio")))
		assertStringEquals(mockUploader.name, hex.EncodeToString(md5Hex.Sum(nil)), t)

		var sttLog sttLog
		json.Unmarshal(body, &sttLog)
		assertStringEquals(sttLog.AudioMd5Hex, hex.EncodeToString(md5Hex.Sum(nil)), t)
	})

	t.Run("check_sttlog_empty_resp", func(t *testing.T) {
		body = []byte{}
		contentType = ""
		logData := &LoggingRequest{
			RecognizeRequest: &speech.RecognizeRequest{
				Audio: &speechpb.RecognitionAudio{
					AudioSource: &speechpb.RecognitionAudio_Content{
						Content: []byte{},
					},
				},
			},
			RecognizeResponse: &speech.RecognizeResponse{},
			LatencyMillis:     0,
			SttService:        "google",
		}
		audioLogger.LogAudio(context.Background(), logData)
		var sttLog sttLog
		json.Unmarshal(body, &sttLog)
		assertIntEquals(int(sttLog.SttResponse.ResultCount), 0, t)
	})

	t.Run("check_sttlog_storage_metadata", func(t *testing.T) {
		body = []byte{}
		contentType = ""
		logData := &LoggingRequest{
			RecognizeRequest: &speech.RecognizeRequest{
				Audio: &speechpb.RecognitionAudio{
					AudioSource: &speechpb.RecognitionAudio_Content{
						Content: []byte{},
					},
				},
			},
			LatencyMillis: 100,
			SttService:    "google",
		}
		audioLogger.LogAudio(context.Background(), logData)
		var sttLog sttLog
		json.Unmarshal(body, &sttLog)

		var audioStoreMetadata audioStoreMetadata
		json.Unmarshal([]byte(sttLog.AudioStoreMetadata), &audioStoreMetadata)
		assertStringEquals(audioStoreMetadata.AzureBlobStorage.ContainerName, "test-container", t)
		assertStringEquals(audioStoreMetadata.AzureBlobStorage.StorageAccount, "test-account", t)
	})

	t.Run("check_sttlog_response", func(t *testing.T) {
		res := &response{
			RecognizeResponse: &speech.RecognizeResponse{
				Results: []*speech.SpeechRecognitionResult{
					{
						Alternatives: []*speech.SpeechRecognitionAlternative{
							{
								Transcript: "Hello",
								Confidence: 0.2,
							},
						},
					},
				},
			},
		}
		body = []byte{}
		contentType = ""
		logData := &LoggingRequest{
			RecognizeRequest: &speech.RecognizeRequest{
				Audio: &speechpb.RecognitionAudio{
					AudioSource: &speechpb.RecognitionAudio_Content{
						Content: []byte{},
					},
				},
			},
			RecognizeResponse: res.RecognizeResponse,
			LatencyMillis:     0,
			SttService:        "google",
		}
		audioLogger.LogAudio(context.Background(), logData)
		var sttLog sttLog
		json.Unmarshal(body, &sttLog)
		assertIntEquals(int(sttLog.SttResponse.ResultCount), 1, t)
		assertStringEquals(sttLog.SttResponse.Transcript, "Hello", t)
		assertFloatEquals(sttLog.SttResponse.Confidence, 0.2, t)
	})

	t.Run("check_sttlog_duration", func(t *testing.T) {
		body = []byte{}
		contentType = ""
		logData := &LoggingRequest{
			RecognizeRequest: &speech.RecognizeRequest{
				Audio: &speechpb.RecognitionAudio{
					AudioSource: &speechpb.RecognitionAudio_Content{
						Content: []byte{},
					},
				},
			},
			LatencyMillis: 100,
			SttService:    "google",
		}
		audioLogger.LogAudio(context.Background(), logData)
		var sttLog sttLog
		json.Unmarshal(body, &sttLog)

		assertIntEquals(int(sttLog.SttResponse.DecodingLatencyMillis), 100, t)
	})
}

func b64Encoded(data []byte) []byte {
	var b bytes.Buffer
	e := base64.NewEncoder(base64.StdEncoding, &b)
	e.Write(data)
	return b.Bytes()
}

func assertConfigEquals(got *speechpb.RecognitionConfig, want *speechpb.RecognitionConfig, t *testing.T) {
	t.Helper()

	if got == want {
		return
	}

	if got == nil {
		t.Fatalf("want %q, got nil", want)
	}

	assertStringEquals(got.LanguageCode, want.LanguageCode, t)
	assertStringEquals(got.Encoding.String(), want.Encoding.String(), t)
	assertIntEquals(int(got.SampleRateHertz), int(want.SampleRateHertz), t)
}

func assertNoError(err error, t *testing.T) {
	t.Helper()

	if err != nil {
		t.Fatalf("Expected no error but found %s", err)
	}
}

func assertIntEquals(got int, want int, t *testing.T) {
	t.Helper()

	if got != want {
		t.Fatalf("Wanted %d, got %d\n", want, got)
	}
}

func assertStringEquals(got string, want string, t *testing.T) {
	t.Helper()

	if got != want {
		t.Errorf("Wanted %q, got %q\n", want, got)
	}
}

func assertFloatEquals(got float32, want float32, t *testing.T) {
	t.Helper()

	if got != want {
		t.Errorf("Wanted %f, got %f\n", want, got)
	}
}
