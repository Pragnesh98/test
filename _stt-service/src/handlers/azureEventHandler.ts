import "dotenv/config";
import pino from "pino";
import * as grpc from "grpc";
import * as acutil from "../utils/recognizeStream";
import { EngineHandler } from "../models/stt_engine_handler";
import * as mssdk from "microsoft-cognitiveservices-speech-sdk";
import { AudioStreamFormat } from "microsoft-cognitiveservices-speech-sdk";
import {
  RecognizeRequest,
  RecognizeResponse,
  RecognitionConfig,
} from "../proto/msstt/msstt_pb";
import { RecognitionMode } from "../models/recognition_mode";

const log = pino({ level: "info" });

export class Azure implements EngineHandler {
  pushStream: mssdk.PushAudioInputStream;
  sdkRecognizer: mssdk.SpeechRecognizer;
  stt_engine: string | undefined;
  recogntionType: RecognitionConfig.RecognitionType;
  recognitionMode: RecognitionMode;
  callSid: string;
  received_bytes: number;

  constructor(
    req: RecognizeRequest,
    call: grpc.ServerDuplexStream<RecognizeRequest, RecognizeResponse>
  ) {
    this.received_bytes = 0;
    this.pushStream = mssdk.AudioInputStream.createPushStream(
      AudioStreamFormat.getWaveFormatPCM(8000, 16, 1)
    );
    this.callSid = `${req.getConfig()?.getCallsid()}`;
    log.info(
      `CallSID: ${req.getConfig()?.getCallsid()}\n Config: ${req.getConfig()}`
    );

    this.recogntionType = req.getConfig()!.getRecognizeType();
    if (this.recogntionType == RecognitionConfig.RecognitionType.CONTINUOUS) {
      log.info(`[${this.callSid}][Azure] ${this.recogntionType.toString()}`);
      this.recognitionMode = new RecognizeContinuous();
    } else {
      log.info(`[${this.callSid}][Azure] ${this.recogntionType.toString()}`);
      this.recognitionMode = new RecognizeOnce(this.callSid);
    }

    this.sdkRecognizer = acutil.InitializeRecognizer(
      req,
      this.pushStream,
      call,
      this.recognitionMode,
      req.getConfig()!.getCallsid()
    );
    this.recognitionMode.onStreamStart(this.sdkRecognizer);

    if (this.sdkRecognizer == null) {
      log.error(
        `[${this.callSid}][Azure] Azure recognizer couldn't be initialized`
      );
      return;
    }
  }

  onData = (req: RecognizeRequest) => {
    this.received_bytes += req.getAudio()!.getContent().length;
    log.info(
      `[${this.callSid}][Azure] onData - Receiving: ${
        req.getAudio()!.getContent().length
      }. Received total: [${this.received_bytes}]`
    );
    this.pushStream.write(req.getAudio()!.getContent_asU8());
  };

  onEnd = async (
    call: grpc.ServerDuplexStream<RecognizeRequest, RecognizeResponse>
  ) => {
    if (this.pushStream != undefined) {
      log.info(`[${this.callSid}] [Azure] Closing push Stream`);
      this.pushStream.close();
    }
    this.recognitionMode.onStreamEnd(this.sdkRecognizer);
  };
}

class RecognizeOnce implements RecognitionMode {
  constructor(private readonly callSid: string) {}
  onStreamStart = async (
    sdkRecognizer: mssdk.SpeechRecognizer
  ): Promise<boolean> => {
    log.info("[Azure] Starting recognition once");
    sdkRecognizer.recognizeOnceAsync(
      (result: any) => {
        if (result) {
          log.info(`${this.callSid} [Azure] RecoOnceResult: ${result.json}`);
        } else {
          log.info(`${this.callSid} [Azure] RecoOnceResult: Empty Result`);
        }
      },
      (err: string) => {
        log.info(`${this.callSid} ${err}`);
      }
    );
    return true;
  };

  onStreamEnd = async (
    sdkRecognizer: mssdk.SpeechRecognizer
  ): Promise<boolean> => {
    log.info(`${this.callSid} [Azure] onStreamEnd:  Write stream EOF`);
    return true;
  };

  onTerminate = async (
    sdkRecognizer: mssdk.SpeechRecognizer
  ): Promise<boolean> => {
    log.info(
      `${this.callSid}[Azure] Terminating recognition once, closing sdk recognizer`
    );
    sdkRecognizer.close();
    return true;
  };

  onRecognizing = async (
    call: grpc.ServerDuplexStream<RecognizeRequest, RecognizeResponse>,
    response: RecognizeResponse
  ): Promise<boolean> => {
    return true;
  };
}

class RecognizeContinuous implements RecognitionMode {
  onStreamStart = async (
    sdkRecognizer: mssdk.SpeechRecognizer
  ): Promise<boolean> => {
    log.info("[Azure] Starting recognition continuous");
    sdkRecognizer.startContinuousRecognitionAsync();
    return true;
  };

  onStreamEnd = async (
    sdkRecognizer: mssdk.SpeechRecognizer
  ): Promise<boolean> => {
    log.info("[Azure] Write stream EOF");
    return true;
  };

  onTerminate = async (
    sdkRecognizer: mssdk.SpeechRecognizer
  ): Promise<boolean> => {
    log.info("[Azure] Terminating recognition continuous");
    sdkRecognizer.stopContinuousRecognitionAsync();
    return true;
  };

  onRecognizing = async (
    call: grpc.ServerDuplexStream<RecognizeRequest, RecognizeResponse>,
    response: RecognizeResponse
  ): Promise<boolean> => {
    call.write(response);
    return true;
  };
}
