package contracts

import (
	"encoding/json"
	"errors"

	"bitbucket.org/yellowmessenger/asterisk-ari/bothelper"
	"github.com/labstack/echo"
)

type EnqueueCallRequest struct {
	From               *string               `json:"from"`
	To                 *string               `json:"to"`
	CallbackURL        *string               `json:"callback_url,omitempty"`
	PipeType           *string               `json:"pipe_type,omitempty"`
	RecordingFileName  *string               `json:"recording_file_name,omitempty"`
	MaxBotFailureCount *int8                 `json:"max_bot_failure_count,omitempty"`
	TTSOptions         *bothelper.TTSOptions `json:"tts_options,omitempty"`
	ExtraParams        interface{}           `json:"extra_params,omitempty"`
}

// ExtraParams contains the extra params from the voice job
type ExtraParams struct {
	BotID      string `json:"botId"`
	CampaignID string `json:"campaignId"`
}

func (ccr *EnqueueCallRequest) ExtractFromHTTP(c echo.Context) error {
	request := c.Request()
	err := json.NewDecoder(request.Body).Decode(ccr)
	if err != nil {
		return err
	}
	return nil
}

func (eP *ExtraParams) ExtractExtraParams(extraParams interface{}) error {
	strExtraParams, err := json.Marshal(extraParams)
	if err != nil {
		return err
	}
	// strParams := string(strExtraParams).(string)
	if err := json.Unmarshal([]byte(strExtraParams), &eP); err != nil {
		return err
	}
	return nil
}

func (ecr *EnqueueCallRequest) Validate() error {
	if ecr.From == nil || len(*ecr.From) <= 0 {
		return errors.New("from parameter is missing or empty")
	}
	if ecr.To == nil || len(*ecr.To) <= 0 {
		return errors.New("to parameter is missing or empty")
	}
	if ecr.PipeType == nil || len(*ecr.PipeType) <= 0 {
		return errors.New("pipe_type parameter is missing or empty")
	}
	return nil
}
