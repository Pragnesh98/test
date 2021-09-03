package eventhandler

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/bothelper"
	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/callstore"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/newrelic"
	"bitbucket.org/yellowmessenger/asterisk-ari/phonenumber"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/amazon"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/asterisk"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/azure"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/google"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/helper"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"

	"github.com/CyCoreSystems/ari"
	guuid "github.com/google/uuid"
)

const (
	TTSEngineWaveNet   = "wavenet"
	TTSEngineMicrosoft = "microsoft"
	TTSURLParamType    = "url"
)

// ParseChannel gets the from and to from the channel data
func ParseChannel(
	ctx context.Context,
	channelData *ari.ChannelData,
	channelID string,
	callSID string,
	direction string,
) {
	var from, to string
	var fromPhoneNumber, toPhoneNumber phonenumber.PhoneNumber
	var err error
	if direction == call.DirectionOutbound.String() {
		return
	}
	from = channelData.GetCaller().GetNumber()
	fromPhoneNumber, err = ParseNumber(ctx, from)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to parse the From number from the channel. Error: [%#v]", err)
		fromPhoneNumber = phonenumber.PhoneNumber{
			RawNumber:      from,
			E164Format:     from,
			LocalFormat:    from,
			NationalFormat: from,
		}
	}
	call.SetDialedNumber(channelID, fromPhoneNumber)

	to = channelData.GetAccountcode()
	toPhoneNumber, err = ParseNumber(ctx, to)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to parse the To number from the channel. Error: [%#v]", err)
		toPhoneNumber = phonenumber.PhoneNumber{
			RawNumber:      to,
			E164Format:     to,
			LocalFormat:    to,
			NationalFormat: to,
		}
	}
	call.SetCallerID(channelID, toPhoneNumber)
	call.SetPipeType(channelID, call.PipeTypePRI.String())
	if strings.Contains(channelData.GetName(), configmanager.ConfStore.SIPIP) {
		call.SetPipeType(channelID, call.PipeTypeSIP.String())
	}
	return
	/*if direction == call.DirectionOutbound.String() {
		from = call.GetDialedNumber(channelID)
		if strings.ToLower(call.GetPipeType(channelID)) == strings.ToLower(call.PipeTypeSIP.String()) {
			return "+" + from, "+" + configmanager.ConfStore.CountryCode + configmanager.ConfStore.RegionCode + channelData.GetAccountcode()
		}
		return "0" + configmanager.ConfStore.CountryCode + from[1:], "+" + configmanager.ConfStore.CountryCode + channelData.GetAccountcode()[1:]
	}
	if strings.HasPrefix(strings.ToUpper(channelData.GetName()), call.PipeTypeSIP.String()) {
		to = "+" + configmanager.ConfStore.CountryCode + channelData.GetAccountcode()
	} else {
		to = "+" + configmanager.ConfStore.CountryCode + configmanager.ConfStore.RegionCode + channelData.GetAccountcode()
	}
	from = "+" + channelData.GetCaller().Number
	// If it is incoming call, set dialed number to the channel
	err = asterisk.SetChannelVariable(ctx, channelID, callSID, string(call.DialedNumber), from)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while setting the EndUserNumber channel variable. Error: [%#v]", err)
	}
	call.SetDialedNumber(channelID, from)
	call.SetCallerID(channelID, to)
	return from, to
	*/
}

// ParseNumber parses the number based on different inputs
func ParseNumber(
	ctx context.Context,
	number string,
) (phonenumber.PhoneNumber, error) {
	var pn phonenumber.PhoneNumber
	if strings.HasPrefix(number, "+") {
		pn = phonenumber.NewPhoneNumber(number)
	} else if strings.HasPrefix(number, "0") {
		pn = phonenumber.NewPhoneNumber("+" + configmanager.ConfStore.CountryCode + number[1:])
	} else if len(number) == 10 {
		pn = phonenumber.NewPhoneNumber("+" + configmanager.ConfStore.CountryCode + number)
	} else if len(number) < 10 {
		pn = phonenumber.NewPhoneNumber("+" + configmanager.ConfStore.CountryCode + configmanager.ConfStore.RegionCode + number)
	} else {
		pn = phonenumber.NewPhoneNumber("+" + number)
	}
	err := pn.Parse(ctx)
	if err != nil {
		return pn, err
	}
	return pn, nil
}

