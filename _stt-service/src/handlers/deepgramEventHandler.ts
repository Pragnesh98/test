import 'dotenv/config';
import * as grpc from 'grpc';
import * as dpg from '../utils/deepgram'
import * as vosk from '../utils/vosk-ws'

import {EngineHandler} from '../models/stt_engine_handler'
import { RecognizeRequest, RecognizeResponse } from '../proto/msstt/msstt_pb';

import pino from "pino";

const log = pino({ level: "info" });

export class Deepgram implements EngineHandler{
    socket: WebSocket;
    
    constructor(req : RecognizeRequest, call: grpc.ServerDuplexStream<RecognizeRequest, RecognizeResponse> ){
        this.socket = dpg.wsstream(call)
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
        log.info("[Deepgram] Closing connection")
        this.socket.send(new Uint8Array(0))
    }
}