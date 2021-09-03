package callstore

import (
	"sync"

	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

type MessageSource string

const (
	Bot MessageSource = "bot"
	User MessageSource = "user"
)

type Message struct {
	StepIdentifier    string         `json:"StepIdentifier"`
	TraceId           string         `json:trace_id`
	Message           string         `json:message`
	Type              MessageSource  `json:type`
	Recording         string         `json:recording`
}

type MessageStore struct {
	mu            sync.Mutex
	messageList 	[]Message
}

//AddNewStep adds new step object by the specified stepname
func (m *MessageStore) AddNewMessage(callSID string, traceID string, stepIdentifier string, message string, msgSource MessageSource, recording string) {
	ymlogger.LogInfof(callSID, "[MessageStore] Appending new message [%s]: [%s]", stepIdentifier, message)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.messageList == nil {
		m.messageList = []Message{}
	}

	ymlogger.LogInfof(callSID, "[MessageStore] Message so far [%v]:", m.messageList)
	m.messageList = append(
		m.messageList, 
		Message{ 
			TraceId: traceID,
			StepIdentifier: stepIdentifier,
			Message: message,
			Type: msgSource,
			Recording: recording,
	})
}

//GetMessages returns the list of messsages
func (m *MessageStore) GetMessages(callSID string) []Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	paramList := make([]Message, len(m.messageList))
	copy(paramList, m.messageList)
	ymlogger.LogInfof(callSID, "[MessageStore] Returning message List [%#v]", paramList)

	return paramList
}
