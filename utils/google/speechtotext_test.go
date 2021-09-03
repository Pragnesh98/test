package google

import (
	"reflect"
	"testing"

	"bitbucket.org/yellowmessenger/asterisk-ari/bothelper"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
)

func TestSpeechContext(t *testing.T) {
	t.Run("no_voice_otp", func(t *testing.T) {
		config := newRecognitionConfig("123", "12312", "en")
		if len(config.SpeechContexts) != 0 {
			t.Error("Expected no speech context to be passed")
		}
	})

	t.Run("voice_otp_only", func(t *testing.T) {
		call.SetCaptureVoiceOTP("channelID", true)
		config := newRecognitionConfig("channelID", "callSID", "en")
		if len(config.SpeechContexts) != 1 {
			t.Fatalf("Want speechcontext length=%d, got=%d", 1, len(config.SpeechContexts))
		}
		if !reflect.DeepEqual(config.SpeechContexts[0].Phrases, voiceOTPPhrases) {
			t.Errorf("Expected voice otp phrases to be passed to speech context")
		}
	})

	t.Run("boost_phrases_only", func(t *testing.T) {
		botOptions := bothelper.BotOptions{
			BoostPhrases: []string{"test1", "test2"},
		}
		call.SetBotOptions("channel2", &botOptions)
		config := newRecognitionConfig("channel2", "callSID", "en")
		if len(config.SpeechContexts) != 1 {
			t.Fatalf("Want speechcontext length=%d, got=%d", 1, len(config.SpeechContexts))
		}
		if !reflect.DeepEqual(config.SpeechContexts[0].Phrases, botOptions.BoostPhrases) {
			t.Errorf("Expected boost phrases to be passed to speech context")
		}
	})

	t.Run("voice_otp_and_boost_phrase", func(t *testing.T) {
		call.SetCaptureVoiceOTP("channel3", true)
		botOptions := bothelper.BotOptions{
			BoostPhrases: []string{"test1", "test2"},
		}
		call.SetBotOptions("channel3", &botOptions)
		config := newRecognitionConfig("channel3", "callSID", "en")
		if len(config.SpeechContexts) != 1 {
			t.Fatalf("Want speechcontext length=%d, got=%d", 1, len(config.SpeechContexts))
		}
		if !reflect.DeepEqual(config.SpeechContexts[0].Phrases, append(botOptions.BoostPhrases, voiceOTPPhrases...)) {
			t.Errorf("Expected voice otp and boost phrases to be passed to speech context")
		}
	})
}
