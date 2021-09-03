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
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/helper"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/grpc/counter"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/speechlogging"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
)

type AzureNew struct {
	CallSID               string
	ChannelID             string
	BoostPhrase           []string
	STTEngine             string
	MsEndpoint            string
	InitialSilenceTimeout int32
	FinalSilenceTimeout   int32
	RecognizeType         pb.RecognitionConfig_RecognitionType
	ctxCancel             context.CancelFunc
	rwCtxCancel           context.CancelFunc
	isCtxClosed           bool
	RecoTypeHandler       handlers.RecoTypeHandler
	RecognitionStatus     string
	DetectLanguage        []string
}

var grpcConnNew *grpc.ClientConn

func InitGRPCConnNew() error {
	var err error
	if len(configmanager.ConfStore.SpeechSDKEndpoints) > 0 {
		r := manual.NewBuilderWithScheme("sttservice")
		var addrs []resolver.Address
		for _, addr := range configmanager.ConfStore.SpeechSDKEndpoints {
			addrs = append(addrs, resolver.Address{
				Addr: addr,
			})
		}
		r.InitialState(resolver.State{
			Addresses: addrs,
		})

		grpcConnNew, err = grpc.Dial(r.Scheme()+":///",
			grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithResolvers(r),
			grpc.WithBalancerName(roundrobin.Name))
		ymlogger.LogInfof("InitGRPCConn", "Initialized loadbalanced connections, %#v", addrs)
	} else {
		grpcConnNew, err = grpc.Dial(configmanager.ConfStore.SpeechSDKEndpoint, grpc.WithInsecure())
		ymlogger.LogInfof("InitGRPCConn", "Initialized single connection, %s", configmanager.ConfStore.SpeechSDKEndpoint)

	}
	if err != nil {
		ymlogger.LogErrorf("InitGRPCConn", "[AzureStreamSpeechToText] Error while initializing GRPC connection. Error: [%#v]", err)
		return err
	}
	return nil
}

func (a *AzureNew) STTStreamingNew(ctx context.Context, r io.Reader) (chan *pb.RecognizeResponse, error) {
	ymlogger.LogInfo(a.CallSID, "[AzureStreamSpeechToText] Creating new STT conn")

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
		ymlogger.LogErrorf(a.CallSID, "[AzureStreamSpeechToText] Error while initializing StreamingRecognize. Error: [%#v]", err)
		close(messages)
		return messages, err
	}

	//send configuration to STT service
	ymlogger.LogDebugf("GoRoutines", "Starting one GoRoutine. [%#v]", runtime.NumGoroutine())

	configReq := &pb.RecognizeRequest{
		Audio: nil,
		Config: &pb.RecognitionConfig{
			LanguageCode:          call.GetSTTLanguage(a.ChannelID),
			Callsid:               a.CallSID,
			BoostPhrase:           a.BoostPhrase,
			SttEngine:             a.STTEngine,
			MsEndpoint:            a.MsEndpoint,
			InitialSilenceTimeout: a.InitialSilenceTimeout,
			FinalSilenceTimeout:   a.FinalSilenceTimeout,
			RecognizeType:         a.RecognizeType,
			DetectLanguages:       a.DetectLanguage,
		},
		SttServiceOptions: a.getSttServiceOptions(),
	}

	ymlogger.LogDebugf("GoRoutines", "Starting one GoRoutine. [%#v]", runtime.NumGoroutine())
	if err := stream.Send(configReq); err != nil {
		log.Fatalf(a.CallSID, "[AzureStreamSpeechToText] Configuration not sent. Error [%#v]", err)
	}
	ymlogger.LogInfof(a.CallSID, "[AzureStreamSpeechToText] Configuration sent successfully")

	counter.Increment()

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

