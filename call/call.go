package call

import (
	"context"
	"math"
	"sync"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/globals"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"

	"bitbucket.org/yellowmessenger/asterisk-ari/utils/ratelimit"

	"bitbucket.org/yellowmessenger/asterisk-ari/bothelper"
	"bitbucket.org/yellowmessenger/asterisk-ari/callstore"
	"bitbucket.org/yellowmessenger/asterisk-ari/phonenumber"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/grpc/model"

	"github.com/CyCoreSystems/ari"
)

// Call contains call data
type Call struct {
	SID                      string
	Direction                string
	CreatedTime              time.Time
	DialingTime              time.Time
	RingingTime              time.Time
	PickupTime               time.Time
	HoldTime                 time.Time
	HoldDuration             int
	EndTime                  time.Time
	DialedNumber             phonenumber.PhoneNumber
	CallerID                 phonenumber.PhoneNumber
	Digits                   []string
	Transcript               string
	Transcripts              []string
	Duration                 int
	RingDuration             int
	BillDuration             int
	DisconnectedBy           string
	Status                   string
	ChildLegStatus           string
	Cause                    []CauseInfo
	PipeType                 string
	CallbackURL              string
	WelcomeMsgAvailable      bool
	TTSEngine                string
	TTSOptions               *bothelper.TTSOptions
	STTEngine                string
	TTSCharacters            int64
	STTDuration              int64
	AuthenticateUser         bool
	AuthProfileID            string
	CaptureVoiceOTP          bool
	CaptureDTMF              bool
	DTMFCaptured             bool
	CaptureVoice             bool
	VoiceLanguage            string
	STTLanguage              string
	InterjectionLanguage     string
	RecordingBeep            bool
	RecordingSilenceDuration time.Duration
	RecordingMaxDuration     time.Duration
	ShouldDisconnect         bool
	HangupString             string
	ShouldForward            bool
	ForwardingNumber         phonenumber.PhoneNumber
	BotFailed                bool
	RecordingURL             string
	BotID                    string
	CampaignID               string
	CurrentBotFailureCount   int8
	MaxBotFailureCount       int8
	ExtraParams              interface{}
	ParentUniqueID           string
	IsChild                  bool
	ChildrenUniqueIDs        []string
	ListenChannelID          string
	BargeINChannelID         string
	BotOptions               *bothelper.BotOptions
	StreamSTTCancel          context.CancelFunc
	STTContextCancel         context.CancelFunc
	STTHandler               model.SpeechToTextNew
	// Handlers
	ChannelHandler         *ari.ChannelHandle
	BridgeHandler          *ari.BridgeHandle
	OpToneSnoopHandler     *ari.ChannelHandle
	InterSnoopHandler      *ari.ChannelHandle
	InterRecordHandler     *ari.LiveRecordingHandle
	RecordingFilename      string
	UtteranceFilename      string
	InterRecordingFilename string
	PlaybackID             string
	PlaybackFinished       bool
	CallFinished           bool
	InterjectedWords       []string
	CallLatencyStore       callstore.LatencyStore
	CallMessageStore       callstore.MessageStore
	BotRateLimiter         *ratelimit.AdaptiveRateLimiter
	DetectedLanguage       string
}

var (
	// CallData is mapping from UniqueID to call data
	CallData     = make(map[string]*Call)
	callMapMutex = sync.RWMutex{}
)

func SetSID(channelID string, sid string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].SID = sid
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].SID = sid
	globals.IncrementNoOfCallObject()
	ymlogger.LogInfof(sid, "Number of calls [%d]. Number of call objects [%d]", globals.GetNoOfCalls(), globals.GetNoOfCallObject())
	callMapMutex.Unlock()
	return
}

func GetSID(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		sid := CallData[channelID].SID
		callMapMutex.RUnlock()
		return sid
	}
	callMapMutex.RUnlock()
	return ""
}

func SetDirection(channelID string, direction string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].Direction = direction
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].Direction = direction
	callMapMutex.Unlock()
	return
}

func GetDirection(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		direction := CallData[channelID].Direction
		callMapMutex.RUnlock()
		return direction
	}
	callMapMutex.RUnlock()
	return ""
}

func SetCreatedTime(channelID string, cTime time.Time) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].CreatedTime = cTime
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].CreatedTime = cTime
	callMapMutex.Unlock()
	return
}

