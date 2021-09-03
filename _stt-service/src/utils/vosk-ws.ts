import 'dotenv/config';
import pino from "pino";
import * as grpc from 'grpc';
(global as any).WebSocket = require('ws');
import * as pb from '../proto/msstt/msstt_pb';
import { RecognizeResponse, RecognizeRequest } from '../proto/msstt/msstt_pb';

const log = pino({ level: "info" });

//connect to vosk engine 
export var wsstream = function(req : RecognizeRequest, call: grpc.ServerDuplexStream<RecognizeRequest, RecognizeResponse>):WebSocket {
    let uri: string = "ws://localhost:2700"
    var socket:WebSocket = new WebSocket(uri);

    log.info("[Vosk] Connected");
    initializeWS(socket,call, req.getConfig()?.getModel(), req.getConfig()?.getBoostphraseList(),req.getConfig()?.getSampleRateHertz());
    return socket
}

//configure websocket behaviour
var initializeWS = function(socket:any, call:grpc.ServerDuplexStream<RecognizeRequest, RecognizeResponse>, model: string|undefined, boostPhrases: string[]|undefined, sampleHertz: number|undefined) {
    interface NBest {
        model: number;
        phrase_list: string;
        sample_rate: number
    }

    
    socket.onopen = (m:Event) => {
        log.info("[Vosk] Socket Opened")
        let config: any = {}
        if (model != undefined)
            config.model = model
        if (boostPhrases != undefined)
            config.phrase_list = boostPhrases
        if (sampleHertz != undefined)
            config.sample_rate = sampleHertz
        
        let configReq: any = {}
        configReq["config"] = config
        log.info(JSON.stringify(configReq))
        socket.send(JSON.stringify(configReq))
    };

    socket.onclose = (m:Event) => {
        log.info("[Vosk] Socket Closed");
        call.end()
    };

    socket.onmessage = (m:MessageEvent) => {
        log.info("[Vosk] RECEIVING")
        let data = JSON.parse(m.data);
        if (data.hasOwnProperty('result')) {
            let resp = recognizeResponse(data.text)
            log.info(resp)
            call.write(resp)
    }
        console.log(data)
    };

    socket.onerror = (m:MessageEvent) => {
        log.info(`[Vosk] Error ${JSON.stringify(m)}`)
        
    };
};

const recognizeResponse = (text: string): pb.RecognizeResponse => {
    let response: pb.RecognizeResponse = new pb.RecognizeResponse()
    let result_list = new Array<pb.SpeechRecognitionResult>()
    let result_item = new pb.SpeechRecognitionResult()
    let alternative_item = new pb.SpeechRecognitionAlternative()
        alternative_item.setTranscript(text)
        alternative_item.setConfidence(1)
    let alternativeList = new Array<pb.SpeechRecognitionAlternative>()
    alternativeList.push(alternative_item)
    result_item.setAlternativesList(alternativeList)
    result_list.push(result_item)
    response.setResultsList(result_list)
    return response
}