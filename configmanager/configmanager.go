package configmanager

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"bitbucket.org/yellowmessenger/asterisk-ari/metrics"
	"bitbucket.org/yellowmessenger/asterisk-ari/queuemanager"
	"bitbucket.org/yellowmessenger/asterisk-ari/utils/model"
	"bitbucket.org/yellowmessenger/asterisk-ari/ymlogger"
	"google.golang.org/api/idtoken"
)

type appconfig struct {
	LoggerConf                  ymlogger.LoggerConf              `json:"logger_conf"`
	MetricsConf                 metrics.Config                   `json:"metrics_conf"`
	QueueConnParams             queuemanager.QueueConnParams     `json:"queue_conn_params"`
	QueueListenerParams         queuemanager.QueueListenerParams `json:"queue_listener_params"`
	QueueMessageParams          queuemanager.QueueMessageParams  `json:"queue_message_params"`
	MySQLUser                   string                           `json:"mysql_user"`
	MySQLPassword               string                           `json:"mysql_password"`
	MySQLDB                     string                           `json:"mysql_db"`
	ARIApplication              string                           `json:"ari_application"`
	ARIUsername                 string                           `json:"ari_username"`
	ARIPassword                 string                           `json:"ari_password"`
	ARIURL                      string                           `json:"ari_url"`
	APIUsername                 string                           `json:"api_username"`
	APIPassword                 string                           `json:"api_password"`
	ARIAPIRetry                 int                              `json:"ari_api_retry"`
	ARIWebsocketURL             string                           `json:"ari_websocket_url"`
	SIPIP                       string                           `json:"sip_ip"`
	SIPCALLTIMEOUT              string                           `json:"sip_call_timeout"`
	SIPPilotNumber              string                           `json:"sip_pilot_number"`
	DialingNumberPrefix         string                           `json:"dialing_number_prefix"`
	CallAPIRequestsPerSecond    int                              `json:"call_api_requests_per_second"`
	TTSFilePath                 string                           `json:"tts_file_path"`
	TTSFrequency                int                              `json:"tts_frequency"`
	TTSFileOutputFormat         string                           `json:"tts_file_output_format"`
	DefaultWelcomeFile          string                           `json:"default_welcome_file"`
	TTSVoiceID                  string                           `json:"tts_voice_id"`
	STTSampleRate               int32                            `json:"stt_sample_rate"`
	STTLanguage                 string                           `json:"stt_language"`
	STTType                     string                           `json:"stt_type"`
	STTStepStreamingType        string                           `json:"stt_step_stream_type"`
	STTInterjectStreamingType   string                           `json:"stt_interject_stream_type"`
	STTEngine                   string                           `json:"stt_engine"`
	STTStreamBufferSize         int                              `json:"stt_stream_buffer_size"`
	STTCancelTimeout            int                              `json:"stt_cancel_timeout"`
	BotEndPoint                 string                           `json:"bot_endpoint"`
	BotTimeoutPeriod            int                              `json:"bot_timeout_period"`
	AnalyticsEndpoint           string                           `json:"analytics_endpoint"`
	LogStoreEndpoint            string                           `json:"logstore_endpoint"`
	GoogleClientID              string                           `json:"google_client_id"`
	GoogleAccessToken           string                           `json:"google_access_token"`
	AzureTokenEndpoint          string                           `json:"azure_token_endpoint"`
	AzureTTSAPIKey              string                           `json:"azure_tts_api_key"`
	AzureTTSAPIKeyNew           string                           `json:"azure_tts_api_key_new"`
	AzureTTSEndpoint            string                           `json:"azure_tts_endpoint"`
	AzureSTTAPIKey              string                           `json:"azure_stt_api_key"`
	AzureSTTAPIKeyNew           string                           `json:"azure_stt_api_key_new"`
	AzureSTTEndpoint            string                           `json:"azure_stt_endpoint"`
	AzureSTTRegion              string                           `json:"azure_stt_region"`
	AzureSpeakerAPIEndpoint     string                           `json:"azure_speaker_api_endpoint"`
	AzureSpeakerAPIKey          string                           `json:"azure_speaker_api_key"`
	YMSTTEndpoint               string                           `json:"ym_stt_endpoint"`
	MicrosoftSDKEndpoint        string                           `json:"microsoft_sdk_endpoint"`
	SpeechSDKEndpoint           string                           `json:"speech_sdk_endpoint"`
	SpeechSDKEndpoints          []string                         `json:"speech_sdk_endpoints"`
	RecordingTerminationKey     string                           `json:"recording_termination_key"`
	RecordingMaxSilence         int                              `json:"recording_max_silence"`
	RecordingMaxDuration        int                              `json:"recording_max_duration"`
	RecordingDirectory          string                           `json:"recording_directory"`
	RecordingFormat             string                           `json:"recording_format"`
	CallRecordingFormat         string                           `json:"call_recording_format"`
	AzureRecordingContainerName string                           `json:"azure_recording_container_name"`
	AzureRecordingStorageCustom model.BotBlobMapping             `json:"recording_storage_custom"`
	DTMFForwardingNumbers       []string                         `json:"dtmf_forwarding_numbers"`
	CallbackMaxTries            int                              `json:"callback_max_tries"`
	InboundCallbackURL          string                           `json:"inbound_callback_url"`
	ContinuousDTMFDelay         int                              `json:"continuous_dtmf_delay"`
	PipeHealthDelay             int                              `json:"pipehealth_delay"`
	CampaignDelayPerCallMS      int                              `json:"campaign_delay_per_call_ms"`
	CampaignMinHour             int                              `json:"campaign_min_hour"`
	CampaignMaxHour             int                              `json:"campaign_max_hour"`
	CountryCode                 string                           `json:"country_code"`
	DefaultRegion               string                           `json:"default_region"`
	RegionCode                  string                           `json:"region_code"`
	ExotelAPIKey                string                           `json:"exotel_api_key"`
	ExotelAPIToken              string                           `json:"exotel_api_token"`
	ExotelAccountSID            string                           `json:"exotel_account_sid"`
	UseAzureSpeechContainer     bool                             `json:"use_azure_speech_container"`
	AzureSpeechContainerUrl     string                           `json:"azure_speech_container_url"`
	BotRateLimitParams          queuemanager.BotRateLimitParams  `json:"bot_rate_limit_params"`
	EnableSpeechLogging         bool                             `json:"enable_speech_logging"`
}

// ConfStore stores the configuration variables
var ConfStore *appconfig

// InitConfig initializes the config
func InitConfig(
	fileName string,
) error {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	if err = json.Unmarshal([]byte(data), &ConfStore); err != nil {
		return err
	}
	return nil
}

// RenewGoogleToken generates Google Token periodically
func RenewGoogleToken(ctx context.Context) {
	ConfStore.GoogleAccessToken = generateToken(ctx)
	ticker := time.NewTicker(30 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ConfStore.GoogleAccessToken = generateToken(ctx)
		}
	}
	return
}

func generateToken(ctx context.Context) string {
	ymlogger.LogDebug("GenToken", "Generating the Google Token")
	ts, err := idtoken.NewTokenSource(ctx, ConfStore.GoogleClientID, idtoken.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	if err != nil {
		ymlogger.LogErrorf("GenToken", "Error while generating the token. Error: [#%v", err)
		return ""
	}
	token, err := ts.Token()
	if err != nil {
		ymlogger.LogErrorf("GenToken", "Error while extracting the token. Error: [#%v", err)
		return ""
	}
	return token.AccessToken
}