// GetTimeDifference gives the difference in seconds between two time.
// If either one of them is not set, it will return 0
func GetTimeDifference(
	ctx context.Context,
	sTime time.Time,
	eTime time.Time,
) int {
	if sTime.IsZero() || eTime.IsZero() {
		return 0
	}
	return int(sTime.Sub(eTime).Seconds())
}

// calculateStatus figures out the correct status of the call
// based on the Dialing, Ringing and Pickup Time
func calculateStatus(
	ctx context.Context,
	channelID string,
) call.Status {
	if call.GetDialingTime(channelID).IsZero() &&
		call.GetRingingTime(channelID).IsZero() &&
		call.GetPickupTime(channelID).IsZero() {
		return call.StatusNotValid
	}
	if call.GetRingingTime(channelID).IsZero() &&
		call.GetPickupTime(channelID).IsZero() {
		return call.StatusFailed
	}
	if call.GetPickupTime(channelID).IsZero() {
		return call.StatusNotAnswered
	}
	return call.StatusAnswered
}

// calculateTelcoMessage identifies the appropriate
// CauseCode and CauseText which needs to be set
func calculateTelcoMessage(
	ctx context.Context,
	channelID string,
	causeCode int,
	causeText string,
) call.CauseInfo {
	if causeCode == call.CauseCodeUnknown || causeCode == call.CauseCodeNormalUnspecified {
		if call.GetDialingTime(channelID).IsZero() {
			return call.CauseUnknown
		}
		if call.GetRingingTime(channelID).IsZero() {
			return call.CauseConnectTimeout
		}
		if call.GetPickupTime(channelID).IsZero() {
			return call.CauseRingTimeout
		}
		return call.CauseAnswered
	}
	return call.CauseInfo{
		Code: causeCode,
		Text: causeText,
	}
}

