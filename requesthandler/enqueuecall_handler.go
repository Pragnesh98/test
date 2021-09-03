package requesthandler

import (
	"net/http"

	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"bitbucket.org/yellowmessenger/asterisk-ari/core/enqueuecall"
	"github.com/labstack/echo"
)

type EnqueueCallHandler struct{}

func (handler EnqueueCallHandler) Any(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodPost:
		return handler.Create(c)
	}
	return RawResponse(c, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

func (EnqueueCallHandler) Create(c echo.Context) error {
	var response *contracts.EnqueueCallResponse
	var ba contracts.BasicAuthCreds
	err := ba.ExtractFromHTTP(c)
	if err != nil {
		response = new(contracts.EnqueueCallResponse)
		responseData := new(contracts.SingleEnqueueCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}
	err = ba.Authenticate()
	if err != nil {
		response = new(contracts.EnqueueCallResponse)
		responseData := new(contracts.SingleEnqueueCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}
	ecReq := new(contracts.EnqueueCallRequest)
	err = ecReq.ExtractFromHTTP(c)
	if err != nil {
		response = new(contracts.EnqueueCallResponse)
		responseData := new(contracts.SingleEnqueueCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusBadRequest)
	}
	err = ecReq.Validate()
	if err != nil {
		response = new(contracts.EnqueueCallResponse)
		responseData := new(contracts.SingleEnqueueCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusBadRequest)
	}
	// dndActive, err := enqueuecall.CheckDNDStatus(*ecReq.To)
	// if dndActive || err != nil {
	// 	response = new(contracts.EnqueueCallResponse)
	// 	responseData := new(contracts.SingleEnqueueCallResponse)
	// 	responseData.SingleResponse.SetErrorData(err)
	// 	response.ResponseData = *responseData
	// 	return Response(c, response, http.StatusForbidden)
	// }
	// HACK for TELEPERFORMANCE to reduce the failures
	// if *ecReq.From == "+918068402350" || *ecReq.From == "+918068402351" || *ecReq.From == "+918068402352" {
	// 	botUp, err := bothelper.CheckIfBotUp(
	// 		nil,
	// 		"",
	// 		"",
	// 		"welcome",
	// 		*ecReq.To,
	// 		*ecReq.From,
	// 		"en",
	// 		"outgoing",
	// 		"",
	// 		"en-IN",
	// 	)
	// 	if !botUp || err != nil {
	// 		response = new(contracts.EnqueueCallResponse)
	// 		responseData := new(contracts.SingleEnqueueCallResponse)
	// 		responseData.SingleResponse.SetErrorData(err)
	// 		response.ResponseData = *responseData
	// 		return Response(c, response, http.StatusForbidden)
	// 	}
	// }
	response, err = enqueuecall.Create(*ecReq)
	if err != nil {
		response = new(contracts.EnqueueCallResponse)
		responseData := new(contracts.SingleEnqueueCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusInternalServerError)
	}
	return Response(c, response, http.StatusOK)
}
