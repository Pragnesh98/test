package requesthandler

import (
	"context"
	"net/http"

	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"bitbucket.org/yellowmessenger/asterisk-ari/core/bargeincall"
	"github.com/labstack/echo"
)

type BargeINHandler struct{}

func (handler BargeINHandler) Any(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodPost:
		return handler.Create(c)
	case http.MethodDelete:
		return handler.Delete(c)
	}
	return RawResponse(c, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

func (BargeINHandler) Create(c echo.Context) error {
	var response *contracts.CreateBargeINCallResponse
	ctx := context.WithValue(nil, "RequestID", "CreateBargeINCall")
	var ba contracts.BasicAuthCreds
	err := ba.ExtractFromHTTP(c)
	if err != nil {
		response = new(contracts.CreateBargeINCallResponse)
		responseData := new(contracts.SingleCreateBargeINCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}
	err = ba.Authenticate()
	if err != nil {
		response = new(contracts.CreateBargeINCallResponse)
		responseData := new(contracts.SingleCreateBargeINCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}
	bargeReq := new(contracts.CreateBargeINCallRequest)
	err = bargeReq.ExtractFromHTTP(c)
	if err != nil {
		response = new(contracts.CreateBargeINCallResponse)
		responseData := new(contracts.SingleCreateBargeINCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusBadRequest)
	}
	err = bargeReq.Validate()
	if err != nil {
		response = new(contracts.CreateBargeINCallResponse)
		responseData := new(contracts.SingleCreateBargeINCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusBadRequest)
	}
	response, err = bargeincall.Create(ctx, *bargeReq)
	if err != nil {
		response = new(contracts.CreateBargeINCallResponse)
		responseData := new(contracts.SingleCreateBargeINCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusInternalServerError)
	}
	return Response(c, response, http.StatusOK)
}

func (BargeINHandler) Delete(c echo.Context) error {
	var response *contracts.DeleteBargeINCallResponse
	ctx := context.WithValue(nil, "RequestID", "DeleteBargeINCall")
	var ba contracts.BasicAuthCreds
	err := ba.ExtractFromHTTP(c)
	if err != nil {
		response = new(contracts.DeleteBargeINCallResponse)
		responseData := new(contracts.SingleDeleteBargeINCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}
	err = ba.Authenticate()
	if err != nil {
		response = new(contracts.DeleteBargeINCallResponse)
		responseData := new(contracts.SingleDeleteBargeINCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}
	bargeReq := new(contracts.DeleteBargeINCallRequest)
	err = bargeReq.ExtractFromHTTP(c)
	if err != nil {
		response = new(contracts.DeleteBargeINCallResponse)
		responseData := new(contracts.SingleDeleteBargeINCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusBadRequest)
	}
	err = bargeReq.Validate()
	if err != nil {
		response = new(contracts.DeleteBargeINCallResponse)
		responseData := new(contracts.SingleDeleteBargeINCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusBadRequest)
	}
	response, err = bargeincall.Delete(ctx, *bargeReq)
	if err != nil {
		response = new(contracts.DeleteBargeINCallResponse)
		responseData := new(contracts.SingleDeleteBargeINCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusInternalServerError)
	}
	return Response(c, response, http.StatusOK)
}