func hitBotAndStoreData(
	ctx context.Context,
	traceID string,
	channelID string,
	callSID string,
	text string,
	from string,
	to string,
	lang string,
	direction string,
	interjected bool,
) (bothelper.BotResponse, error) {
	ymlogger.LogInfof(callSID, "From number is [%s] and To number is [%s] on ChannelID: [%s]", from, to, channelID)
	call.SetBotFailed(channelID, true)
	var botResp bothelper.BotResponse
	var err error
	if text == "welcome" && call.GetWelcomeMsgAvailable(channelID) && call.GetTTSOptions(channelID) != nil {
		ttsOptions := call.GetTTSOptions(channelID)
		botResp.Data.Message = ttsOptions.Message
		botResp.Data.TextType = ttsOptions.TextType
		botResp.Data.TTSEngine = ttsOptions.TTSEngine
		botResp.Data.Options.Pitch = ttsOptions.Pitch
		botResp.Data.Options.Speed = ttsOptions.Speed
		botResp.Data.Options.VoiceID = ttsOptions.VoiceID
		botResp.Data.Language = ttsOptions.VoiceLanguage
		botResp.Disconnect = ttsOptions.Disconnect
		botResp.Success = true
	} else {
		// Hit Bot API to get the response
		botResp, err = bothelper.GetBotResponse(
			ctx,
			traceID,
			channelID,
			callSID,
			call.GetBotID(channelID),
			call.GetCampaignID(channelID),
			text,
			from,
			to,
			lang,
			call.GetDetectedLanguage(channelID),
			direction,
			call.GetRecordingFilename(channelID),
			call.GetSTTLanguage(channelID),
			interjected,
			call.GetInterjectedWords(channelID),
			call.GetChildLegStatus(channelID),
			call.GetExtraParams(channelID),
			call.GetBotRateLimiter(channelID),
		)
		if err != nil {
			ymlogger.LogErrorf(callSID, "Failed to get response from bot. Error: [%#v]. ChannelID: [%s]", err.Error(), channelID)
			// Set Recording Max Duration
			call.SetRecordingMaxDuration(channelID, time.Duration(configmanager.ConfStore.RecordingMaxDuration)*time.Second)
			// Set Recording Silence Duration
			call.SetRecordingSilenceDuration(channelID, time.Duration(configmanager.ConfStore.RecordingMaxSilence)*time.Second)
			// Reset the transcript
			call.SetTranscript(channelID, "")
			// Increment Current Bot Failure count
			call.IncrementCurrentBotFailureCount(channelID)
			// Disconnect if the max failure count has been reached
			if call.GetCurrentBotFailureCount(channelID) >= call.GetMaxBotFailureCount(channelID) {
				call.SetShouldDisconnect(channelID, true)
			}
			return botResp, err
		}
	}

	if !botResp.Success {
		ymlogger.LogErrorf(callSID, "Bot response with Success: [%#v]. ChannelID: [%s]. Retrying.....", botResp.Success, channelID)
		if err := newrelic.SendCustomEvent("bot_respStatus", map[string]interface{}{
			"botId":      call.GetBotID(channelID),
			"campaignId": call.GetCampaignID(channelID),
			"sId":        callSID,
			"success":    false,
		}); err != nil {
			ymlogger.LogErrorf("NewRelicMetric", "Failed to send bot_RespStatus metric to newrelic. Error: [%#v]", err)
		}
	}

	ymlogger.LogInfof(callSID, "Got the response from Bot. Response: [%#v].", botResp)
	call.SetBotFailed(channelID, false)

	// Set bot options on the channel
	call.SetBotOptions(channelID, &botResp.Data.Options)

	// Reset Current Failure count
	call.ResetCurrenttBotFailureCount(channelID)

	// Set Max Bot Failure Count
	if botResp.Data.Options.MaxBotFailureCount > 0 {
		call.SetMaxBotFailureCount(channelID, botResp.Data.Options.MaxBotFailureCount)
	}

	// Set BotID
	call.SetBotID(channelID, botResp.BotID)

	// Set empty transcript
	call.SetTranscript(channelID, "")
	// Reset Interjected Words
	call.ResetInterjectedWords(channelID)
	// Set Playbackfinished to false
	call.SetPlaybackFinished(channelID, false)
	// Set if the user needs to authenticated
	call.SetAuthenticateUser(channelID, botResp.Data.Options.AuthenticateUser)
	// Set the Auth Profile ID as well
	call.SetAuthProfileID(channelID, botResp.Data.Options.AuthProfileID)
	// Set TTS Engine
	call.SetTTSEngine(channelID, botResp.Data.TTSEngine)
	if len(botResp.Data.Options.TTSEngine) > 0 {
		call.SetTTSEngine(channelID, botResp.Data.Options.TTSEngine)
	}
	// Set STT Engine
	call.SetSTTEngine(channelID, botResp.Data.Options.STTEngine)
	// Set ShouldDisconnect in the call data
	call.SetShouldDisconnect(channelID, botResp.Disconnect)
	// Set the reason with which the call should be disconnected
	call.SetHangupString(channelID, botResp.Data.Options.HangupString)
	// Set if the DTMF should be captured in call data
	call.SetCaptureDTMF(channelID, botResp.Data.CaptureDTMF)
	call.SetDTMFCaptured(channelID, false)
	// Set if Voice needs to be captured
	call.SetCaptureVoice(channelID, botResp.Data.Options.CaptureVoice)
	// Set if the VoiceOTP has to be captured
	call.SetCaptureVoiceOTP(channelID, botResp.Data.Options.VoiceOTP)
	// Set the Text to speech language
	call.SetVoiceLanguage(channelID, configmanager.ConfStore.STTLanguage)
	call.SetSTTLanguage(channelID, configmanager.ConfStore.STTLanguage)
	if len(botResp.Data.Language) > 0 {
		call.SetVoiceLanguage(channelID, botResp.Data.Language)
		call.SetSTTLanguage(channelID, botResp.Data.Language)
	}
	// Set the Speech to text language
	if len(botResp.Data.Options.STTLanguage) > 0 {
		call.SetSTTLanguage(channelID, botResp.Data.Options.STTLanguage)
	}
	// Set Interjection Language
	call.SetInterjectionLanguage(channelID, call.GetSTTLanguage(channelID))
	if len(botResp.Data.Options.InterjectionLanguage) > 0 {
		call.SetInterjectionLanguage(channelID, botResp.Data.Options.InterjectionLanguage)
	}
	// Set if the Recording Beep should be played
	call.SetRecordingBeep(channelID, botResp.Data.Options.RecordingBeep)
	// Set Recording Silence Duration
	call.SetRecordingSilenceDuration(channelID, time.Duration(configmanager.ConfStore.RecordingMaxSilence)*time.Second)
	if botResp.Data.Options.RecordingSilenceDuration > 0 {
		call.SetRecordingSilenceDuration(channelID, time.Duration(botResp.Data.Options.RecordingSilenceDuration)*time.Second)
	}
	// Set Recording Max Duration
	call.SetRecordingMaxDuration(channelID, time.Duration(configmanager.ConfStore.RecordingMaxDuration)*time.Second)
	if botResp.Data.Options.RecordingMaxDuration > 0 {
		call.SetRecordingMaxDuration(channelID, time.Duration(botResp.Data.Options.RecordingMaxDuration)*time.Second)
	}
	if botResp.Data.Forward && len(botResp.Data.ForwardingNum) > 0 {
		call.SetShouldForward(channelID, botResp.Data.Forward)
		num, err := ParseNumber(ctx, botResp.Data.ForwardingNum)
		if err != nil || strings.HasPrefix(strings.ToLower(botResp.Data.ForwardingNum), "sip") {
			ymlogger.LogErrorf(callSID, "Failed to parse the Forwarding number from the bot response. Error: [%#v]", err)
			num = phonenumber.PhoneNumber{
				RawNumber:              botResp.Data.ForwardingNum,
				E164Format:             botResp.Data.ForwardingNum,
				LocalFormat:            botResp.Data.ForwardingNum,
				NationalFormat:         botResp.Data.ForwardingNum,
				WithZeroNationalFormat: botResp.Data.ForwardingNum,
			}
		}
		call.SetForwardingNumber(channelID, num)
	}
	ymlogger.LogInfo(callSID, "All options set")

	return botResp, nil
}

