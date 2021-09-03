import 'dotenv/config';
import pino from "pino";
import * as grpc from 'grpc';
import { Azure } from './azureEventHandler'
import { Deepgram } from './deepgramEventHandler'
import { VoskClient } from './voskEventHandler'
import { STTService, ISTTServer } from '../proto/msstt/msstt_grpc_pb';
import { RecognizeRequest, RecognizeResponse, RecognitionConfig } from '../proto/msstt/msstt_pb';

import { EngineHandler } from '../models/stt_engine_handler'
const log = pino({ level: "info" });
class STTHandler implements ISTTServer {
    /**
     * Greet the user nicely
     * @param call
     * @param callback
     */

    streamSpeechToText = (call: grpc.ServerDuplexStream<RecognizeRequest, RecognizeResponse>) => {
        log.info("[StreamSTTHandler] gRPC connections established")
        let stt_engine: string | undefined;
        // engine handler interface object
        let engineHandler: EngineHandler
        call.on('data', (req: RecognizeRequest) => {
            if (req.getAudio() === undefined) {
                switch (req.getConfig()?.getSttEngine().toLowerCase()) {
                    case 'azure':
                        engineHandler = new Azure(req, call)
                        log.info("[StreamSTTHandler] Engine: Azure")
                        break;
                    case 'deepgram':
                        engineHandler = new Deepgram(req, call) 
                        break;
                    case 'vosk':
                        engineHandler = new VoskClient(req, call)
                        break;
                    default:
                        log.error(`STT engine ${stt_engine} NOT FOUND`)
                        break;
                }
                return
            }
            if (engineHandler === undefined) {
                log.error(`STT engine handler UNDEFINED`)
                return;
            }
            engineHandler.onData(req)
        })
            .on('end', () => {
                log.info("[StreamSTTHandler] Stream EOF from client")
                if (engineHandler === undefined) {
                    log.error(`[StreamSTTHandler] STT engine handler UNDEFINED`)
                    return;
                }
                engineHandler.onEnd(call)
            });
    };
};

export default {
    service: STTService,
    handler: new STTHandler(),
};
