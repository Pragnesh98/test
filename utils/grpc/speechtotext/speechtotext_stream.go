package speechtotext

import (
	// "fmt"
	"context"
	"io"
	"log"
	"runtime"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/call"
	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	pb "bitbucket.org/yellowmessenger/asterisk-ari/utils/grpc/proto"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/speechlogging"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"google.golang.org/grpc"
)

type SpeechToText interface {
	STTStreaming(context.Context, io.Reader) (chan *pb.RecognizeResponse, chan bool, error)
}

type Azure struct {
	CallSID               string
	ChannelID             string
	BoostPhrase           []string
	STTEngine             string
	MsEndpoint            string
	InitialSilenceTimeout int32
	FinalSilenceTimeout   int32
	RecognizeType         pb.RecognitionConfig_RecognitionType
}

var grpcConn *grpc.ClientConn

func InitGRPCConn() error {
	var err error
	grpcConn, err = grpc.Dial(configmanager.ConfStore.SpeechSDKEndpoint, grpc.WithInsecure())
	if err != nil {
		ymlogger.LogErrorf("InitGRPCConn", "[AzureStreamSpeechToText] Error while initializing GRPC connection. Error: [%#v]", err)
		return err
	}
	return nil
}

func (a Azure) STTStreaming(ctx context.Context, r io.Reader) (chan *pb.RecognizeResponse, chan bool, error) {
	ymlogger.LogDebugf("GoRoutines", "GoRoutine Started. [%#v]", runtime.NumGoroutine())
	messages := make(chan *pb.RecognizeResponse, 10)
	done := make(chan bool)

	ymlogger.LogDebugf("GoRoutines", "Starting one GoRoutine. [%#v]", runtime.NumGoroutine())
	client := pb.NewSTTClient(grpcConn)

	// get stream(read-write)
	ymlogger.LogDebugf("GoRoutines", "Starting one GoRoutine. [%#v]", runtime.NumGoroutine())
	stream, err := client.StreamSpeechToText(ctx)
	if err != nil {
		ymlogger.LogErrorf(a.CallSID, "[AzureStreamSpeechToText] Error while initializing StreamingRecognize. Error: [%#v]", err)
		close(messages)
		return messages, done, err
	}

	//send configuration to STT service
	ymlogger.LogDebugf("GoRoutines", "Starting one GoRoutine. [%#v]", runtime.NumGoroutine())

	configReq := &pb.RecognizeRequest{
		Audio: nil,
		Config: &pb.RecognitionConfig{
			LanguageCode:          call.GetInterjectionLanguage(a.ChannelID),
			Callsid:               a.CallSID,
			BoostPhrase:           a.BoostPhrase,
			SttEngine:             a.STTEngine,
			MsEndpoint:            a.MsEndpoint,
			InitialSilenceTimeout: a.InitialSilenceTimeout,
			FinalSilenceTimeout:   a.FinalSilenceTimeout,
			RecognizeType:         a.RecognizeType,
		},
	}
	ymlogger.LogDebugf("GoRoutines", "Starting one GoRoutine. [%#v]", runtime.NumGoroutine())
	if err := stream.Send(configReq); err != nil {
		log.Fatalf("can not send %v", err)
	}

	// write routine
	ymlogger.LogDebugf("GoRoutines", "Starting one GoRoutine. [%#v]", runtime.NumGoroutine())
	go a.writeToStream(ctx, done, r, stream)

	// receive routine
	ymlogger.LogDebugf("GoRoutines", "Starting one GoRoutine. [%#v]", runtime.NumGoroutine())
	go a.readFromStream(ctx, done, messages, stream)

	ymlogger.LogDebugf("GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
	// return channel
	return messages, done, nil
}

func (a Azure) writeToStream(ctx context.Context, done chan bool, r io.Reader, stream pb.STT_StreamSpeechToTextClient) {
	defer stream.CloseSend()
	ymlogger.LogDebugf("GoRoutines", "GoRoutine Started. [%#v]", runtime.NumGoroutine())
	startTime := time.Now()
	ymlogger.LogInfof(a.CallSID, "[AzureStreamSpeechToText] Starting send to the stream: [%#v]", startTime)
	logsRequest := &speechlogging.LoggingRequest{}
	logsRequest.SetAudio(false, "languageCode", nil)

	var chunkIndices = []int{0}
	var compBuf []byte
	var sendEmptyAudio bool = true
	buf := make([]byte, configmanager.ConfStore.STTStreamBufferSize)

	for {
		select {
		case <-done:
			if err := stream.CloseSend(); err != nil {
				ymlogger.LogErrorf(a.CallSID, "[AzureStreamSpeechToText] Could not close stream: %v", err)
			}
			ymlogger.LogInfo(a.CallSID, "[AzureStreamSpeechToText] Returning from Stream send. It's over")
			go prepareAndSendLogs(logsRequest, a.ChannelID, a.CallSID, compBuf, chunkIndices, []string{})
			ymlogger.LogDebugf("GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
			return

		default:
			if time.Since(startTime).Seconds() > 60 {
				ymlogger.LogInfo(a.CallSID, "[AzureStreamSpeechToText] Time is over. Returning")
				ymlogger.LogDebugf("GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
				return
			}
			n, err := r.Read(buf)
			if n > 0 {
				req := &pb.RecognizeRequest{
					Audio: &pb.RecognitionAudio{
						AudioSource: &pb.RecognitionAudio_Content{
							Content: buf[:n],
						},
					},
					Config: &pb.RecognitionConfig{
						LanguageCode: call.GetInterjectionLanguage(a.ChannelID),
						Callsid:      a.CallSID,
						BoostPhrase:  a.BoostPhrase,
						SttEngine:    a.STTEngine,
					},
				}

				sendEmptyAudio = false
				if err := stream.Send(req); err != nil {
					ymlogger.LogErrorf(a.CallSID, "[AzureStreamSpeechToText] Could not send audio: %v", err)
				}
				compBuf = append(compBuf, buf...)
				chunkIndices = append(chunkIndices, chunkIndices[len(chunkIndices)-1]+n)

			} else if sendEmptyAudio /*|| time.Since(initTime).Seconds() > 8 */ {
				b := []byte{0}
				req := &pb.RecognizeRequest{
					Audio: &pb.RecognitionAudio{
						AudioSource: &pb.RecognitionAudio_Content{
							Content: b,
						},
					},
					Config: nil,
				}
				if err := stream.Send(req); err != nil {
					ymlogger.LogErrorf(a.CallSID, "[AzureStreamSpeechToText] Could not send audio: %v", err)
				}
			}
			if err != nil {
				continue
			}
		}
	}
	return
}

func (a Azure) readFromStream(ctx context.Context, done chan bool, messages chan *pb.RecognizeResponse, stream pb.STT_StreamSpeechToTextClient) {
	ymlogger.LogDebugf("GoRoutines", "GoRoutine Started. [%#v]", runtime.NumGoroutine())
	defer close(messages)
	start := time.Now()
	initTime := time.Now()
	var latency float64

	ymlogger.LogInfo(a.CallSID, "[AzureStreamSpeechToText] Starting receive from the stream")

	// var response *speechpb.RecognizeResponse
	for {
		if time.Since(start).Seconds() >= 2 {
			start = time.Now()
			ymlogger.LogInfo(a.CallSID, "[AzureStreamSpeechToText] Going to check if playback is finished..")
			if call.GetPlaybackFinished(a.ChannelID) || call.GetCallFinished(a.ChannelID) {
				ymlogger.LogInfo(a.CallSID, "[AzureStreamSpeechToText] Playback/Call is finished. Breaking the stream")
				done <- true
				break
			}
		}

		response, err := stream.Recv()
		ymlogger.LogInfof(a.CallSID, "[AzureStreamSpeechToText] Response from the stream: [%#v]", response)
		timeSince := time.Since(initTime).Seconds()
		if err == io.EOF {
			// Final response at the end of stream
			ymlogger.LogInfo(a.CallSID, "[AzureStreamSpeechToText] Reached end of file. Breaking")
			break
		}
		if err != nil {
			ymlogger.LogErrorf(a.CallSID, "[AzureStreamSpeechToText] Cannot stream results: %v", err)
			break
		}
		// //Parse text response to response object
		// err = json.Unmarshal([]byte(resp.GetMessage()), &response)
		// if err != nil {
		// 	ymlogger.LogErrorf(a.CallSID, "Couldn't unmarshall response: [%v]", resp.GetMessage())
		// 	ymlogger.LogDebugf("GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
		// 	break
		// }
		if response != nil {

			latency = timeSince
			if len(response.Results) > 0 {
				latency = timeSince - float64((response.Results[0].Offset+response.Results[0].Duration)/10000000)
			}
			for _, result := range response.Results {
				if len(result.Alternatives) > 0 {
					ymlogger.LogInfof(a.CallSID, "[AzureStreamSpeechToText] Result: %+v, %s, [ListenStart-Receive = %f], [Offset+Duration = %f]  [Latency: %f]", result, result.Alternatives[0].Transcript, timeSince, float64((result.Offset+result.Duration)/10000000), latency)
				}
			}
			messages <- response
		}
		time.Sleep(30 * time.Millisecond)
	}
	done <- true
	ymlogger.LogDebugf("GoRoutines", "GoRoutine Ended. [%#v]", runtime.NumGoroutine())
	return
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
		SetSttService("azure").
		SetBotID(botID).
		SetUserID(userID).
		SetCallStartTime(callStartTime).
		SetStreamingRecognizeInfo(buf, chunkIndices, transcripts)

	speechlogging.Send(logsRequest, speechlogging.URL)
	return
}
