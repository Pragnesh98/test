export type RecognizeRequest = {

    audio: {
      AudioSource: {
        Content: Uint8Array;
      };
    };
    config: {
      sample_rate_hertz: number;
      language_code: string;
      encoding: string;
      [k: string]: unknown;
    };
    options?: {
      azure_options?: {
        endpoint_id?: string;
      };
    };
}