func GetCreatedTime(channelID string) time.Time {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		cTime := CallData[channelID].CreatedTime
		callMapMutex.RUnlock()
		return cTime
	}
	callMapMutex.RUnlock()
	return time.Time{}
}

func SetDialingTime(channelID string, dTime time.Time) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].DialingTime = dTime
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].DialingTime = dTime
	callMapMutex.Unlock()
	return
}

func GetDialingTime(channelID string) time.Time {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		dTime := CallData[channelID].DialingTime
		callMapMutex.RUnlock()
		return dTime
	}
	callMapMutex.RUnlock()
	return time.Time{}
}

func SetRingingTime(channelID string, rTime time.Time) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].RingingTime = rTime
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].RingingTime = rTime
	callMapMutex.Unlock()
	return
}

func GetRingingTime(channelID string) time.Time {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		rTime := CallData[channelID].RingingTime
		callMapMutex.RUnlock()
		return rTime
	}
	callMapMutex.RUnlock()
	return time.Time{}
}

func SetPickupTime(channelID string, pTime time.Time) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].PickupTime = pTime
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].PickupTime = pTime
	callMapMutex.Unlock()
	return
}

func GetPickupTime(channelID string) time.Time {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		pTime := CallData[channelID].PickupTime
		callMapMutex.RUnlock()
		return pTime
	}
	callMapMutex.RUnlock()
	return time.Time{}
}

func SetHoldTime(channelID string, hTime time.Time) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].HoldTime = hTime
	}
	return
}

func GetHoldTime(channelID string) time.Time {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()
	if _, ok := CallData[channelID]; ok {
		hTime := CallData[channelID].HoldTime
		return hTime
	}
	return time.Time{}
}

func SetHoldDuration(channelID string, hDur int) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].HoldDuration += hDur
	}
	return
}

func GetHoldDuration(channelID string) int {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; ok {
		return CallData[channelID].HoldDuration
	}
	return 0
}

func SetEndTime(channelID string, eTime time.Time) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].EndTime = eTime
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].EndTime = eTime
	callMapMutex.Unlock()
	return
}

func GetEndTime(channelID string) time.Time {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		eTime := CallData[channelID].EndTime
		callMapMutex.RUnlock()
		return eTime
	}
	callMapMutex.RUnlock()
	return time.Time{}
}

func SetDialedNumber(channelID string, number phonenumber.PhoneNumber) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].DialedNumber = number
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].DialedNumber = number
	callMapMutex.Unlock()
	return
}

func GetDialedNumber(channelID string) phonenumber.PhoneNumber {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		dNum := CallData[channelID].DialedNumber
		callMapMutex.RUnlock()
		return dNum
	}
	callMapMutex.RUnlock()
	return phonenumber.PhoneNumber{}
}

func SetCallerID(channelID string, callerID phonenumber.PhoneNumber) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].CallerID = callerID
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].CallerID = callerID
	callMapMutex.Unlock()
	return
}

func GetCallerID(channelID string) phonenumber.PhoneNumber {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		callerID := CallData[channelID].CallerID
		callMapMutex.RUnlock()
		return callerID
	}
	callMapMutex.RUnlock()
	return phonenumber.PhoneNumber{}
}

func SetDigit(channelID string, digit string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].Digits = append(CallData[channelID].Digits, digit)
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].Digits = append(CallData[channelID].Digits, digit)
	callMapMutex.Unlock()
	return
}

func ResetDigits(channelID string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].Digits = []string{}
	}
	callMapMutex.Unlock()
	return
}

func GetDigits(channelID string) []string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		digits := CallData[channelID].Digits
		callMapMutex.RUnlock()
		return digits
	}
	callMapMutex.RUnlock()
	return nil
}

func SetTranscript(channelID string, text string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].Transcript = text
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].Transcript = text
	callMapMutex.Unlock()
	return
}

func GetTranscript(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		transcript := CallData[channelID].Transcript
		callMapMutex.RUnlock()
		return transcript
	}
	callMapMutex.RUnlock()
	return ""
}

func AddTranscript(channelID string, transcript string) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].Transcripts = append(CallData[channelID].Transcripts, transcript)
		return

	}
	return

}

func GetTranscripts(channelID string) []string {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()
	if _, ok := CallData[channelID]; ok {
		transcripts := CallData[channelID].Transcripts
		return transcripts
	}
	return nil
}

func SetCause(channelID string, causeInfo CauseInfo) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].Cause = append(CallData[channelID].Cause, causeInfo)
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].Cause = append(CallData[channelID].Cause, causeInfo)
	callMapMutex.Unlock()
	return
}

