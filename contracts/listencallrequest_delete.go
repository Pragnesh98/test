package contracts

import (
	"encoding/json"
	"errors"

	"github.com/labstack/echo"
)

type DeleteListenCallRequest struct {
	ChannelID *string `json:"channel_id"`
}

func (dcr *DeleteListenCallRequest) ExtractFromHTTP(c echo.Context) error {
	request := c.Request()
	err := json.NewDecoder(request.Body).Decode(dcr)
	if err != nil {
		return err
	}
	return nil
}

func (dcr *DeleteListenCallRequest) Validate() error {
	if dcr.ChannelID == nil || len(*dcr.ChannelID) <= 0 {
		return errors.New("channel_id parameter is missing or empty")
	}
	return nil
}