func getTTSFile(
	ctx context.Context,
	channelID string,
	callSID string,
	ttsengine string,
	text string,
	textType string,
	speed float64,
	voiceID string,
	pitch float64,
	deviceProfiles []string,
) (string, error) {
	var filePath string
	var err error
	switch ttsengine {
	case TTSEngineWaveNet:
		filePath, err = google.GetSpeechFile(ctx, channelID, callSID, call.GetBotID(channelID), call.GetCallerID(channelID).NationalFormat, text, textType, call.GetVoiceLanguage(channelID), speed, pitch, voiceID, deviceProfiles)
	case TTSEngineMicrosoft:
		filePath, err = azure.GetSpeechFile(ctx, channelID, callSID, call.GetBotID(channelID), call.GetCallerID(channelID).NationalFormat, text, textType, call.GetVoiceLanguage(channelID))
	default:
		filePath, err = amazon.GetSpeechFile(ctx, channelID, callSID, call.GetBotID(channelID), call.GetCallerID(channelID).NationalFormat, text, textType, voiceID)
	}
	if err != nil {
		return filePath, err
	}
	return filePath, nil
}

func (cH *CallHandlers) processUserText(
	ctx context.Context,
	channelID string,
	text string,
	interjected bool,
) {

	callSID := call.GetSID(channelID)
	latencyStore := call.GetCallLatencyStore(cH.ChannelHandler.ID())
	messageStore := call.GetCallMessageStore(cH.ChannelHandler.ID())
	ymlogger.LogInfof(callSID, "latencyStore: [%#v].", latencyStore)
	traceID := guuid.New().String()

	// channelHandler.MOH()
	// If the call is already over, return
	if len(callSID) == 0 {
		ymlogger.LogDebugf(channelID, "Call is already over. Returning..... ChannelID: [%s]", channelID)
		return
	}
	if messageStore != nil {
		go messageStore.AddNewMessage(callSID, traceID, "unknown", text, callstore.User, "")
	} else {
		ymlogger.LogError(callSID, "[MesageStore] MessageStore value is NIL")
	}
	botResponseTimeInit := time.Now()
	botResp, err := hitBotAndStoreData(ctx, traceID, channelID, callSID, text, call.GetDialedNumber(channelID).E164Format, call.GetCallerID(channelID).E164Format, "en", call.GetDirection(channelID), interjected)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to get response from bot. Error: [%#v]", err)
	}
	if text == "welcome" {
		latencyStore.AddNewStep(callSID, "first_step")
	}

	botResponseTime := time.Since(botResponseTimeInit).Milliseconds()
	if !latencyStore.RecordLatency(callSID, "unknown", callstore.BotResponseTimeinMs, botResponseTime) {
		ymlogger.LogErrorf(callSID, "[LatencyParams] Failed to record latency. [BotResponseTimeMs: %d]", botResponseTime)
	}

	if messageStore != nil {
		go messageStore.AddNewMessage(callSID, traceID, "unknown", botResp.Data.Message, callstore.Bot, "")
	} else {
		ymlogger.LogError(callSID, "[MesageStore] MessageStore value is NIL")
	}

	if botResp.Data.Message == "" {
		ymlogger.LogInfo(callSID, "Got empty message from Bot")
		cH.PlaybackHandler, err = playDefaultWelcomeMessage(ctx, channelID, callSID, cH.ChannelHandler)
		return
	}
	ttsEngine := botResp.Data.TTSEngine
	if len(botResp.Data.Options.TTSEngine) > 0 {
		ttsEngine = botResp.Data.Options.TTSEngine
	}
	speed := botResp.Data.Speed
	if botResp.Data.Options.Speed > 0 {
		speed = botResp.Data.Options.Speed
	}
	textType := botResp.Data.TextType
	if len(botResp.Data.Options.TextType) > 0 {
		textType = botResp.Data.Options.TextType
	}
	if len(botResp.Data.Options.PrefetchTTS) > 0 {
		go prefetchTTS(ctx, channelID, callSID, ttsEngine, botResp.Data.Options.PrefetchTTS, botResp.Data.Options.TTS, textType, speed, botResp.Data.Options.VoiceID, botResp.Data.Options.Pitch, botResp.Data.Options.DeviceProfiles)
	}

	ttsLatencyInit := time.Now()
	filePath, err := getSpeechFile(ctx, channelID, callSID, ttsEngine, botResp.Data.Message, botResp.Data.Options.TTS, textType, speed, botResp.Data.Options.VoiceID, botResp.Data.Options.Pitch, botResp.Data.Options.DeviceProfiles)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to get speech file from Text to Speech engine. Engine: [%v]. Error: [%#v]", botResp.Data.TTSEngine, err)
		cH.PlaybackHandler, err = playDefaultWelcomeMessage(ctx, channelID, callSID, cH.ChannelHandler)
		return
	}
	ttsResponseTime := time.Since(ttsLatencyInit).Milliseconds()
	if !latencyStore.RecordLatency(callSID, "unknown", callstore.TTSResponseTimeinMs, ttsResponseTime) {
		ymlogger.LogErrorf(callSID, "[LatencyParams] Failed to record latency. [TTSResponseTimeinMs: %d]", ttsResponseTime)
	}

	cH.PlaybackHandler, err = asterisk.Play(ctx, cH.ChannelHandler, channelID, callSID, filePath)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to play TTS file to the channel. Error: [%#v]", err)
	}
	return
}

