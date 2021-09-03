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
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/handlers"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/speechlogging"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
)

type VoskSTT struct {
	CallSID               string
	ChannelID             string
	BoostPhrase           []string
	STTEngine             string
	InitialSilenceTimeout int32
	FinalSilenceTimeout   int32
	RecognizeType         pb.RecognitionConfig_RecognitionType
	ctxCancel             context.CancelFunc
	rwCtxCancel           context.CancelFunc
	isCtxClosed           bool
	RecoTypeHandler       handlers.RecoTypeHandler
	Model                 string
	SampleRateHertz       int32
}

func (a *VoskSTT) STTStreamingNew(ctx context.Context, r io.Reader) (chan *pb.RecognizeResponse, error) {
	ymlogger.LogInfo(a.CallSID, "[VoskStreamSpeechToText] Creating new STT conn")

	messages := make(chan *pb.RecognizeResponse, 10)

	ymlogger.LogDebugf("GoRoutines", "Starting one GoRoutine. [%#v]", runtime.NumGoroutine())
	client := pb.NewSTTClient(grpcConnNew)

	sttCtx, cancel := context.WithCancel(ctx)
	a.ctxCancel = cancel
	//call.SetStreamSTTCancel(a.channelID, cancel)

	// get stream(read-write)
	ymlogger.LogDebugf("GoRoutines", "Starting one GoRoutine. [%#v]", runtime.NumGoroutine())
	stream, err := client.StreamSpeechToText(sttCtx)
	if err != nil {
		ymlogger.LogErrorf(a.CallSID, "[VoskStreamSpeechToText] Error while initializing StreamingRecognize. Error: [%#v]", err)
		close(messages)
		return messages, err
	}

	//send configuration to STT service
	ymlogger.LogDebugf("GoRoutines", "Starting one GoRoutine. [%#v]", runtime.NumGoroutine())

	configReq := &pb.RecognizeRequest{
		Audio: nil,
		Config: &pb.RecognitionConfig{
			LanguageCode:    call.GetInterjectionLanguage(a.ChannelID),
			Callsid:         a.CallSID,
			BoostPhrase:     a.BoostPhrase,
			SttEngine:       a.STTEngine, //vosk
			Model:           a.Model,
			SampleRateHertz: 8000,
		},
	}

	ymlogger.LogDebugf("GoRoutines", "Starting one GoRoutine. [%#v]", runtime.NumGoroutine())
	if err := stream.Send(configReq); err != nil {
		log.Fatalf(a.CallSID, "[VoskStreamSpeechToText] Configuration not sent. Error [%#v]", err)
	}
	ymlogger.LogInfof(a.CallSID, "[VoskStreamSpeechToText] Configuration sent successfully")

	rwctx, cancel := context.WithCancel(ctx)
	a.rwCtxCancel = cancel

	// write routine
	ymlogger.LogDebugf("GoRoutines", "Starting one GoRoutine. [%#v]", runtime.NumGoroutine())
	go a.writeToStreamNew(rwctx, r, stream)

	// receive routine
	ymlogger.LogDebugf("GoRoutines", "Starting one GoRoutine. [%#v]", runtime.NumGoroutine())
	go a.readFromStreamNew(rwctx, messages, stream)

	return messages, nil
}

func (a *VoskSTT) writeToStreamNew(ctx context.Context, r io.Reader, stream pb.STT_StreamSpeechToTextClient) {
	defer stream.CloseSend()
	ymlogger.LogDebugf(a.CallSID, "[VoskStreamSpeechToText writeToStreamNew] GoRoutines GoRoutine Started. [%#v]", runtime.NumGoroutine())

	startTime := time.Now()

	ymlogger.LogInfof(a.CallSID, "[VoskStreamSpeechToText] Starting send to the stream: [%#v]", startTime)
	logsRequest := &speechlogging.LoggingRequest{}
	logsRequest.SetAudio(false, "languageCode", nil)

	var chunkIndices = []int{0}
	var compBuf []byte
	var sendEmptyAudio bool = true
	buf := make([]byte, configmanager.ConfStore.STTStreamBufferSize)

	for {
		select {
		case <-ctx.Done():
			ymlogger.LogInfo(a.CallSID, "[VoskStreamSpeechToText] CONTEXT DONE. Write")

			// if err := stream.CloseSend(); err != nil {
			// 	ymlogger.LogErrorf(a.CallSID, "[VoskStreamSpeechToText] Could not close stream: %v", err)
			// }
			ymlogger.LogInfo(a.CallSID, "[VoskStreamSpeechToText] Returning from Stream send. It's over")
			ymlogger.LogDebugf(a.CallSID, "[VoskStreamSpeechToText writeToStreamNew] GoRoutines GoRoutine Ended. [%#v]", runtime.NumGoroutine())
			return

		default:
			if time.Since(startTime).Seconds() > 60 {
				ymlogger.LogInfo(a.CallSID, "[VoskStreamSpeechToText] Time is over. Returning")
				ymlogger.LogDebugf(a.CallSID, "[VoskStreamSpeechToText writeToStreamNew] GoRoutines GoRoutine Ended. [%#v]", runtime.NumGoroutine())
				return
			}
			n, err := r.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				ymlogger.LogDebugf(a.CallSID, "[VoskStreamSpeechToText writeToStreamNew] Error reading the file. [%#v]", err)
				ymlogger.LogDebugf(a.CallSID, "[VoskStreamSpeechToText writeToStreamNew] GoRoutines GoRoutine Ended. [%#v]", runtime.NumGoroutine())
				return
			}
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
					ymlogger.LogErrorf(a.CallSID, "[VoskStreamSpeechToText] Could not send audio: %v", err)
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
					ymlogger.LogErrorf(a.CallSID, "[VoskStreamSpeechToText] Could not send audio: %v", err)
				}
			}
		}
	}
	ymlogger.LogDebugf(a.CallSID, "[VoskStreamSpeechToText writeToStreamNew] GoRoutines GoRoutine Ended. [%#v]", runtime.NumGoroutine())
}

