import 'dotenv/config';
import * as grpc from 'grpc';
import * as vosk from '../utils/vosk-ws'

import {EngineHandler} from '../models/stt_engine_handler'
import { RecognizeRequest, RecognizeResponse } from '../proto/msstt/msstt_pb';

import pino from "pino";

const log = pino({ level: "info" });

export class VoskClient implements EngineHandler{
    socket: WebSocket;
    
    constructor(req : RecognizeRequest, call: grpc.ServerDuplexStream<RecognizeRequest, RecognizeResponse> ){
        this.socket = vosk.wsstream(req, call)
    }

    onData = (req : RecognizeRequest) => {
        let retry = 5
        while(retry > 0) {
            if(this.socket.readyState == 1){
                this.socket.send(req.getAudio()!.getContent_asU8())
        }
        else {
            log.info("Not connected yet")
            }
        retry -= 1
        }
    }

    onEnd = async (call: grpc.ServerDuplexStream<RecognizeRequest, RecognizeResponse>) => {
        log.info("[VoskClient] Closing connection")
        this.socket.send('{"eof" : 1}');
    }
}