package asterisk

import (
	"context"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"github.com/CyCoreSystems/ari"
	guuid "github.com/google/uuid"
)

func Play(
	ctx context.Context,
	h *ari.ChannelHandle,
	channelID string,
	callSID string,
	fileName string,
) (*ari.PlaybackHandle, error) {

	ymlogger.LogInfof(callSID, "Running Play on channel: [%#v] FileName: [%s]", channelID, fileName)
	fileWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	playbackHandle, err := h.Play(guuid.New().String(), "sound:"+fileWithoutExt)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to play sound. Error: [%#v]", err)
		return playbackHandle, err
	}

	ymlogger.LogInfo(callSID, "Completed Playback")
	return playbackHandle, err
}

func PlayWithREST(
	ctx context.Context,
	h *ari.ChannelHandle,
	channelID string,
	callSID string,
	fileName string,
) error {

	ymlogger.LogInfof(callSID, "Running Play on channel: [%#v] FileName: [%s]", channelID, fileName)
	fileWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	// Prepare the http request for destroying the channel
	chanPlayReq, err := http.NewRequest(http.MethodPost, configmanager.ConfStore.ARIURL+"/channels/"+channelID+"/play", nil)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Failed to form chan play request. Error: [%#v]", err)
		return err
	}

	// Set Basic authentication for the request
	chanPlayReq.SetBasicAuth(configmanager.ConfStore.ARIUsername, configmanager.ConfStore.ARIPassword)

	// Set required query parameters
	q := chanPlayReq.URL.Query()
	q.Add("media", "sound:"+fileWithoutExt)
	q.Add("offsetms", "0")
	q.Add("skipms", "0")
	chanPlayReq.URL.RawQuery = q.Encode()
	chanPlayReq.Header.Set("Connection", "close")

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
	response, err := client.Do(chanPlayReq)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		ymlogger.LogErrorf(callSID, "Error while playing to the channel. StatusCode: [%#v]", response.StatusCode)
		return err
	}
	return nil
}
