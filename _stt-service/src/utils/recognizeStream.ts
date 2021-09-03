import "dotenv/config";
import pino from "pino";
import * as grpc from "grpc";
import * as pb from "../proto/msstt/msstt_pb";
import * as mssdk from "microsoft-cognitiveservices-speech-sdk";
import { RecognitionMode } from "../models/recognition_mode";

const log = pino({ level: "info" });

export const InitializeRecognizer = (
  request: pb.RecognizeRequest,
  pushStream: mssdk.PushAudioInputStream,
  callback: grpc.ServerDuplexStream<pb.RecognizeRequest, pb.RecognizeResponse>,
  recognitionMode: RecognitionMode,
  callSID: string
): mssdk.SpeechRecognizer => {
  const configuration = request.getConfig()!;

  let subscriptionKey: string = request
    .getSttServiceOptions()
    ?.getAzureOptions()
    ?.getFromSubscription()
    ?.getSubscriptionKey()!;
  let region: string = request
    .getSttServiceOptions()
    ?.getAzureOptions()
    ?.getFromSubscription()
    ?.getRegion()!;

  if (!subscriptionKey) {
    if (!process.env.MSSDK_SPEECH_SUBSCRIPTION_KEY) {
      log.error("env MSSDK_SPEECH_SUBSCRIPTON_KEY is undefined");
      throw "env MSSDK_SPEECH_SUBSCRIPTION_KEY is undefined";
    }
    subscriptionKey = process.env.MSSDK_SPEECH_SUBSCRIPTION_KEY;
  }
  const speechConfig = mssdk.SpeechConfig.fromSubscription(
    subscriptionKey,
    region
  );
  const detectLanguageFlag = configuration.getDetectLanguagesList() && configuration.getDetectLanguagesList().length > 0

  speechConfig.speechRecognitionLanguage = configuration.getLanguageCode();
  speechConfig.outputFormat = mssdk.OutputFormat.Detailed;
  speechConfig.setProperty(
    mssdk.PropertyId[
    mssdk.PropertyId.SpeechServiceConnection_InitialSilenceTimeoutMs
    ],
    (2 * 1000 * configuration.getInitialSilenceTimeout()).toString()
  );
  speechConfig.setProperty(
    mssdk.PropertyId[
    mssdk.PropertyId.SpeechServiceConnection_EndSilenceTimeoutMs
    ],
    (1000).toString()
  );

  const audioConfig = mssdk.AudioConfig.fromStreamInput(pushStream);
  let autoDetectConfig: mssdk.AutoDetectSourceLanguageConfig;
  let sdkRecognizer: mssdk.SpeechRecognizer

  if (detectLanguageFlag) {
    autoDetectConfig = mssdk.AutoDetectSourceLanguageConfig.fromLanguages(configuration.getDetectLanguagesList());
    sdkRecognizer = mssdk.SpeechRecognizer.FromConfig(speechConfig, autoDetectConfig, audioConfig);
  }
  else {
    sdkRecognizer = new mssdk.SpeechRecognizer(speechConfig, audioConfig)
  }

  if (configuration.getMsEndpoint()) {
    log.info(`[${callSID}] using endpoint id ${configuration.getMsEndpoint()}`);
    speechConfig.endpointId = configuration.getMsEndpoint();
  }

  SetUpRecognizer(sdkRecognizer, callback, recognitionMode, callSID, detectLanguageFlag);

  var phraseListGrammar = mssdk.PhraseListGrammar.fromRecognizer(sdkRecognizer);
  var phraseList = configuration.getBoostphraseList();

  if (phraseList && phraseList.length > 0) {
    log.info(`[${callSID}] Phrases list found. adding boost phrases`);
    phraseList.forEach((phrase: string) => {
      phraseListGrammar.addPhrase(phrase);
    });
  }
  return sdkRecognizer;
};

