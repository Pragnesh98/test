package asterisk

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
	"github.com/google/uuid"
)

// Record records the user response
func Record(
	ctx context.Context,
	callSID string,
	h *ari.ChannelHandle,
	authenticateUser bool,
	beep bool,
) (*ari.LiveRecordingHandle, error) {
	ymlogger.LogInfof(callSID, "Going to record the call. Channel ID=[%#v]", h.ID())
	recordingFormat := configmanager.ConfStore.RecordingFormat
	if authenticateUser {
		recordingFormat = recordingFormat + "16"
	}
	if call.GetSTTEngine(h.ID()) == "microsoft" {
		recordingFormat = "wav"
	}
	//generating unique filename
	var uid uuid.UUID
	uid, err := uuid.NewRandom()
	if err != nil {
		uid, _ = uuid.NewRandom()
	}

	recordHandler, err := h.Record(uid.String(), &ari.RecordingOptions{
		Format:      recordingFormat,
		MaxDuration: call.GetRecordingMaxDuration(h.ID()),
		MaxSilence:  call.GetRecordingSilenceDuration(h.ID()),
		Exists:      "overwrite",
		Beep:        beep,
		Terminate:   configmanager.ConfStore.RecordingTerminationKey,
	})
	if err != nil {
		return recordHandler, err
	}
	call.SetUtteranceFilename(h.ID(), uid.String())
	ymlogger.LogInfof(callSID, "Recording is enqueued. Channel=[%#v]", h.ID())
	return recordHandler, nil
}

// RecordCall records the complete call
func RecordCall(
	ctx context.Context,
	fileName string,
	handler *ari.ChannelHandle,
) (*ari.LiveRecordingHandle, error) {
	callSID := call.GetSID(handler.ID())
	if len(fileName) <= 0 {
		fileName = callSID
	}
	ymlogger.LogInfof(callSID, "Going to record the complete call. Channel ID=[%#v] FileName: [%s]", handler.ID(), fileName)
	callRecordHandle, err := handler.Record(fileName, &ari.RecordingOptions{
		Format:      configmanager.ConfStore.CallRecordingFormat,
		MaxDuration: 3600 * time.Second,
		MaxSilence:  900 * time.Second,
		Exists:      "overwrite",
		Terminate:   "none",
	})
	if err != nil {
		return callRecordHandle, err
	}

	ymlogger.LogInfof(callSID, "Call Recording is enqueued. Channel=[%#v]. CallSID: [%s]", handler.ID(), fileName)
	return callRecordHandle, nil
}

func RecordWithREST(
	ctx context.Context,
	callSID string,
	h *ari.ChannelHandle,
	authenticateUser bool,
	beep bool,
) error {

	ymlogger.LogInfof(callSID, "Going to record the call. Channel ID=[%#v]", h.ID())
	recordingFormat := configmanager.ConfStore.RecordingFormat
	if authenticateUser {
		recordingFormat = recordingFormat + "16"
	}
	if call.GetSTTEngine(h.ID()) == "microsoft" {
		recordingFormat = "wav"
	}
	// Prepare the http request for destroying the channel
	chanRecReq, err := http.NewRequest(http.MethodPost, configmanager.ConfStore.ARIURL+"/channels/"+h.ID()+"/record", nil)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to form chan play request. Error: [%#v]", err)
		return err
	}

	// Set Basic authentication for the request
	chanRecReq.SetBasicAuth(configmanager.ConfStore.ARIUsername, configmanager.ConfStore.ARIPassword)

	// Set required query parameters
	q := chanRecReq.URL.Query()
	q.Add("name", h.ID())
	q.Add("format", recordingFormat)
	q.Add("maxDurationSeconds", strconv.Itoa(int(call.GetRecordingMaxDuration(h.ID()).Seconds())))
	q.Add("maxDurationSeconds", strconv.Itoa(int(call.GetRecordingSilenceDuration(h.ID()).Seconds())))
	q.Add("ifExists", "overwrite")
	q.Add("beep", strconv.FormatBool(beep))
	q.Add("terminateOn", "none")
	chanRecReq.URL.RawQuery = q.Encode()
	chanRecReq.Header.Set("Connection", "close")

	// Initlialize HTTP client
	client := &http.Client{
		Transport: &http.Transport{
			Dial:                (&net.Dialer{Timeout: 3 * time.Second}).Dial,
			TLSHandshakeTimeout: 3 * time.Second,
		},
		Timeout: time.Duration(5 * time.Second),
	}
	defer client.CloseIdleConnections()
	// Make the http request
	response, err := client.Do(chanRecReq)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		ymlogger.LogErrorf(callSID, "Error while recording to the channel. StatusCode: [%#v]", response.StatusCode)
		return err
	}
	return nil
}