func GetCause(channelID string) []CauseInfo {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		cause := CallData[channelID].Cause
		callMapMutex.RUnlock()
		return cause
	}
	callMapMutex.RUnlock()
	return nil
}

func SetDuration(channelID string, dur int) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].Duration = dur
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].Duration = dur
	callMapMutex.Unlock()
	return
}

func GetDuration(channelID string) int {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		duration := CallData[channelID].Duration
		callMapMutex.RUnlock()
		return duration
	}
	callMapMutex.RUnlock()
	return 0
}

func SetRingDuration(channelID string, rDur int) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].RingDuration = rDur
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].RingDuration = rDur
	callMapMutex.Unlock()
	return
}

func GetRingDuration(channelID string) int {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		rDur := CallData[channelID].RingDuration
		callMapMutex.RUnlock()
		return rDur
	}
	callMapMutex.RUnlock()
	return 0
}

func SetBillDuration(channelID string, bDur int) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].BillDuration = bDur
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].BillDuration = bDur
	callMapMutex.Unlock()
	return
}

func GetBillDuration(channelID string) int {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		bDur := CallData[channelID].BillDuration
		callMapMutex.RUnlock()
		return bDur
	}
	callMapMutex.RUnlock()
	return 0
}

func SetStatus(channelID string, status string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].Status = status
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].Status = status
	callMapMutex.Unlock()
	return
}

func GetStatus(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		status := CallData[channelID].Status
		callMapMutex.RUnlock()
		return status
	}
	callMapMutex.RUnlock()
	return StatusUnknown.String()
}

func SetChildLegStatus(channelID string, status string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].ChildLegStatus = status
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].ChildLegStatus = status
	callMapMutex.Unlock()
	return
}

func GetChildLegStatus(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		status := CallData[channelID].ChildLegStatus
		callMapMutex.RUnlock()
		return status
	}
	callMapMutex.RUnlock()
	return ""
}

func SetDisconnectedBy(channelID string, disconnectedBy string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].DisconnectedBy = disconnectedBy
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].DisconnectedBy = disconnectedBy
	callMapMutex.Unlock()
	return
}

func GetDisconnectedBy(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		disconnectedBy := CallData[channelID].DisconnectedBy
		callMapMutex.RUnlock()
		return disconnectedBy
	}
	callMapMutex.RUnlock()
	return "user"
}

func SetPipeType(channelID string, pipe string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].PipeType = pipe
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].PipeType = pipe
	callMapMutex.Unlock()
	return
}

func GetPipeType(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		pType := CallData[channelID].PipeType
		callMapMutex.RUnlock()
		return pType
	}
	callMapMutex.RUnlock()
	return ""
}

func SetCallbackURL(channelID string, url string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].CallbackURL = url
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].CallbackURL = url
	callMapMutex.Unlock()
	return
}

func GetCallbackURL(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		cURL := CallData[channelID].CallbackURL
		callMapMutex.RUnlock()
		return cURL
	}
	callMapMutex.RUnlock()
	return ""
}

func SetWelcomeMsgAvailable(channelID string, available bool) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].WelcomeMsgAvailable = available
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].WelcomeMsgAvailable = available
	return
}

func GetWelcomeMsgAvailable(channelID string) bool {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()
	if _, ok := CallData[channelID]; ok {
		available := CallData[channelID].WelcomeMsgAvailable
		return available
	}
	return false
}

func SetTTSEngine(channelID string, engine string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].TTSEngine = engine
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].TTSEngine = engine
	callMapMutex.Unlock()
	return
}

func GetTTSEngine(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		ttsEngine := CallData[channelID].TTSEngine
		callMapMutex.RUnlock()
		return ttsEngine
	}
	callMapMutex.RUnlock()
	return ""
}

func SetTTSOptions(channelID string, ttsOptions *bothelper.TTSOptions) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].TTSOptions = ttsOptions
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].TTSOptions = ttsOptions
	return
}

func GetTTSOptions(channelID string) *bothelper.TTSOptions {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()
	if _, ok := CallData[channelID]; ok {
		options := CallData[channelID].TTSOptions
		return options
	}
	return nil
}

func SetSTTEngine(channelID string, engine string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].STTEngine = engine
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].STTEngine = engine
	callMapMutex.Unlock()
	return
}

func GetSTTEngine(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		sttEngine := CallData[channelID].STTEngine
		callMapMutex.RUnlock()
		return sttEngine
	}
	callMapMutex.RUnlock()
	return ""
}

