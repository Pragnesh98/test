package requesthandler

import (
	"context"
	"net/http"

	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"bitbucket.org/yellowmessenger/asterisk-ari/core/sipregister"
	"github.com/labstack/echo"
)

// SIPRegisterHandler holds the sip register request
type SIPRegisterHandler struct{}

// Any is hander for SIP register request
func (handler SIPRegisterHandler) Any(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodPost:
		return handler.Create(c)
	}
	return RawResponse(c, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

// Create create the given response
func (SIPRegisterHandler) Create(c echo.Context) error {
	var response *contracts.CreateSIPRegisterResponse
	ctx := context.WithValue(nil, "RequestID", "SIPRegisterRequest")
	var ba contracts.BasicAuthCreds
	err := ba.ExtractFromHTTP(c)
	if err != nil {
		response = new(contracts.CreateSIPRegisterResponse)
		responseData := new(contracts.SingleCreateSIPRegisterResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}
	err = ba.Authenticate()
	if err != nil {
		response = new(contracts.CreateSIPRegisterResponse)
		responseData := new(contracts.SingleCreateSIPRegisterResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}
	srReq := new(contracts.CreateSIPRegisterRequest)
	err = srReq.ExtractFromHTTP(c)
	if err != nil {
		response = new(contracts.CreateSIPRegisterResponse)
		responseData := new(contracts.SingleCreateSIPRegisterResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusBadRequest)
	}
	err = srReq.Validate()
	if err != nil {
		response = new(contracts.CreateSIPRegisterResponse)
		responseData := new(contracts.SingleCreateSIPRegisterResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusBadRequest)
	}
	alredyExists, _ := sipregister.UserAlreadyExists(ctx, *srReq)
	if alredyExists {
		response = new(contracts.CreateSIPRegisterResponse)
		responseData := new(contracts.SingleCreateSIPRegisterResponse)
		responseData.Msg = "User already exists"
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusConflict)
	}
	response, err = sipregister.Create(ctx, *srReq)
	if err != nil {
		response = new(contracts.CreateSIPRegisterResponse)
		responseData := new(contracts.SingleCreateSIPRegisterResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusInternalServerError)
	}
	return Response(c, response, http.StatusOK)
}
