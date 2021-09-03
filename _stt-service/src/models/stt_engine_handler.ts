import * as grpc from 'grpc';
import { RecognizeRequest, RecognizeResponse } from '../proto/msstt/msstt_pb';

export interface EngineHandler {
    onData(req: RecognizeRequest): void;
    onEnd(call: grpc.ServerDuplexStream<RecognizeRequest, RecognizeResponse>): void;
}