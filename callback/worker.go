package callback

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/models/mysql"
	"bitbucket.org/yellowmessenger/asterisk-ari/newrelic"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

var client *http.Client

// InitCallbackClient initializes the HTTP client for callbacks
func InitCallbackClient() {
	client = &http.Client{
		Transport: &http.Transport{
			Dial:                (&net.Dialer{Timeout: 3 * time.Second}).Dial,
			TLSHandshakeTimeout: 3 * time.Second,
		},
		Timeout: time.Duration(10 * time.Second),
	}
	return
}

func callbackQueueWorker(ctx context.Context, queue chan mysql.Callback) {
	for {
		callback := <-queue
		HitCallback(ctx, callback)
	}
}

// StartWorker starts the callback worker
func StartWorker(ctx context.Context) {
	callbackQueue := make(chan mysql.Callback, 1)

	for i := 0; i < 20; i++ {
		go callbackQueueWorker(ctx, callbackQueue)
	}

	for {
		ymlogger.LogInfo("CallbackWorker", "Trying to get the scheduled callbacks")
		callbacks, iDs, err := mysql.GetScheduledCallbacks()
		if err != nil {
			ymlogger.LogErrorf("CallbackWorker", "Error while getting the callbacks. Error: [%#v]", err)
			continue
		}
		if len(callbacks) <= 0 {
			ymlogger.LogInfo("CallbackWorker", "No pending callback event found. Sleep for 10 Seconds")
			time.Sleep(10 * time.Second)
			continue
		}
		if err = mysql.MarkCallbackInProgress(iDs); err != nil {
			ymlogger.LogErrorf("CallbackWorker", "Error while marking the callbacks in-progress. Error: [%#v]", err)
			continue
		}
		for _, callback := range callbacks {
			callbackQueue <- callback
		}

		ymlogger.LogInfo("CallbackWorker", "Sleep for 500 millisecond")
		time.Sleep(500 * time.Millisecond)
	}
}

// HitCallback makes the callback to callback url
func HitCallback(
	ctx context.Context,
	callback mysql.Callback,
) {

	ymlogger.LogDebugf(callback.SID, "Hitting Callback URL with request Body: [%s]", callback.Payload)
	callbackReq, err := http.NewRequest(http.MethodPost, callback.CallbackURL, bytes.NewBuffer([]byte(callback.Payload)))
	if err != nil {
		ymlogger.LogErrorf(callback.SID, "Failed to prepare the call back request. Error: [%#v]", err)
		ChangeCallbackStatus(callback.ID, callback.SID, callback.CallbackURL, "scheduled")
		return
	}
	callbackReq.Host = "app.yellowmessenger.com"
	callbackReq.Header.Set("Content-Type", "application/json")
	callbackReq.Header.Set("Authorization", configmanager.ConfStore.GoogleAccessToken)
	callbackReq.Header.Set("Connection", "close")

	var response *http.Response
	for i := 0; i < configmanager.ConfStore.CallbackMaxTries; i++ {
		response, err = client.Do(callbackReq)
		newrelic.SendCustomEvent("callbacks_metrics", map[string]interface{}{
			"status": "request_sent",
			"value":  1,
		})
		if response == nil || response.StatusCode < 200 || response.StatusCode >= 300 || err != nil {
			ymlogger.LogErrorf(callback.SID, "Retry: [%d]. Failed hitting the callback URL. Response: [%#v]. Error: [%#v]. Retrying", (i + 1), response, err)
			urlError, ok := err.(*url.Error)
			if ok {
				ymlogger.LogErrorf(callback.SID, "Logging the exact error. Error: [%#v]", urlError)
			}
			time.Sleep(time.Duration((i+1)*10) * time.Second)
			continue
		}
		break
	}
	if response == nil || response.StatusCode < 200 || response.StatusCode >= 300 || err != nil {
		ymlogger.LogErrorf(callback.SID, "Failed to hit the callback URL. Error: [%#v]", err)
		ChangeCallbackStatus(callback.ID, callback.SID, callback.CallbackURL, "scheduled")
		return
	}
	defer response.Body.Close()
	newrelic.SendCustomEvent("callbacks_metrics", map[string]interface{}{
		"status": "success",
		"value":  1,
	})
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ymlogger.LogErrorf(callback.SID, "Failed to get body from the response. Error: [%#v]", err)
		ChangeCallbackStatus(callback.ID, callback.SID, callback.CallbackURL, "scheduled")
		return
	}
	ymlogger.LogInfof(callback.SID, "Successful response from the callback. Body: [%#v]", string(respBody))
	ChangeCallbackStatus(callback.ID, callback.SID, callback.CallbackURL, "completed")
	return
}

func ChangeCallbackStatus(
	ID int64,
	callSID string,
	callbackURL string,
	status string,
) {
	ymlogger.LogInfof(callSID, "Changing the callback status to [%s]", status)
	var err error
	for i := 0; i < 3; i++ {
		switch status {
		case "scheduled":
			err = mysql.MarkCallbackScheduled(ID, callbackURL)
		case "completed":
			err = mysql.MarkCallbackCompleted(ID, callbackURL)
		default:
			ymlogger.LogErrorf(callSID, "Invalid status")
		}
		if err != nil {
			ymlogger.LogErrorf(callSID, "Failed marking the status: [%s] Error: [%#v]", status, err)
			time.Sleep(time.Duration((i+1)*10) * time.Second)
			continue
		}
		break
	}
	return
}