func AddSTTDuration(channelID string, dur int64) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].STTDuration = CallData[channelID].STTDuration + dur
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].STTDuration = dur
	return
}

func GetSTTDuration(channelID string) int64 {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()
	if _, ok := CallData[channelID]; ok {
		sttDur := CallData[channelID].STTDuration
		return sttDur
	}
	return 0
}

func AddTTSCharacters(channelID string, numChar int64) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].TTSCharacters = CallData[channelID].TTSCharacters + numChar
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].TTSCharacters = numChar
	return
}

func GetTTSCharacters(channelID string) int64 {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()
	if _, ok := CallData[channelID]; ok {
		sttDur := CallData[channelID].TTSCharacters
		return sttDur
	}
	return 0
}

func SetAuthenticateUser(channelID string, authenticate bool) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].AuthenticateUser = authenticate
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].AuthenticateUser = authenticate
	callMapMutex.Unlock()
	return
}

func GetAuthenticateUser(channelID string) bool {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		authenticate := CallData[channelID].AuthenticateUser
		callMapMutex.RUnlock()
		return authenticate
	}
	callMapMutex.RUnlock()
	return false
}

func SetAuthProfileID(channelID string, profileID string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].AuthProfileID = profileID
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].AuthProfileID = profileID
	callMapMutex.Unlock()
	return
}

func GetAuthProfileID(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		authProfileID := CallData[channelID].AuthProfileID
		callMapMutex.RUnlock()
		return authProfileID
	}
	callMapMutex.RUnlock()
	return ""
}

func SetCaptureVoiceOTP(channelID string, voiceOTP bool) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].CaptureVoiceOTP = voiceOTP
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].CaptureVoiceOTP = voiceOTP
	callMapMutex.Unlock()
	return
}

func GetCaptureVoiceOTP(channelID string) bool {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		capVoiceOTP := CallData[channelID].CaptureVoiceOTP
		callMapMutex.RUnlock()
		return capVoiceOTP
	}
	callMapMutex.RUnlock()
	return false
}

func SetCaptureDTMF(channelID string, capDtmf bool) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].CaptureDTMF = capDtmf
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].CaptureDTMF = capDtmf
	callMapMutex.Unlock()
	return
}

func GetCaptureDTMF(channelID string) bool {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		capDTMF := CallData[channelID].CaptureDTMF
		callMapMutex.RUnlock()
		return capDTMF
	}
	callMapMutex.RUnlock()
	return false
}

func SetDTMFCaptured(channelID string, dtmfCap bool) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].DTMFCaptured = dtmfCap
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].DTMFCaptured = dtmfCap
	callMapMutex.Unlock()
	return
}

func GetDTMFCaptured(channelID string) bool {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		dtmfCaptured := CallData[channelID].DTMFCaptured
		callMapMutex.RUnlock()
		return dtmfCaptured
	}
	callMapMutex.RUnlock()
	return false
}

func SetCaptureVoice(channelID string, capVoice bool) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].CaptureVoice = capVoice
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].CaptureVoice = capVoice
	callMapMutex.Unlock()
	return
}

func GetCaptureVoice(channelID string) bool {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		capVoice := CallData[channelID].CaptureVoice
		callMapMutex.RUnlock()
		return capVoice
	}
	callMapMutex.RUnlock()
	return false
}

func SetVoiceLanguage(channelID string, language string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].VoiceLanguage = language
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].VoiceLanguage = language
	callMapMutex.Unlock()
	return
}

func GetVoiceLanguage(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		voiceLang := CallData[channelID].VoiceLanguage
		callMapMutex.RUnlock()
		return voiceLang
	}
	callMapMutex.RUnlock()
	return ""
}

func SetSTTLanguage(channelID string, language string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].STTLanguage = language
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].STTLanguage = language
	callMapMutex.Unlock()
	return
}

func GetSTTLanguage(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		voiceLang := CallData[channelID].STTLanguage
		callMapMutex.RUnlock()
		return voiceLang
	}
	callMapMutex.RUnlock()
	return ""
}

func SetInterjectionLanguage(channelID string, language string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].InterjectionLanguage = language
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].InterjectionLanguage = language
	callMapMutex.Unlock()
	return
}

func GetInterjectionLanguage(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		voiceLang := CallData[channelID].InterjectionLanguage
		callMapMutex.RUnlock()
		return voiceLang
	}
	callMapMutex.RUnlock()
	return ""
}

