package helper

import (
	"bytes"
	"context"
	"encoding/xml"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/newrelic"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/analytics"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"

	"github.com/go-audio/wav"
)

type STTOutcome struct {
	MetricName      string
	Status          string 
	Reason          string 
	Stt_engine      string
	Stt_type        string
}

func TrimSilence(
	ctx context.Context,
	audioFile string,
) (string, error) {
	fileWithoutExt := strings.TrimSuffix(audioFile, filepath.Ext(audioFile))
	trimmedmp3File := fileWithoutExt + "_trimmed.sln"
	cmdArguments := []string{audioFile, trimmedmp3File, "silence", "1", "0", "1%", "reverse", "silence", "1", "0", "1%", "reverse"}
	cmd := exec.Command("/usr/local/bin/sox", cmdArguments...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return trimmedmp3File, err
	}
	return trimmedmp3File, nil
}
func ConvertAudioFile(
	ctx context.Context,
	audioFile string,
) (string, error) {
	fileWithoutExt := strings.TrimSuffix(audioFile, filepath.Ext(audioFile))
	slnFile := fileWithoutExt + ".sln"
	cmdArguments := []string{"-y", "-i", audioFile, "-ar", "8000",
		"-ac", "1", "-acodec", "pcm_s16le", "-f", "s16le", slnFile}
	cmd := exec.Command("/usr/local/bin/ffmpeg", cmdArguments...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return slnFile, err
	}
	return slnFile, nil
}

func ConvertToWAV(audioFile string) (string, error) {
	fileWithoutExt := strings.TrimSuffix(audioFile, filepath.Ext(audioFile))
	wavFile := fileWithoutExt + ".wav"
	cmdArguments := []string{"-t", "raw", "-r", "16000", "-b", "16", "-c", "1",
		"-L", "-e", "signed-integer", audioFile, wavFile}

	cmd := exec.Command("/usr/local/bin/sox", cmdArguments...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return wavFile, err
	}
	return wavFile, nil
}

func ConvertToWAV8000(audioFile string) (string, error) {
	fileWithoutExt := strings.TrimSuffix(audioFile, filepath.Ext(audioFile))
	wavFile := fileWithoutExt + ".wav"
	cmdArguments := []string{"-t", "raw", "-r", "8000", "-b", "16", "-c", "1",
		"-L", "-e", "signed-integer", audioFile, wavFile}

	cmd := exec.Command("/usr/local/bin/sox", cmdArguments...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return wavFile, err
	}
	return wavFile, nil
}

