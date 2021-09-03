package contracts

import (
	"encoding/json"
	"errors"

	"github.com/labstack/echo"
)

type CreateBargeINCallRequest struct {
	From      *string `json:"from"`
	To        *string `json:"to"`
	PipeType  *string `json:"pipe_type,omitempty"`
	ChannelID *string `json:"channel_id,omitempty"`
}

func (cbr *CreateBargeINCallRequest) ExtractFromHTTP(c echo.Context) error {
	request := c.Request()
	err := json.NewDecoder(request.Body).Decode(cbr)
	if err != nil {
		return err
	}
	return nil
}

func (cbr *CreateBargeINCallRequest) Validate() error {
	if cbr.From == nil || len(*cbr.From) <= 0 {
		return errors.New("from parameter is missing or empty")
	}
	if cbr.To == nil || len(*cbr.To) <= 0 {
		return errors.New("to parameter is missing or empty")
	}
	if cbr.PipeType == nil || len(*cbr.PipeType) <= 0 {
		return errors.New("pipe_type parameter is missing or empty")
	}
	if cbr.ChannelID == nil || len(*cbr.ChannelID) <= 0 {
		return errors.New("channel_id parameter is missing or empty")
	}
	return nil
}