func SetRecordingBeep(channelID string, beep bool) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].RecordingBeep = beep
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].RecordingBeep = beep
	callMapMutex.Unlock()
	return
}

func GetRecordingBeep(channelID string) bool {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		recBeep := CallData[channelID].RecordingBeep
		callMapMutex.RUnlock()
		return recBeep
	}
	callMapMutex.RUnlock()
	return false
}

func SetRecordingSilenceDuration(channelID string, dur time.Duration) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].RecordingSilenceDuration = dur
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].RecordingSilenceDuration = dur
	callMapMutex.Unlock()
	return
}

func GetRecordingSilenceDuration(channelID string) time.Duration {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		recSilenceDur := CallData[channelID].RecordingSilenceDuration
		callMapMutex.RUnlock()
		return recSilenceDur
	}
	callMapMutex.RUnlock()
	return 0
}

func SetRecordingMaxDuration(channelID string, dur time.Duration) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].RecordingMaxDuration = dur
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].RecordingMaxDuration = dur
	callMapMutex.Unlock()
	return
}

func GetRecordingMaxDuration(channelID string) time.Duration {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		recMaxDur := CallData[channelID].RecordingMaxDuration
		callMapMutex.RUnlock()
		return recMaxDur
	}
	callMapMutex.RUnlock()
	return 0
}

func SetShouldDisconnect(channelID string, disconnect bool) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].ShouldDisconnect = disconnect
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].ShouldDisconnect = disconnect
	callMapMutex.Unlock()
	return
}

func GetShouldDisconnect(channelID string) bool {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		disconnect := CallData[channelID].ShouldDisconnect
		callMapMutex.RUnlock()
		return disconnect
	}
	callMapMutex.RUnlock()
	return false
}

func SetHangupString(channelID string, hangupString string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].HangupString = hangupString
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].HangupString = hangupString
	callMapMutex.Unlock()
	return
}

func GetHangupString(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		hString := CallData[channelID].HangupString
		callMapMutex.RUnlock()
		return hString
	}
	callMapMutex.RUnlock()
	return ""
}

func SetShouldForward(channelID string, forward bool) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].ShouldForward = forward
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].ShouldForward = forward
	callMapMutex.Unlock()
	return
}

func GetShouldForward(channelID string) bool {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		forward := CallData[channelID].ShouldForward
		callMapMutex.RUnlock()
		return forward
	}
	callMapMutex.RUnlock()
	return false
}

func SetBotFailed(channelID string, failed bool) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].BotFailed = failed
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].BotFailed = failed
	callMapMutex.Unlock()
	return
}

func GetBotFailed(channelID string) bool {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		botFailed := CallData[channelID].BotFailed
		callMapMutex.RUnlock()
		return botFailed
	}
	callMapMutex.RUnlock()
	return false
}

func SetRecordingURL(channelID string, rUrl string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].RecordingURL = rUrl
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].RecordingURL = rUrl
	callMapMutex.Unlock()
	return
}

func GetRecordingURL(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		rURL := CallData[channelID].RecordingURL
		callMapMutex.RUnlock()
		return rURL
	}
	callMapMutex.RUnlock()
	return ""
}

func SetBotID(channelID string, botID string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].BotID = botID
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].BotID = botID
	callMapMutex.Unlock()
	return
}

func GetBotID(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		botID := CallData[channelID].BotID
		callMapMutex.RUnlock()
		return botID
	}
	callMapMutex.RUnlock()
	return ""
}

func SetCampaignID(channelID string, campaignID string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].CampaignID = campaignID
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].CampaignID = campaignID
	callMapMutex.Unlock()
	return
}

func GetCampaignID(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		campaignID := CallData[channelID].CampaignID
		callMapMutex.RUnlock()
		return campaignID
	}
	callMapMutex.RUnlock()
	return ""
}

func IncrementCurrentBotFailureCount(channelID string) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].CurrentBotFailureCount++
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].CurrentBotFailureCount = 0
	return
}

func GetCurrentBotFailureCount(channelID string) int8 {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()
	if _, ok := CallData[channelID]; ok {
		count := CallData[channelID].CurrentBotFailureCount
		return count
	}
	return 0
}

func ResetCurrenttBotFailureCount(channelID string) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].CurrentBotFailureCount = 0
		return
	}
	return
}

func SetMaxBotFailureCount(channelID string, count int8) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].MaxBotFailureCount = count
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].MaxBotFailureCount = count
	return
}