func (a *VoskSTT) readFromStreamNew(ctx context.Context, messages chan *pb.RecognizeResponse, stream pb.STT_StreamSpeechToTextClient) {
	ymlogger.LogDebugf(a.CallSID, "[VoskStreamSpeechToText readFromStreamNew] GoRoutines GoRoutine Started. [%#v]", runtime.NumGoroutine())

	defer func(messages chan *pb.RecognizeResponse) {
		ymlogger.LogDebug(a.CallSID, "[VoskStreamSpeechToText readFromStreamNew] Closing messages")
		close(messages)
	}(messages)

	// defer close(messages)
	start := time.Now()
	initTime := time.Now()
	var latency float64

	ymlogger.LogInfo(a.CallSID, "[VoskStreamSpeechToText] Starting receive from the stream")

	for {
		select {
		case <-ctx.Done():
			if err := stream.CloseSend(); err != nil {
				ymlogger.LogErrorf(a.CallSID, "[VoskStreamSpeechToText] Could not close stream: %v", err)
			}
			ymlogger.LogInfo(a.CallSID, "[VoskStreamSpeechToText] CONTEXT DONE. Read")
			ymlogger.LogDebugf(a.CallSID, "[VoskStreamSpeechToText readFromStreamNew] GoRoutines GoRoutine Ended. [%#v]", runtime.NumGoroutine())
			return
		default:
			if time.Since(start).Seconds() >= 2 {
				start = time.Now()
				ymlogger.LogInfo(a.CallSID, "[VoskStreamSpeechToText] Going to check if playback is finished..")
				if call.GetPlaybackFinished(a.ChannelID) || call.GetCallFinished(a.ChannelID) {
					ymlogger.LogInfo(a.CallSID, "[VoskStreamSpeechToText] Playback/Call is finished. Breaking the stream")
					ymlogger.LogDebugf(a.CallSID, "[VoskStreamSpeechToText readFromStreamNew] GoRoutines GoRoutine Ended. [%#v]", runtime.NumGoroutine())

					return
				}
			}

			response, err := stream.Recv()
			ymlogger.LogInfof(a.CallSID, "[VoskStreamSpeechToText] Response from the stream: [%#v]", response)
			timeSince := time.Since(initTime).Seconds()
			if err == io.EOF {
				// Final response at the end of stream
				ymlogger.LogInfo(a.CallSID, "[VoskStreamSpeechToText] Reached end of file. Breaking")
				ymlogger.LogDebugf(a.CallSID, "[VoskStreamSpeechToText readFromStreamNew] GoRoutines GoRoutine Ended. [%#v]", runtime.NumGoroutine())

				return
			}
			if err != nil {
				ymlogger.LogErrorf(a.CallSID, "[VoskStreamSpeechToText] Cannot stream results: %v", err)
				ymlogger.LogDebugf(a.CallSID, "[VoskStreamSpeechToText readFromStreamNew] GoRoutines GoRoutine Ended. [%#v]", runtime.NumGoroutine())

				return
			}
			if response != nil {
				latency = timeSince

				if len(response.Results) > 0 {
					latency = timeSince - float64((response.Results[0].Offset+response.Results[0].Duration)/10000000)
				}

				//Needed?
				for _, result := range response.Results {
					if len(result.Alternatives) > 0 {
						ymlogger.LogInfof(a.CallSID, "[VoskStreamSpeechToText] Result: %+v, %s, [ListenStart-Receive = %f], [Offset+Duration = %f]  [Latency: %f]", result, result.Alternatives[0].Transcript, timeSince, float64((result.Offset+result.Duration)/10000000), latency)
					}
				}
				// if a.IsCtxClosed {
				// 	break
				// }
				messages <- response
			}
			time.Sleep(30 * time.Millisecond)
		}
	}
}

func (a *VoskSTT) Close() {
	ymlogger.LogInfo(a.CallSID, "[VoskStreamSpeechToText] STTHandler closing")

	if a.isCtxClosed {
		return
	}
	a.rwCtxCancel()
	ymlogger.LogInfo(a.CallSID, "[VoskStreamSpeechToText] RWContext cancelled")

	isClosed := a.RecoTypeHandler.Close()
	ymlogger.LogInfof(a.CallSID, "[VoskStreamSpeechToText] TrancriptHandler closed: [%v]", isClosed)
	a.isCtxClosed = true
	a.ctxCancel()
}