func prefetchTTS(
	ctx context.Context,
	channelID string,
	callSID string,
	ttsEngine string,
	texts []string,
	tts []bothelper.TTSParams,
	textType string,
	speed float64,
	voiceID string,
	pitch float64,
	deviceProfiles []string,
) {
	ymlogger.LogInfof(callSID, "Prefetching text to speech for %d items", len(texts))
	for _, text := range texts {
		getSpeechFile(ctx, channelID, callSID, ttsEngine, text, tts, textType, speed, voiceID, pitch, deviceProfiles)
	}
	ymlogger.LogInfo(callSID, "Prefetch done.")
}

func playDefaultWelcomeMessage(
	ctx context.Context,
	channelID string,
	callSID string,
	channelHandler *ari.ChannelHandle,
) (*ari.PlaybackHandle, error) {
	playbackHandle, err := asterisk.Play(ctx, channelHandler, channelID, callSID, configmanager.ConfStore.DefaultWelcomeFile)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to play file to the channel. Error: [%#v]", err)
	}
	return playbackHandle, err
}

func getSpeechFile(
	ctx context.Context,
	channelID string,
	callSID string,
	ttsengine string,
	text string,
	ttsParams []bothelper.TTSParams,
	textType string,
	speed float64,
	voiceID string,
	pitch float64,
	deviceProfiles []string,
) (string, error) {
	if len(ttsParams) > 0 {
		var filePaths []string
		var filePath string
		var err error
		for _, tts := range ttsParams {
			if len(tts.Message) <= 0 {
				ymlogger.LogDebug(callSID, "Message is empty")
				continue
			}
			if tts.TextType == TTSURLParamType {
				filePath, err = downloadAudioFile(ctx, channelID, callSID, tts.Message)
				if err != nil {
					ymlogger.LogErrorf(callSID, "Error while downloading the file from the URL. Error: [%#v]", err)
					return filePath, err
				}
			} else {
				filePath, err = getTTSFile(ctx, channelID, callSID, ttsengine, tts.Message, tts.TextType, speed, voiceID, pitch, deviceProfiles)
				if err != nil {
					ymlogger.LogErrorf(callSID, "Error while getting TTS file for the text. Error: [%#v]", err)
					return filePath, err
				}
			}
			filePaths = append(filePaths, filePath)
		}
		ymlogger.LogDebugf(callSID, "All the files: [%#v]", filePaths)
		file, err := concatenateFiles(filePaths)
		if err != nil {
			ymlogger.LogErrorf(callSID, "Error while getting TTS file for the TTS Params. Error: [%#v]", err)
			return file, err
		}
		return file, nil
	}
	if isURL(text) {
		filePath, err := downloadAudioFile(ctx, channelID, callSID, text)
		if err == nil {
			return filePath, nil
		}
		ymlogger.LogErrorf(callSID, "Error while downloading the file from the URL. Error: [%#v]", err)
	}
	filePath, err := getTTSFile(ctx, channelID, callSID, ttsengine, text, textType, speed, voiceID, pitch, deviceProfiles)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while getting TTS file for the text. Error: [%#v]", err)
		return filePath, err
	}
	return filePath, nil
}

