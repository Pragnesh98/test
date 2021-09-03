package requesthandler

import (
	"context"
	"net/http"

	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"bitbucket.org/yellowmessenger/asterisk-ari/core/pipehealth"
	"github.com/labstack/echo"
)

type PipeHealthHandler struct{}

func (handler PipeHealthHandler) Any(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodPost:
		return handler.Post(c)
	}

	return RawResponse(c, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

func (PipeHealthHandler) Post(c echo.Context) error {
	var response *contracts.GetPipeHealthResponse
	ctx := context.WithValue(nil, "RequestID", "CreateCall")
	var ba contracts.BasicAuthCreds
	err := ba.ExtractFromHTTP(c)
	if err != nil {
		response = new(contracts.GetPipeHealthResponse)
		responseData := new(contracts.SingleGetPipeHealthResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}
	err = ba.Authenticate()
	if err != nil {
		response = new(contracts.GetPipeHealthResponse)
		responseData := new(contracts.SingleGetPipeHealthResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}
	response, err = pipehealth.Get(ctx)
	if err != nil {
		response = new(contracts.GetPipeHealthResponse)
		responseData := new(contracts.SingleGetPipeHealthResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusInternalServerError)
	}
	return Response(c, response, http.StatusOK)
}