func (a *AzureNew) getSttServiceOptions() *pb.SttServiceOptions {
	sttServiceOptions := &pb.SttServiceOptions{
		Options: &pb.SttServiceOptions_AzureOptions{
			AzureOptions: &pb.AzureOptions{},
		},
	}

	if configmanager.ConfStore.UseAzureSpeechContainer {
		sttServiceOptions.GetAzureOptions().SpeechConfig = &pb.AzureOptions_FromUrl{
			FromUrl: configmanager.ConfStore.AzureSpeechContainerUrl,
		}
	} else {
		sttRegion := configmanager.ConfStore.AzureSTTRegion
		if sttRegion == "" {
			sttRegion = "centralindia"
		}

		azureSTTAPIKey := configmanager.ConfStore.AzureSTTAPIKey

		if call.GetBotOptions(a.ChannelID) == nil {
			ymlogger.LogErrorf(a.CallSID, "BotOptions is nil: [%v]", a.ChannelID)
			return sttServiceOptions
		}

		if call.GetBotOptions(a.ChannelID).UseNewMsSubscritpion {
			azureSTTAPIKey = configmanager.ConfStore.AzureSTTAPIKeyNew
		}

		sttServiceOptions.GetAzureOptions().SpeechConfig = &pb.AzureOptions_FromSubscription_{
			FromSubscription: &pb.AzureOptions_FromSubscription{
				Region:          sttRegion,
				SubscriptionKey: azureSTTAPIKey,
			},
		}
	}

	return sttServiceOptions

}

func (a *AzureNew) writeToStreamNew(ctx context.Context, r io.Reader, stream pb.STT_StreamSpeechToTextClient) {
	defer stream.CloseSend()
	ymlogger.LogDebugf(a.CallSID, "[AzureStreamSpeechToText writeToStreamNew] GoRoutines GoRoutine Started. [%#v]", runtime.NumGoroutine())

	startTime := time.Now()

	ymlogger.LogInfof(a.CallSID, "[AzureStreamSpeechToText] Starting send to the stream: [%#v]", startTime)
	logsRequest := &speechlogging.LoggingRequest{}
	logsRequest.SetAudio(false, "languageCode", nil)

	var chunkIndices = []int{0}
	var compBuf []byte
	var sendEmptyAudio bool = true
	var totalBytesSent int
	buf := make([]byte, configmanager.ConfStore.STTStreamBufferSize)
	defer func() {
		ymlogger.LogDebugf(a.CallSID,
			"[AzureStreamSpeechToText writeToStreamNew] GoRoutines GoRoutine Ended. [%#v]",
			runtime.NumGoroutine())
	}()

	var firstEOF bool

	for {
		select {
		case <-ctx.Done():
			ymlogger.LogInfof(a.CallSID, "[AzureStreamSpeechToText] CONTEXT DONE. Write. num go routine = %d", runtime.NumGoroutine())

			// if err := stream.CloseSend(); err != nil {
			// 	ymlogger.LogErrorf(a.CallSID, "[AzureStreamSpeechToText] Could not close stream: %v", err)
			// }
			return

		default:
			sendEmptyAudio = false

			if time.Since(startTime).Seconds() > 60 {
				ymlogger.LogInfo(a.CallSID, "[AzureStreamSpeechToText] Time is over. Returning, num goroutines=%d", runtime.NumGoroutine())
				return
			}
			n, ioError := r.Read(buf)
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
				totalBytesSent += n
				ymlogger.LogDebugf(a.CallSID, "Sending audio - %d bytes, total bytes = %d, buf read error = %s", n, totalBytesSent, ioError)
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

			if ioError == io.EOF {
				if !firstEOF {
					ymlogger.LogInfof(a.CallSID, "Got eof while reading recording file, n = %d", n)
					firstEOF = true
				}
				time.Sleep(100 * time.Millisecond)
				break
			}
			if ioError != nil {
				ymlogger.LogDebugf(a.CallSID, "[AzureStreamSpeechToText writeToStreamNew] Error reading the file. %s, num goroutine=%d, bytes read = %d", ioError.Error(), runtime.NumGoroutine(), n)
				return
			}

		}
	}
}

