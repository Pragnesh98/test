package contracts

import (
	"encoding/json"
	"errors"

	"github.com/labstack/echo"
)

type DeleteBargeINCallRequest struct {
	ChannelID *string `json:"channel_id"`
}

func (dbcr *DeleteBargeINCallRequest) ExtractFromHTTP(c echo.Context) error {
	request := c.Request()
	err := json.NewDecoder(request.Body).Decode(dbcr)
	if err != nil {
		return err
	}
	return nil
}

func (dbcr *DeleteBargeINCallRequest) Validate() error {
	if dbcr.ChannelID == nil || len(*dbcr.ChannelID) <= 0 {
		return errors.New("channel_id parameter is missing or empty")
	}
	return nil
}