func GetMaxBotFailureCount(channelID string) int8 {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()
	if _, ok := CallData[channelID]; ok {
		count := CallData[channelID].MaxBotFailureCount
		if count > 0 {
			return count
		}
	}
	return math.MaxInt8
}

func SetExtraParams(channelID string, params interface{}) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].ExtraParams = params
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].ExtraParams = params
	callMapMutex.Unlock()
	return
}

func GetExtraParams(channelID string) interface{} {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		rURL := CallData[channelID].ExtraParams
		callMapMutex.RUnlock()
		return rURL
	}
	callMapMutex.RUnlock()
	return ""
}

func SetParentUniqueID(channelID string, uniqueID string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].ParentUniqueID = uniqueID
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].ParentUniqueID = uniqueID
	callMapMutex.Unlock()
	return
}

func GetParentUniqueID(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		pUniqueID := CallData[channelID].ParentUniqueID
		callMapMutex.RUnlock()
		return pUniqueID
	}
	callMapMutex.RUnlock()
	return ""
}

func SetForwardingNumber(channelID string, num phonenumber.PhoneNumber) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].ForwardingNumber = num
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].ForwardingNumber = num
	callMapMutex.Unlock()
	return
}

func GetForwardingNumber(channelID string) phonenumber.PhoneNumber {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		fNum := CallData[channelID].ForwardingNumber
		callMapMutex.RUnlock()
		return fNum
	}
	callMapMutex.RUnlock()
	return phonenumber.PhoneNumber{}
}

func SetIsChild(channelID string, child bool) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].IsChild = child
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].IsChild = child
	callMapMutex.Unlock()
	return
}

func GetIsChild(channelID string) bool {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		child := CallData[channelID].IsChild
		callMapMutex.RUnlock()
		return child
	}
	callMapMutex.RUnlock()
	return false
}

func SetChildUniqueID(channelID string, uniqueID string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].ChildrenUniqueIDs = append(CallData[channelID].ChildrenUniqueIDs, uniqueID)
		callMapMutex.Unlock()
		return

	}
	CallData[channelID] = &Call{}
	CallData[channelID].Digits = append(CallData[channelID].ChildrenUniqueIDs, uniqueID)
	callMapMutex.Unlock()
	return

}

func GetChildrenUniqueIDs(channelID string) []string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		cUniqueIDs := CallData[channelID].ChildrenUniqueIDs
		callMapMutex.RUnlock()
		return cUniqueIDs
	}
	callMapMutex.RUnlock()
	return nil
}

func SetListenChannelID(channelID string, listenChannelID string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].ListenChannelID = listenChannelID
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].ListenChannelID = listenChannelID
	callMapMutex.Unlock()
	return
}

func GetListenChannelID(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		lChannelID := CallData[channelID].ListenChannelID
		callMapMutex.RUnlock()
		return lChannelID
	}
	callMapMutex.RUnlock()
	return ""
}

func SetBargeINChannelID(channelID string, bargeINChannelID string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].BargeINChannelID = bargeINChannelID
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].BargeINChannelID = bargeINChannelID
	callMapMutex.Unlock()
	return
}

func GetBargeINChannelID(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		bChannelID := CallData[channelID].BargeINChannelID
		callMapMutex.RUnlock()
		return bChannelID
	}
	callMapMutex.RUnlock()
	return ""
}

func SetChannelHandler(channelID string, handler *ari.ChannelHandle) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].ChannelHandler = handler
		callMapMutex.Unlock()
		return
	}
	callMapMutex.Unlock()
	return
}

func GetChannelHandler(channelID string) *ari.ChannelHandle {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		cHandler := CallData[channelID].ChannelHandler
		callMapMutex.RUnlock()
		return cHandler
	}
	callMapMutex.RUnlock()
	return nil
}

func SetBridgeHandler(channelID string, handler *ari.BridgeHandle) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].BridgeHandler = handler
		callMapMutex.Unlock()
		return
	}
	callMapMutex.Unlock()
	return
}

func GetBridgeHandler(channelID string) *ari.BridgeHandle {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		bHandler := CallData[channelID].BridgeHandler
		callMapMutex.RUnlock()
		return bHandler
	}
	callMapMutex.RUnlock()
	return nil
}

func SetOpToneSnoopHandler(channelID string, handler *ari.ChannelHandle) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].OpToneSnoopHandler = handler
		callMapMutex.Unlock()
		return
	}
	callMapMutex.Unlock()
	return
}

