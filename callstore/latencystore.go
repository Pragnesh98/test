package callstore

import (
	"sync"

	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

type Action int

const (
	BotResponseTimeinMs Action = iota
	STTResponseTimeinMs Action = iota
	TTSResponseTimeinMs Action = iota
)

type LatencyParameter struct {
	StepIdentifier    string `json:"step_identifier"`
	BotResponseTimeMs int64  `json:"bot_response_time"`
	TTSLatencyMs      int64  `json:"text_to_speech"`
	STTLatencyMs      int64  `json:"speech_to_text"`
}

type LatencyStore struct {
	mu            sync.Mutex
	latencyparams []LatencyParameter
}

//AddNewStep adds new step object by the specified stepname
func (l *LatencyStore) AddNewStep(callSID string, stepIdentifier string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.latencyparams == nil {
		l.latencyparams = []LatencyParameter{}
	}
	ymlogger.LogInfof(callSID, "[LatencyParams] Appending new step %s", stepIdentifier)
	l.latencyparams = append(l.latencyparams, LatencyParameter{StepIdentifier: stepIdentifier})
}

//RecordLatency records respective latency
func (l *LatencyStore) RecordLatency(callSID string, stepIdentifier string, action Action, ResponseTimeMs int64) bool {
	//assign value with field.
	l.mu.Lock()
	defer l.mu.Unlock()
	//get step using stepIdentifier when stepidentifier is unique
	if len(l.latencyparams) == 0 {
		return false
	}
	index := len(l.latencyparams) - 1

	switch action {
	case BotResponseTimeinMs:
		l.latencyparams[index].BotResponseTimeMs = ResponseTimeMs
		break
	case STTResponseTimeinMs:
		l.latencyparams[index].STTLatencyMs = ResponseTimeMs
		break
	case TTSResponseTimeinMs:
		l.latencyparams[index].TTSLatencyMs = ResponseTimeMs
		break
	default:
		return false
	}
	ymlogger.LogInfof(callSID, "[LatencyParams] Latencies[ %#v] ", l.latencyparams)

	return true
}

//GetLatencies returns the list of latencies
func (l *LatencyStore) GetLatencies(callSID string) []LatencyParameter {
	l.mu.Lock()
	defer l.mu.Unlock()
	paramList := make([]LatencyParameter, len(l.latencyparams))
	copy(paramList, l.latencyparams)
	return paramList
}