const SetUpRecognizer = (
  sdkRecognizer: mssdk.SpeechRecognizer,
  call: grpc.ServerDuplexStream<pb.RecognizeRequest, pb.RecognizeResponse>,
  recognitionMode: RecognitionMode,
  callSID: string,
  detectLanguageFlag: boolean
) => {
  let res: pb.RecognizeResponse = new pb.RecognizeResponse();
  let recognized: boolean = false;
  sdkRecognizer.sessionStarted = (s, e) => {
    log.info(`${callSID} SessionStarted. SessionId: ${e.sessionId}`);
  };
  sdkRecognizer.speechStartDetected = (s, e) => {
    log.info(`${callSID} [${e.sessionId}]SpeechStartDetected`);
  };
  sdkRecognizer.speechEndDetected = (s, e) => {
    log.info(`${callSID}  [${e.sessionId}] SpeechEndDetected`);
  };
  sdkRecognizer.recognizing = (s, e) => {
    log.info(`[${callSID}]  [${e.sessionId}] [AzureEvent] RECOGNIZING`);

    let fmtResult = recognizeResponse(e, callSID, detectLanguageFlag);
    res = fmtResult;
    recognitionMode.onRecognizing(call, res);
  };

  sdkRecognizer.recognized = (s, e) => {
    if (e.result.reason == mssdk.ResultReason.RecognizedSpeech) {
      let res: pb.RecognizeResponse = new pb.RecognizeResponse();
      log.info(`[${callSID}] [${e.sessionId}] [AzureEvent] RECOGNIZED RESULTS.`);

      //format results
      let fmtResult = recognizeResponse(e, callSID, detectLanguageFlag);
      log.debug(
        `[${callSID}] [AzureEvent] New Response: ${JSON.stringify(fmtResult)}`
      );
      res = fmtResult;
      call.write(res);

      log.info(`[${callSID}] [AzureEvent] RECOGNIZED`);
    } else if (e.result.reason == mssdk.ResultReason.NoMatch) {
      call.write(res);
      log.error(
        `[${callSID}]  [${e.sessionId}] [AzureEvent] NOMATCH: Speech could not be recognized.`
      );
    }
    recognized = true;
    recognitionMode.onTerminate(sdkRecognizer);
  };

  sdkRecognizer.canceled = (s, e) => {
    if (e.reason == mssdk.CancellationReason.Error) {
      log.error(
        `[${callSID}] [${e.sessionId}] [AzureEvent] CANCELED: Reason=${e.reason} ErrorCode=${e.errorCode}; Details=${e.errorDetails}`
      );
    }
    recognitionMode.onTerminate(sdkRecognizer);
    call.end();
  };

  sdkRecognizer.sessionStopped = (s, e) => {
    if (!recognized) {
      call.write(res);
    }
    log.error(`[${callSID}] [${e.sessionId}][AzureEvent] Session stopped event.`);
    call.end();
  };
};

const recognizeResponse = (
  sdkResult: mssdk.SpeechRecognitionEventArgs,
  callSID: string,
  detectLanguageFlag: boolean,
): pb.RecognizeResponse => {
  interface NBest {
    Confidence: number;
    Display: string;
  }

  let response: pb.RecognizeResponse = new pb.RecognizeResponse();
  let results = new Array<pb.SpeechRecognitionResult>();
  let result = new pb.SpeechRecognitionResult();
  let alternatives = new Array<pb.SpeechRecognitionAlternative>();

  if (sdkResult && sdkResult.result && sdkResult.result.json) {
    if (detectLanguageFlag) {
      let languageDetectionResult = mssdk.AutoDetectSourceLanguageResult.fromResult(sdkResult.result);
      log.info(`Detected language: ${languageDetectionResult.language}`)
      response.setDetectedLanguage(languageDetectionResult.language)
    }
    const sdkJson = JSON.parse(sdkResult.result.json);

    response.setRecognitionstatus(sdkJson.RecognitionStatus);

    result.setOffset(sdkResult.result.offset);
    result.setDuration(sdkResult.result.duration);

    if (sdkJson.DisplayText) {
      let displayAlternative: pb.SpeechRecognitionAlternative = new pb.SpeechRecognitionAlternative();
      displayAlternative.setConfidence(1);
      displayAlternative.setTranscript(sdkJson.DisplayText);
      log.info(`[${callSID}] [AzureEvent] Formatting: ${sdkJson.DisplayText}`);
      alternatives.push(displayAlternative);
    }
    else {
      if (sdkJson.Text) {
        let alternative: pb.SpeechRecognitionAlternative = new pb.SpeechRecognitionAlternative();
        alternative.setConfidence((sdkJson.Offset + sdkJson.Duration) / 10000000);
        alternative.setTranscript(sdkJson.Text);
        log.info(`[${callSID}] [AzureEvent] Formatting: ${sdkJson.Text}`);
        alternatives.push(alternative);
      }

      if (sdkJson.NBest) {
        const nbest: NBest[] = Array.from(sdkJson.NBest);
        alternatives = nbest.map<pb.SpeechRecognitionAlternative>((nb: NBest) => {
          let alternative: pb.SpeechRecognitionAlternative = new pb.SpeechRecognitionAlternative();
          alternative.setConfidence(nb.Confidence);
          alternative.setTranscript(nb.Display);
          log.info(
            `[${callSID}] [AzureEvent] Alternative ${alternative.getTranscript()}`
          );
          return alternative;
        });
      }
    }

    result.setAlternativesList(alternatives);
    results.push(result);
    response.setResultsList(results);
  } else {
    log.error(
      `${callSID} unable to parse response in recognizeResponse. ${sdkResult}`
    );
  }
  return response;
};
