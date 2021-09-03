import * as grpc from 'grpc';
import * as mssdk from "microsoft-cognitiveservices-speech-sdk";
import * as pb from '../proto/msstt/msstt_pb';
import { RecognizeRequest, RecognizeResponse, RecognitionConfig } from '../proto/msstt/msstt_pb';



export interface RecognitionMode {
    onStreamStart(sdkRecognizer: mssdk.SpeechRecognizer) : Promise<boolean>;
    onStreamEnd(sdkRecognizer: mssdk.SpeechRecognizer) : Promise<boolean>;
    onTerminate(sdkRecognizer: mssdk.SpeechRecognizer) : Promise<boolean>;
    onRecognizing(call: grpc.ServerDuplexStream<RecognizeRequest, RecognizeResponse> ,resp:RecognizeResponse): Promise<boolean>;
}