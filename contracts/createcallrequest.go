package contracts

import (
	"encoding/json"
	"errors"

	"github.com/labstack/echo"
)

type CreateCallRequest struct {
	From              *string `json:"from"`
	To                *string `json:"to"`
	CallbackURL       *string `json:"callback_url,omitempty"`
	PipeType          *string `json:"pipe_type,omitempty"`
	RecordingFileName *string `json:"recording_file_name,omitempty"`
}

func (ccr *CreateCallRequest) ExtractFromHTTP(c echo.Context) error {
	request := c.Request()
	err := json.NewDecoder(request.Body).Decode(ccr)
	if err != nil {
		return err
	}
	return nil
}

func (ccr *CreateCallRequest) Validate() error {
	if ccr.From == nil || len(*ccr.From) <= 0 {
		return errors.New("from parameter is missing or empty")
	}
	if ccr.To == nil || len(*ccr.To) <= 0 {
		return errors.New("to parameter is missing or empty")
	}
	if ccr.PipeType == nil || len(*ccr.PipeType) <= 0 {
		return errors.New("pipe_type parameter is missing or empty")
	}
	return nil
}