func (a *AzureNew) readFromStreamNew(ctx context.Context, messages chan *pb.RecognizeResponse, stream pb.STT_StreamSpeechToTextClient) {
	ymlogger.LogDebugf(a.CallSID, "[AzureStreamSpeechToText readFromStreamNew] GoRoutines GoRoutine Started. [%#v]", runtime.NumGoroutine())

	defer func(messages chan *pb.RecognizeResponse) {
		ymlogger.LogDebug(a.CallSID, "[AzureStreamSpeechToText readFromStreamNew] Closing messages")
		close(messages)
	}(messages)

	// defer close(messages)
	// start := time.Now()
	initTime := time.Now()
	var latency float64

	ymlogger.LogInfo(a.CallSID, "[AzureStreamSpeechToText] Starting receive from the stream")

	for {
		select {
		case <-ctx.Done():
			if err := stream.CloseSend(); err != nil {
				ymlogger.LogErrorf(a.CallSID, "[AzureStreamSpeechToText] Could not close stream: %v", err)
			}
			ymlogger.LogInfo(a.CallSID, "[AzureStreamSpeechToText] CONTEXT DONE. Read")
			ymlogger.LogDebugf(a.CallSID, "[AzureStreamSpeechToText readFromStreamNew] GoRoutines GoRoutine Ended. [%#v]", runtime.NumGoroutine())
			return
		default:
			// if time.Since(start).Seconds() >= 2 {
			// 	start = time.Now()
			// 	ymlogger.LogInfo(a.CallSID, "[AzureStreamSpeechToText] Going to check if playback is finished..")
			// 	if call.GetPlaybackFinished(a.ChannelID) || call.GetCallFinished(a.ChannelID) {
			// 		ymlogger.LogInfo(a.CallSID, "[AzureStreamSpeechToText] Playback/Call is finished. Breaking the stream")
			// 		ymlogger.LogDebugf(a.CallSID, "[AzureStreamSpeechToText readFromStreamNew] GoRoutines GoRoutine Ended. [%#v]", runtime.NumGoroutine())

			// 		return
			// 	}
			// }

			response, err := stream.Recv()
			ymlogger.LogInfof(a.CallSID, "[AzureStreamSpeechToText] Response from the stream: [%#v]", response)
			timeSince := time.Since(initTime).Seconds()
			if err == io.EOF {
				// Final response at the end of stream
				ymlogger.LogInfo(a.CallSID, "[AzureStreamSpeechToText] Reached end of file. Breaking")
				ymlogger.LogDebugf(a.CallSID, "[AzureStreamSpeechToText readFromStreamNew] GoRoutines GoRoutine Ended. [%#v]", runtime.NumGoroutine())

				return
			}
			if err != nil {
				ymlogger.LogErrorf(a.CallSID, "[AzureStreamSpeechToText] Cannot stream results: %v", err)
				ymlogger.LogDebugf(a.CallSID, "[AzureStreamSpeechToText readFromStreamNew] GoRoutines GoRoutine Ended. [%#v]", runtime.NumGoroutine())

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
						ymlogger.LogInfof(a.CallSID, "[AzureStreamSpeechToText] Result: %+v, %s, [ListenStart-Receive = %f], [Offset+Duration = %f]  [Latency: %f]", result, result.Alternatives[0].Transcript, timeSince, float64((result.Offset+result.Duration)/10000000), latency)
					}
				}
				// if a.IsCtxClosed {
				// 	break
				// }
				messages <- response
				a.RecognitionStatus = response.RecognitionStatus
			}
			//time.Sleep(30 * time.Millisecond)
		}
	}
}

func (a *AzureNew) Close() {
	if a.isCtxClosed {
		return
	}
	a.rwCtxCancel()
	isClosed := a.RecoTypeHandler.Close()
	counter.Decrement()

	if call.GetTranscript(a.ChannelID) == "" {
		if isClosed {
			helper.SendSTTOutcome("STTOutcome", a.ChannelID, a.CallSID, "UNRECOGNIZED", a.RecognitionStatus, "azure", "streaming", counter.Get())
		} else {
			helper.SendSTTOutcome("STTOutcome", a.ChannelID, a.CallSID,"UNRECOGNIZED", "TIMEOUT", "azure", "streaming", counter.Get())
		}
	} else {
		helper.SendSTTOutcome("STTOutcome", a.ChannelID, a.CallSID,"RECOGNIZED", a.RecognitionStatus, "azure", "streaming",  counter.Get())
	}


	ymlogger.LogInfof(a.CallSID, "[AzureStreamSpeechToText] TrancriptHandler closed: [%v]", isClosed)
	a.isCtxClosed = true
	a.ctxCancel()
}