func ConvertToWAV8000Bytes(audio []byte) ([]byte, error) {
	cmdArguments := []string{"-t", "raw", "-r", "8000", "-b", "16", "-c", "1",
		"-L", "-e", "signed-integer", "-", "-t", "wav", "-"}

	cmd := exec.Command("/usr/local/bin/sox", cmdArguments...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stdin = bytes.NewReader(audio)

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

// SendAccuracyMetric sends accuracy metric to new relic
func SendAccuracyMetric(metricName string, campaignID, respText string) {
	eventData := map[string]interface{}{
		"event_type":  "accuracy",
		"campaign_id": campaignID,
	}
	if len(respText) > 0 {
		eventData["status"] = "success"
	} else {
		eventData["status"] = "failure"
	}
	// Send accuracy to new relic
	if err := newrelic.SendCustomEvent(metricName, eventData); err != nil {
		ymlogger.LogErrorf("NewRelicMetric", "Failed to send "+metricName+" accuracy metric to newrelic. Error: [%#v]", err)
	}
	return
}

// SendSTTDurationMetric sends STT duration
func SendSTTDurationMetric(callSID, channelID, metricName, filePath, botID, callerID string) {
	duration, err := GetDuration(filePath)
	if err != nil {
		ymlogger.LogErrorf("NewRelicMetric", "Error while getting the duration of the file. Error: [%#v]", err)
		return
	}
	call.AddSTTDuration(channelID, duration)
	// Send STTDuration to new relic
	if err := newrelic.SendCustomEvent(metricName, map[string]interface{}{
		"event_type":     "stt_duration",
		"bot_id":         botID,
		"caller_id":      callerID,
		"campaign_id":    call.GetCampaignID(channelID),
		"duration_in_ms": duration,
	}); err != nil {
		ymlogger.LogErrorf("NewRelicMetric", "Failed to send "+metricName+" metric to newrelic. Error: [%#v]", err)
	}
	// Send STT Duration metric to druid
	event, err := analytics.PrepareAnalyticsEvent(
		analytics.STTDur,
		call.GetBotID(channelID),
		call.GetCallerID(channelID).E164Format,
		call.GetDialedNumber(channelID).E164Format,
		call.GetDirection(channelID),
		analytics.AdditionalParams{
			CallSID:     callSID,
			UTMCampaign: call.GetCampaignID(channelID),
			Value:       strconv.Itoa(int(duration)),
		},
	)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while preparing the analytics event. Error: [%#v]", err)
	} else {
		go event.Push(nil, callSID)
	}
	return
}

// GetDuration gets the duration of audio file
func GetDuration(filepath string) (int64, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	dur, err := wav.NewDecoder(f).Duration()
	if err != nil {
		return 0, err
	}
	return dur.Milliseconds(), nil
}

// SendTTSCharactersMetric sends STT duration
func SendTTSCharactersMetric(callSID, channelID, metricName, text, botID, callerID string) {
	call.AddTTSCharacters(channelID, int64(len(text)))
	// Send TTS Characters count to new relic
	if err := newrelic.SendCustomEvent(metricName, map[string]interface{}{
		"event_type":       "tts_characters",
		"bot_id":           botID,
		"caller_id":        callerID,
		"campaign_id":      call.GetCampaignID(channelID),
		"total_characters": len(text),
	}); err != nil {
		ymlogger.LogErrorf("NewRelicMetric", "Failed to send "+metricName+" metric to newrelic. Error: [%#v]", err)
	}
	// Send TTS Character metric to druid
	event, err := analytics.PrepareAnalyticsEvent(
		analytics.TTSChar,
		botID,
		callerID,
		call.GetDialedNumber(channelID).E164Format,
		call.GetDirection(channelID),
		analytics.AdditionalParams{
			CallSID:     callSID,
			UTMCampaign: call.GetCampaignID(channelID),
			Value:       strconv.Itoa(len(text)),
		},
	)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while preparing the analytics event. Error: [%#v]", err)
	} else {
		go event.Push(nil, callSID)
	}
	return
}

// SendResponseTimeMetric sends response time metric to new relic
func SendResponseTimeMetric(metricName, campaignID string, responseTime int64) {
	// Send response_time to new relic
	if err := newrelic.SendCustomEvent(metricName, map[string]interface{}{
		"campaign_id":   campaignID,
		"response_time": responseTime,
	}); err != nil {
		ymlogger.LogErrorf("NewRelicMetric", "Failed to send "+metricName+" response time metric to newrelic. Error: [%#v]", err)
	}
	return
}

func GetSSMLList(text string) []string {
	s := strings.Split(text, "</speak>")
	var ssmlList []string

	type SSMLdefault struct {
		XMLName   xml.Name `xml:"speak,omitempty"`
		Locations string   `xml:",innerxml"`
	}
	for _, ele := range s[:len(s)-1] {
		stext := ele + "</speak>"
		// check if vaid Xml && valid SSML
		if xml.Unmarshal([]byte(stext), new(interface{})) != nil && xml.Unmarshal([]byte(stext), new(SSMLdefault)) != nil {
			ymlogger.LogErrorf("Invalid xml [%v]", stext)
			continue
		}
		ssmlList = append(ssmlList, stext)
	}
	return ssmlList
}

func SendSTTLatency(metricName, campaignID string, latency int64, stt_engine string, stt_type string) {
	if err := newrelic.SendCustomEvent(metricName, map[string]interface{}{
		"campaign_id": campaignID,
		"latency":     latency,
		"type":        stt_type,
		"stt_engine":  stt_engine,
	}); err != nil {
		ymlogger.LogErrorf("NewRelicMetric", "Failed to send "+metricName+" response time metric to newrelic. Error: [%#v]", err)
	}
	return
}

func SendSTTOutcome(metricName, channelID string, callSID string, status string, reason string, stt_engine string, stt_type string, sttConnCount int32) {
	if err := newrelic.SendCustomEvent(metricName, map[string]interface{}{
		"campaign_id" : call.GetCampaignID(channelID),
		"callSID"     :  callSID,
		"status"      :  status,
		"reason"      :  reason,
		"stt_engine"  :  stt_engine,
		"stt_type"    :  stt_type,
		"conn_count"  :  sttConnCount,
	}); err != nil {
		ymlogger.LogErrorf("NewRelicMetric", "Failed to send "+metricName+" response time metric to newrelic. Error: [%#v]", err)
	}
	return
}
