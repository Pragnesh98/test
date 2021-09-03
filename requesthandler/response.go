package requesthandler

import (
	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"github.com/labstack/echo"
)

// CreateCallResponse holds the response for the create call request
type CreateCallResponse struct {
	SID         string `json:"sid"`
	CreatedTime string `json:"created_time"`
	From        string `json:"from"`
	To          string `json:"to"`
	Status      string `json:"status"`
	CallbackURL string `json:"callback_url"`
}

// EnqueueCallResponse holds the response for the enqueue call request
type EnqueueCallResponse struct {
	SID         string `json:"sid"`
	CreatedTime string `json:"created_time"`
	From        string `json:"from"`
	To          string `json:"to"`
	Status      string `json:"status"`
	CallbackURL string `json:"callback_url"`
}

func Response(c echo.Context, response contracts.Response, httpCode int) error {
	response.SetHTTPCode(httpCode)
	response.SetHTTPText(httpCode)
	response.SetMethod(c.Request().Method)
	return RawResponse(c, response, httpCode)
}

func RawResponse(c echo.Context, response interface{}, httpCode int) error {
	return c.JSON(httpCode, response)
}
