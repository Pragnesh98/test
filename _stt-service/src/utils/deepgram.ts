import 'dotenv/config';
import pino from "pino";
import * as grpc from 'grpc';
(global as any).WebSocket = require('ws');
import * as pb from '../proto/msstt/msstt_pb';
import { RecognizeResponse, RecognizeRequest } from '../proto/msstt/msstt_pb';

const log = pino({ level: "info" });

//connect to deepgram engine 
export var wsstream = function(call:grpc.ServerDuplexStream<RecognizeRequest, RecognizeResponse>):WebSocket {
    if (!process.env.DEEPGRAM_SUBSCRIPTION_KEY) {
        log.error("env DEEPGRAM_SUBSCRIPTION_KEY is undefined");
        throw "env DEEPGRAM_SUBSCRIPTION_KEY is undefined";
    }
    let subscriptionKey: string = process.env.DEEPGRAM_SUBSCRIPTION_KEY!

    var socket:WebSocket = new WebSocket(
        'wss://brain.deepgram.com/v2/listen/stream',
        ['Basic', subscriptionKey]
    );
    log.info("[Deepgram] Connected");
    initializeWS(socket,call);
    return socket
}

//configure websocket behaviour
var initializeWS = function(socket:any, call:grpc.ServerDuplexStream<RecognizeRequest, RecognizeResponse>) {
    socket.onopen = (m:Event) => {
        log.info("[Deepgram] Socket Opened")
    };
    socket.onclose = (m:Event) => {
        log.info("[Deepgram] Socket Closed");
        call.end()
    };
    socket.onmessage = (m:MessageEvent) => {
        log.info("[Deepgram] RECEIVING")
        let data = JSON.parse(m.data);
       
        if (data.hasOwnProperty('channel')) {
                let resp = recognizeResponse(data.channel.alternatives)
                log.info(resp)
                call.write(resp)
        }
    };
};

// convert deepgram response to common response structure
const recognizeResponse = (alternatives:any):pb.RecognizeResponse => {
    let response: pb.RecognizeResponse = new pb.RecognizeResponse()
    let result_list = new Array<pb.SpeechRecognitionResult>()
    let result_item = new pb.SpeechRecognitionResult()

    let alternativeList = alternatives.map((al:any) => {
        let alternative_item = new pb.SpeechRecognitionAlternative()
        alternative_item.setTranscript(al.transcript)
        alternative_item.setConfidence(al.confidence)
        return alternative_item;
    });
    result_item.setAlternativesList(alternativeList)
    result_list.push(result_item)
    response.setResultsList(result_list)
    return response
}