func GetOpToneSnoopHandler(channelID string) *ari.ChannelHandle {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		opToneHandler := CallData[channelID].OpToneSnoopHandler
		callMapMutex.RUnlock()
		return opToneHandler
	}
	callMapMutex.RUnlock()
	return nil
}

func SetRecordingFilename(channelID string, filename string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].RecordingFilename = filename
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].RecordingFilename = filename
	callMapMutex.Unlock()
	return
}

func GetRecordingFilename(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		recordingFilename := CallData[channelID].RecordingFilename
		callMapMutex.RUnlock()
		return recordingFilename
	}
	callMapMutex.RUnlock()
	return ""
}

func SetUtteranceFilename(channelID string, filename string) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()

	if _, ok := CallData[channelID]; !ok {
		CallData[channelID] = &Call{}
	}
	CallData[channelID].UtteranceFilename = filename
	return
}

func GetUtteranceFilename(channelID string) string {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()

	if _, ok := CallData[channelID]; !ok {
		return ""
	}
	utteranceFilename := CallData[channelID].UtteranceFilename
	return utteranceFilename
}

func SetInterRecordingFilename(channelID string, filename string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].InterRecordingFilename = filename
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].InterRecordingFilename = filename
	callMapMutex.Unlock()
	return
}

func GetInterRecordingFilename(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		recordingFilename := CallData[channelID].InterRecordingFilename
		callMapMutex.RUnlock()
		return recordingFilename
	}
	callMapMutex.RUnlock()
	return ""
}

func SetPlaybackID(channelID string, playbackID string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].PlaybackID = playbackID
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].PlaybackID = playbackID
	callMapMutex.Unlock()
	return
}

func GetPlaybackID(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		playbackID := CallData[channelID].PlaybackID
		callMapMutex.RUnlock()
		return playbackID
	}
	callMapMutex.RUnlock()
	return ""
}

func SetInterRecordHandler(channelID string, handler *ari.LiveRecordingHandle) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].InterRecordHandler = handler
		callMapMutex.Unlock()
		return
	}
	callMapMutex.Unlock()
	return
}

func GetInterRecordHandler(channelID string) *ari.LiveRecordingHandle {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		interRecordHandler := CallData[channelID].InterRecordHandler
		callMapMutex.RUnlock()
		return interRecordHandler
	}
	callMapMutex.RUnlock()
	return nil
}

func SetInterSnoopHandler(channelID string, handler *ari.ChannelHandle) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].InterSnoopHandler = handler
		callMapMutex.Unlock()
		return
	}
	callMapMutex.Unlock()
	return
}

func GetInterSnoopHandler(channelID string) *ari.ChannelHandle {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		interRecordHandler := CallData[channelID].InterSnoopHandler
		callMapMutex.RUnlock()
		return interRecordHandler
	}
	callMapMutex.RUnlock()
	return nil
}

func SetBotOptions(channelID string, botOptions *bothelper.BotOptions) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()

	if _, ok := CallData[channelID]; !ok {
		CallData[channelID] = &Call{}
	}
	CallData[channelID].BotOptions = botOptions
}

func GetBotOptions(channelID string) *bothelper.BotOptions {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()

	if _, ok := CallData[channelID]; !ok {
		return nil
	}

	if CallData[channelID].BotOptions == nil {
		return nil
	}

	// Create a copy of bot options
	botOptions := *CallData[channelID].BotOptions
	return &botOptions
}

func SetStreamSTTCancel(channelID string, cancel context.CancelFunc) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].StreamSTTCancel = cancel
		callMapMutex.Unlock()
		return
	}
	callMapMutex.Unlock()
	return
}

func GetStreamSTTCancel(channelID string) context.CancelFunc {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		streamSTTCancel := CallData[channelID].StreamSTTCancel
		callMapMutex.RUnlock()
		return streamSTTCancel
	}
	callMapMutex.RUnlock()
	return nil
}

func SetSTTContextCancel(channelID string, cancel context.CancelFunc) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; !ok {
		CallData[channelID] = &Call{}
	}
	CallData[channelID].STTContextCancel = cancel
	return
}

func GetSTTContextCancel(channelID string) context.CancelFunc {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()
	if _, ok := CallData[channelID]; !ok {
		return nil
	}
	STTContextCancel := CallData[channelID].STTContextCancel
	return STTContextCancel
}

