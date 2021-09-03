import 'dotenv/config';
import pino from "pino";
import * as grpc from 'grpc';
import * as pb from '../proto/msstt/msstt_pb';
import * as mssdk from "microsoft-cognitiveservices-speech-sdk";

const log = pino({ level: "info" });

export const InitializeRecognizer =  (configuration:pb.RecognitionConfig , pushStream : mssdk.PushAudioInputStream,  callback: grpc.ServerDuplexStream<pb.RecognizeRequest, pb.TextResponse>) : mssdk.SpeechRecognizer => {
    if (!process.env.MSSDK_SPEECH_SUBSCRIPTION_KEY) {
        log.error("env MSSDK_SPEECH_SUBSCRIPTON_KEY is undefined");
        throw "env MSSDK_SPEECH_SUBSCRIPTION_KEY is undefined";
    }
    let subscriptionKey: string = process.env.MSSDK_SPEECH_SUBSCRIPTION_KEY
    //crete speech config
    const speechConfig = mssdk.SpeechConfig.fromSubscription(subscriptionKey, "centralindia");
    speechConfig.speechRecognitionLanguage = configuration.getLanguageCode();
    speechConfig.outputFormat = mssdk.OutputFormat.Detailed;
    const audioConfig = mssdk.AudioConfig.fromStreamInput(pushStream);
    let sdkRecognizer = new mssdk.SpeechRecognizer(speechConfig, audioConfig);
    
    SetUpRecognizer(sdkRecognizer, callback)

    var phraseListGrammar = mssdk.PhraseListGrammar.fromRecognizer(sdkRecognizer);
    var phraseList = configuration.getBoostphraseList()

    if(phraseList && phraseList.length > 0)  {
        log.info("Boost Phrases found")
        phraseList.forEach((phrase)=>{
            phraseListGrammar.addPhrase(phrase)
        })
    }
    return sdkRecognizer
}

const res: pb.TextResponse = new pb.TextResponse();
const SetUpRecognizer = (sdkRecognizer : mssdk.SpeechRecognizer, call: grpc.ServerDuplexStream<pb.RecognizeRequest, pb.TextResponse> ) => {
    
    sdkRecognizer.recognizing = (s, e) => {
        log.info(`RECOGNIZING:${JSON.stringify(e)}`)
        let fmtResult = toRecognizeResult(e)
        log.info("Sending")
        res.setMessage(JSON.stringify(fmtResult.Recognize_Response, null, 2));
        call.write(res)
    };
    
    sdkRecognizer.recognized = (s, e) => {
        if (e.result.reason == mssdk.ResultReason.RecognizedSpeech) {
            const res: pb.TextResponse = new pb.TextResponse();
            //format results
            let fmtResult = toRecognizeResult(e) 
            res.setMessage(JSON.stringify(fmtResult.Recognize_Response));
            call.write(res);
            log.debug(`\nRECOGNIZED results sent`);
            sdkRecognizer.close();
            call.end()
            // sdkRecognizer = undefined;
            
        }
        else if (e.result.reason == mssdk.ResultReason.NoMatch) {
            let fmtResult = toRecognizeResult(e)
            res.setMessage(JSON.stringify(fmtResult.Recognize_Response));
            call.write(res);
            log.error("NOMATCH: Speech could not be recognized.");
        }
       
        sdkRecognizer.stopContinuousRecognitionAsync();
    };
    
    sdkRecognizer.canceled = (s, e) => {
        log.error(`CANCELLED: Reason=${e.reason}`);
    
        if (e.reason == mssdk.CancellationReason.Error) {
            log.error(`"CANCELED: ErrorCode=${e.errorCode}`);
            log.error(`"CANCELED: ErrorDetails=${e.errorDetails}`);
            log.error("CANCELED: Did you update the subscription info?");
        }
        call.end()
        sdkRecognizer.stopContinuousRecognitionAsync();
    };
    
    sdkRecognizer.sessionStopped = (s, e) => {
        log.error("\n    Session stopped event.");
        sdkRecognizer.stopContinuousRecognitionAsync();
    };
}

const toRecognizeResult = (sdkResult: mssdk.SpeechRecognitionEventArgs): any => {
    interface NBest {
        Confidence: number;
        Lexical: string;
    }

    const sdkJson = JSON.parse(sdkResult.result.json);

    if(sdkJson.Text){
        return {
            Recognize_Response : {
                Results : [
                    {
                        Alternatives : [
                            { Transcript: sdkJson.Text,
                            Confidence: (sdkJson.Offset + sdkJson.Duration)/10000000}
                    ],
                    ResultEndTime : { 
                        Seconds : (sdkJson.Offset + sdkJson.Duration)/10000000,
                        Nanos : (sdkJson.Offset + sdkJson.Duration)/100}
                 } ]
            }
        }
    }

    if (!sdkJson.NBest) {
        return {
            Recognize_Response : {
                Results : []
            }
        };
    }

    const nbest : NBest[] = Array.from(sdkJson.NBest);
    return {
        Recognize_Response : {
            Results : [{
                Alternatives : nbest.map<any>((nb: NBest) => {
                    return {
                        Transcript: nb.Lexical,
                        Confidence: nb.Confidence,
                    };
                })
                
            }]
        }
    }
};