func concatenateFiles(files []string) (string, error) {
	var cmdArguments []string
	for _, file := range files {
		cmdArguments = append(cmdArguments, file)
	}
	fileName := configmanager.ConfStore.TTSFilePath + guuid.New().String() + ".sln"
	cmdArguments = append(cmdArguments, fileName)
	cmd := exec.Command("/usr/local/bin/sox", cmdArguments...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return fileName, err
	}
	return fileName, nil
}

func downloadAudioFile(
	ctx context.Context,
	channelID,
	callSID,
	url string,
) (string, error) {
	fileName := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	mp3FilePath := configmanager.ConfStore.TTSFilePath + fileName + "_url.mp3"
	// Check if the file already exists
	if _, err := os.Stat(mp3FilePath); !os.IsNotExist(err) {
		ymlogger.LogInfof(callSID, "File already exists: [%#v]", mp3FilePath)
		audioFile, _ := helper.ConvertAudioFile(ctx, mp3FilePath)
		return audioFile, nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return mp3FilePath, err
	}
	defer resp.Body.Close()

	// Create the file
	outFile, err := os.Create(mp3FilePath)
	if err != nil {
		return mp3FilePath, err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return mp3FilePath, err
	}
	audioFile, err := helper.ConvertAudioFile(ctx, mp3FilePath)
	if err != nil {
		return mp3FilePath, err
	}
	return audioFile, nil
}

func isSnoopChannel(
	channelName string,
) bool {
	return strings.HasPrefix(strings.ToLower(channelName), "snoop")
}

func isListenChannel(
	channelName string,
) bool {
	return strings.HasPrefix(strings.ToLower(channelName), "listen")
}

func isBargeINChannel(
	channelName string,
) bool {
	return strings.HasPrefix(strings.ToLower(channelName), "bargein")
}

// isURL tests a string to determine if it is a well-structured url or not.
func isURL(s string) bool {
	_, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}

	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return true
}

func isDTMFForwardingNumber(number string, dtmfForwardingNumbers []string) bool {
	for _, dtmfForwardingNumber := range dtmfForwardingNumbers {
		if dtmfForwardingNumber == number {
			return true
		}
	}
	return false
}

func uploadCallRecording(
	ctx context.Context,
	channelID string,
	callSID string,
	botID string,
) (string, error) {
	fileURL, err := azure.UploadRecording(ctx, callSID, channelID, botID, configmanager.ConfStore.RecordingDirectory+call.GetRecordingFilename(channelID)+".wav")
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to upload the recording. Error: [%#v]", err)
		return "", err
	}
	return fileURL, nil
}

func ParseRecordingFilename(
	recordingFilename string,
) (string, bool) {
	recordingFilename = strings.TrimSpace(recordingFilename)
	if len(recordingFilename) <= 0 {
		return recordingFilename, false
	}
	//can add other validation for filenames
	return recordingFilename, true
}
