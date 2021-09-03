package eventhandler

import (
	"io/ioutil"
	"os"
	"testing"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
)

func TestEnsureWavFile(t *testing.T) {
	f, err := ioutil.TempFile(".", "tmp")
	if err == nil {
		defer os.Remove(f.Name())
	}
	ioutil.WriteFile(f.Name(), []byte("{}"), 0600)
	configmanager.InitConfig(f.Name())
	if configmanager.ConfStore == nil {
		t.Fatal("Confstore is not available")
	}
	configmanager.ConfStore.RecordingDirectory = "."
	recordingName := "test"
	configmanager.ConfStore.RecordingFormat = "raw"
	t.Run("file_already_exists", func(t *testing.T) {
		ioutil.WriteFile("./test.wav", []byte("test"), 0600)
		ensureWavFile(".", recordingName)
		data, _ := ioutil.ReadFile("./test.wav")
		if string(data) != "test" {
			t.Errorf("Expected %q, got %q", "test", string(data))
		}
		os.Remove("./test.wav")
	})
	t.Run("raw_file_doesnt_exist", func(t *testing.T) {
		_, err := ensureWavFile(".", "invalid")
		if err == nil {
			t.Errorf("Expected error while creating wav file")
		}
	})
	t.Run("raw_file_exists_no_wav_file", func(t *testing.T) {
		os.Remove("./recording_test_audio.wav")
		_, err := ensureWavFile(".", "recording_test_audio")
		if err != nil {
			t.Errorf("Expected no error while creating wav file")
		}
		fileExists, err := exists("./recording_test_audio.wav")
		if err != nil {
			t.Errorf("Failed to check file existence")
		}
		if !fileExists {
			t.Errorf("File recording_test_audio.wav wasn't created")
		}
		os.Remove("./recording_test_audio.wav")

	})
}
