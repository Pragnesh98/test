package google

import (
	"context"
	"os"
	"strings"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/helper"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/speechlogging"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	speech "cloud.google.com/go/speech/apiv1"
	"github.com/CyCoreSystems/ari"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

func GetStreamTextFromSpeech(
	ctx context.Context,
	channelID string,
	callSID string,
	fileName string,
	interjectUtterances []string,
	playbackHandle *ari.PlaybackHandle,
	playbackId string,
	boostPhrase []string,
) (string, error) {
	var text string
	streamCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := speech.NewClient(streamCtx)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while initializing the speech client. Error: [%#v]", err)
		return text, err
	}
	defer client.Close()

	stream, err := client.StreamingRecognize(streamCtx)
	if err != nil {
		ymlogger.LogErrorf(callSID, "Error while initializing StreamingRecognize. Error: [%#v]", err)
		return text, err
	}
	defer stream.CloseSend()

	languageCode := configmanager.ConfStore.STTLanguage
	if len(call.GetInterjectionLanguage(channelID)) > 0 {
		languageCode = call.GetInterjectionLanguage(channelID)
	}
	//Initialize Recognition config
	recConfig := newRecognitionConfig(channelID, callSID, languageCode)

	// Send the initial configuration message.
	if err := stream.Send(&speechpb.StreamingRecognizeRequest{
		StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
			StreamingConfig: &speechpb.StreamingRecognitionConfig{
				Config: recConfig,
			},
		},
	}); err != nil {
		ymlogger.LogErrorf(callSID, "Error while sending the configuration to stream. Error: [%#v]", err)
		return text, err
	}

	startTime := time.Now()
	for {
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			if time.Since(startTime).Milliseconds() > 6000 {
				ymlogger.LogErrorf(callSID, "Error while opening the file. Error: [%#v]", err)
				break
			}
			continue
		}
		break
	}
	startTime = time.Now()
	var f *os.File
	for {
		f, err = os.Open(fileName)
		if err != nil {
			if time.Since(startTime).Milliseconds() > 6000 {
				ymlogger.LogErrorf(callSID, "Error while opening the file. Error: [%#v]", err)
				break
			}
			continue
		}
		break
	}
	defer f.Close()

	startTime = time.Now()
	done := make(chan bool)
	var words []string
	ymlogger.LogInfof(callSID, "[StreamSpeechToText] Starting send to the stream: [%#v]", startTime)

	go func(languageCode string, words *[]string) {
		logsRequest := &speechlogging.LoggingRequest{}
		logsRequest.SetAudio(false, languageCode, nil)
		var chunkIndices = []int{0}
		var compBuf []byte
		// initTime := time.Now()
		var sendEmptyAudio bool = true
		buf := make([]byte, configmanager.ConfStore.STTStreamBufferSize)
		for {
			select {
			case <-done:
				if err := stream.CloseSend(); err != nil {
					ymlogger.LogErrorf(callSID, "[StreamSpeechToText] Could not close stream: %v", err)
				}
				ymlogger.LogInfo(callSID, "[StreamSpeechToText] Returning from Stream send. It's over")
				go prepareAndSendLogs(logsRequest, channelID, callSID, compBuf, chunkIndices, *words)
				return
			default:
				if time.Since(startTime).Seconds() > 60 {
					ymlogger.LogInfo(callSID, "[StreamSpeechToText] Time is over. Returning")
					return
				}
				n, err := f.Read(buf)
				if n > 0 {
					// ymlogger.LogInfo(callSID, "[StreamSpeechToText] Got the buffer from the file.")
					sendEmptyAudio = false
					// initTime = time.Now()
					if err := stream.Send(&speechpb.StreamingRecognizeRequest{
						StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
							AudioContent: buf[:n],
						},
					}); err != nil {
						ymlogger.LogErrorf(callSID, "[StreamSpeechToText] Could not send audio: %v", err)
					}
					compBuf = append(compBuf, buf...)
					chunkIndices = append(chunkIndices, chunkIndices[len(chunkIndices)-1]+n)
				} else if sendEmptyAudio /*|| time.Since(initTime).Seconds() > 8*/ {
					// initTime = time.Now()
					b := []byte{0}
					// b := make([]byte, configmanager.ConfStore.STTStreamBufferSize)
					if err := stream.Send(&speechpb.StreamingRecognizeRequest{
						StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
							AudioContent: b,
						},
					}); err != nil {
						ymlogger.LogErrorf(callSID, "[StreamSpeechToText] Could not send audio: %v", err)
					}
				}
				if err != nil {
					continue
				}
			}
		}
	}(languageCode, &words)

	start := time.Now()
	initTime := time.Now()

	ymlogger.LogInfo(callSID, "[StreamSpeechToText] Starting receive from the stream")
	for {
		if time.Since(start).Seconds() >= 2 {
			start = time.Now()
			ymlogger.LogInfo(callSID, "[StreamSpeechToText] Going to check if playback is finished..")
			if call.GetPlaybackFinished(channelID) || call.GetCallFinished(channelID) {
				ymlogger.LogInfo(callSID, "[StreamSpeechToText] Playback is finished. Breaking the stream")
				done <- true
				break
			}
		}
		resp, err := stream.Recv()
		if err != nil {
			ymlogger.LogInfof(callSID, "[StreamSpeechToText] Error while receiving the stream [%#v]", err)
			done <- true
			break
		}
		if resp != nil && resp.Error != nil {
			ymlogger.LogErrorf(callSID, "[StreamSpeechToText] Could not recognize: %v", err)
			continue
		}
		if resp != nil {
			for _, result := range resp.Results {
				latency := time.Since(initTime).Seconds() + float64(result.ResultEndTime.Seconds)
				ymlogger.LogInfof(callSID, "[StreamSpeechToText] Result: %+v, %s PlaybackID [%#v] Latency: [%f]", result, result.Alternatives[0].Transcript, playbackHandle.ID(), latency)
				if call.GetPlaybackID(channelID) == playbackId {
					ymlogger.LogInfo(callSID, "[StreamSpeechToText] Playback is same. Will break from here")
				}
				text = text + result.Alternatives[0].Transcript
				call.SetInterjectedWords(channelID, result.Alternatives[0].Transcript)
				words = append(words, result.Alternatives[0].Transcript)
				if utteranceExists(interjectUtterances, result.Alternatives[0].Transcript) {
					ymlogger.LogDebugf(callSID, "[StreamSpeechToText] Stream Send: Found user utterance, closing the stream")
					if call.GetPlaybackID(channelID) == playbackId {
						if playbackHandle != nil {
							// ymlogger.LogDebug(callSID, "[StreamSpeechToText] Mocking the playback stop")
							playbackHandle.Stop()
						}
						// call.SetTranscript(channelID, text)
					}
					done <- true
					go helper.SendSTTDurationMetric(callSID, channelID, "voice_stt_google_stream", fileName, call.GetBotID(channelID), call.GetCallerID(channelID).NationalFormat)
					return text, nil
				}
			}
		}
	}
	go helper.SendSTTDurationMetric(callSID, channelID, "voice_stt_google_stream", fileName, call.GetBotID(channelID), call.GetCallerID(channelID).NationalFormat)
	return text, nil
}

func utteranceExists(utterances []string, text string) bool {
	for _, utterance := range utterances {
		if strings.Contains(strings.ToLower(text), strings.ToLower(utterance)) {
			return true
		}
	}
	return false
}

func prepareAndSendLogs(
	logsRequest *speechlogging.LoggingRequest,
	channelID string,
	callSID string,
	buf []byte,
	chunkIndices []int,
	transcripts []string,
) {
	// Send request and response for logging
	callStartTime := call.GetPickupTime(channelID)
	botID := call.GetCallerID(channelID).E164Format
	userID := call.GetDialedNumber(channelID).E164Format
	logsRequest.
		SetCallSID(callSID).
		SetSttService("google").
		SetBotID(botID).
		SetUserID(userID).
		SetCallStartTime(callStartTime).
		SetStreamingRecognizeInfo(buf, chunkIndices, transcripts)

	speechlogging.Send(logsRequest, speechlogging.URL)
	return
}
