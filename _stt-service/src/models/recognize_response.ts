interface SpeechRecognitionAlternative {
  transcript: string;
  confidence: number;
}

interface SpeechRecognitionResult {
  alternatives: SpeechRecognitionAlternative[];
}

export type RecognizeResponse = {
  
    results: SpeechRecognitionResult[];
  
}