func GetSTTHandler(channelID string) model.SpeechToTextNew {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()
	if _, ok := CallData[channelID]; !ok {
		return nil
	}
	sTTHandler := CallData[channelID].STTHandler
	return sTTHandler
}

func SetSTTHandler(channelID string, sttHandler model.SpeechToTextNew) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; !ok {
		CallData[channelID] = &Call{}
	}
	CallData[channelID].STTHandler = sttHandler
	return
}

func SetPlaybackFinished(channelID string, finished bool) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].PlaybackFinished = finished
		return
	}
	return
}

func GetPlaybackFinished(channelID string) bool {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()
	if _, ok := CallData[channelID]; ok {
		finished := CallData[channelID].PlaybackFinished
		return finished
	}
	return false
}

func SetCallFinished(channelID string, finished bool) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].CallFinished = finished
		return
	}
	return
}

func GetCallFinished(channelID string) bool {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()
	if _, ok := CallData[channelID]; ok {
		finished := CallData[channelID].CallFinished
		return finished
	}
	return false
}

func SetInterjectedWords(channelID string, word string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].InterjectedWords = append(CallData[channelID].InterjectedWords, word)
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].InterjectedWords = append(CallData[channelID].InterjectedWords, word)
	callMapMutex.Unlock()
	return
}

func ResetInterjectedWords(channelID string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].InterjectedWords = []string{}
	}
	callMapMutex.Unlock()
	return
}

func GetInterjectedWords(channelID string) []string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		words := CallData[channelID].InterjectedWords
		callMapMutex.RUnlock()
		return words
	}
	callMapMutex.RUnlock()
	return nil
}

func GetCompleteCall(channelID string) *Call {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()
	if _, ok := CallData[channelID]; ok {
		return CallData[channelID]
	}
	return nil
}

func DeleteCall(channelID string) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; ok {
		sId := CallData[channelID].SID
		delete(CallData, channelID)
		globals.DecrementNoOfCallObject()
		ymlogger.LogInfof(sId, "Number of calls [%d]. Number of call objects [%d]", globals.GetNoOfCalls(), globals.GetNoOfCallObject())
	}
	return
}

//CallLatencyStore
func GetCallLatencyStore(channelID string) *callstore.LatencyStore {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()
	if _, ok := CallData[channelID]; !ok {
		return new(callstore.LatencyStore)
	}
	return &CallData[channelID].CallLatencyStore
}
func SetCallLatencyStore(channelID string, latencyStore callstore.LatencyStore) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; !ok {
		CallData[channelID] = &Call{}
	}
	CallData[channelID].CallLatencyStore = latencyStore
	return
}

//recheck
func GetCallMessageStore(channelID string) *callstore.MessageStore {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()
	if _, ok := CallData[channelID]; !ok {
		return new(callstore.MessageStore)
	}

	return &CallData[channelID].CallMessageStore
}
func SetCallMessageStore(channelID string, messageStore callstore.MessageStore) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()
	if _, ok := CallData[channelID]; !ok {
		CallData[channelID] = &Call{}
	}
	CallData[channelID].CallMessageStore = messageStore
	return
}

func SetBotRateLimiter(channelID string, botRateLimiter *ratelimit.AdaptiveRateLimiter) {
	callMapMutex.Lock()
	defer callMapMutex.Unlock()

	if _, ok := CallData[channelID]; !ok {
		CallData[channelID] = &Call{}
	}
	CallData[channelID].BotRateLimiter = botRateLimiter
}

func GetBotRateLimiter(channelID string) *ratelimit.AdaptiveRateLimiter {
	callMapMutex.RLock()
	defer callMapMutex.RUnlock()

	if _, ok := CallData[channelID]; !ok {
		return nil
	}

	return CallData[channelID].BotRateLimiter
}

func SetDetectedLanguage(channelID string, text string) {
	callMapMutex.Lock()
	if _, ok := CallData[channelID]; ok {
		CallData[channelID].DetectedLanguage = text
		callMapMutex.Unlock()
		return
	}
	CallData[channelID] = &Call{}
	CallData[channelID].DetectedLanguage = text
	callMapMutex.Unlock()
	return
}

func GetDetectedLanguage(channelID string) string {
	callMapMutex.RLock()
	if _, ok := CallData[channelID]; ok {
		detectedLanguage := CallData[channelID].DetectedLanguage
		callMapMutex.RUnlock()
		return detectedLanguage
	}
	callMapMutex.RUnlock()
	return ""
}
