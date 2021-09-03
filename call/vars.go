package call

type CallDirection int
type Operator int
type PipeType int
type State int
type Status int

const (
	CallSID      = "CallSid"
	DialedNumber = "DialedNumber"
	Direction    = "Direction"
)

//go:generate stringer -type=CallDirection -linecomment
const (
	DirectionInbound  CallDirection = iota + 1 // inbound
	DirectionOutbound                          // outbound
)

//go:generate stringer -type=State -trimprefix=State
const (
	StateDialing State = iota + 1
	StateRinging
	StateUp
)

//go:generate stringer -type=PipeType -trimprefix=PipeType
const (
	PipeTypeSIP PipeType = iota + 1
	PipeTypePRI
)

//go:generate stringer -type=Operator -linecomment
const (
	OperatorTataPRI Operator = iota + 1 //Tata_PRI
	OperatorTataSIP                     // Tata_SIP
)

//go:generate stringer -type=Status -linecomment
const (
	StatusInitiated   Status = iota + 1 // initiated
	StatusAnswered                      // answered
	StatusNotAnswered                   // not_answered
	StatusFailed                        // failed
	StatusNotValid                      // not_valid
	StatusUnknown                       // unknown
)

const (
	CauseCodeUnknown           int = 0
	CauseCodeNormalUnspecified     = 31
	CauseCodeAnswered              = 130
	CauseCodeConnectTimeout        = 131
	CauseCodeRingTimeout           = 132
	CauseCodeUnidentified          = 133
)

const (
	CauseTextUnknown           string = "Unknown"
	CauseTextNormalUnspecified        = "Normal, unspecified"
	CauseTextAnswered                 = "Answered"
	CauseTextConnectTimeout           = "Connect Timeout"
	CauseTextRingTimeout              = "Ring Timeout"
	CauseTextUnidentified             = "Unidentified"
)

type CauseInfo struct {
	Code int
	Text string
}

func NewCauseInfo(code int, text string) CauseInfo {
	return CauseInfo{
		Code: code,
		Text: text,
	}
}

var (
	CauseUnknown           = NewCauseInfo(CauseCodeUnknown, CauseTextUnknown)
	CauseNormalUnspecified = NewCauseInfo(CauseCodeNormalUnspecified, CauseTextNormalUnspecified)
	CauseAnswered          = NewCauseInfo(CauseCodeAnswered, CauseTextAnswered)
	CauseConnectTimeout    = NewCauseInfo(CauseCodeConnectTimeout, CauseTextConnectTimeout)
	CauseRingTimeout       = NewCauseInfo(CauseCodeRingTimeout, CauseTextRingTimeout)
	CauseUnidentified      = NewCauseInfo(CauseCodeUnidentified, CauseTextUnidentified)
)
