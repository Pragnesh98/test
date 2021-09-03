package contracts

import (
	"encoding/json"
	"errors"

	"github.com/labstack/echo"
)

type CreateListenCallRequest struct {
	From      *string `json:"from"`
	To        *string `json:"to"`
	PipeType  *string `json:"pipe_type,omitempty"`
	ChannelID *string `json:"channel_id,omitempty"`
}

func (lcr *CreateListenCallRequest) ExtractFromHTTP(c echo.Context) error {
	request := c.Request()
	err := json.NewDecoder(request.Body).Decode(lcr)
	if err != nil {
		return err
	}
	return nil
}

func (lcr *CreateListenCallRequest) Validate() error {
	if lcr.From == nil || len(*lcr.From) <= 0 {
		return errors.New("from parameter is missing or empty")
	}
	if lcr.To == nil || len(*lcr.To) <= 0 {
		return errors.New("to parameter is missing or empty")
	}
	if lcr.PipeType == nil || len(*lcr.PipeType) <= 0 {
		return errors.New("pipe_type parameter is missing or empty")
	}
	if lcr.ChannelID == nil || len(*lcr.ChannelID) <= 0 {
		return errors.New("channel_id parameter is missing or empty")
	}
	return nil
}
