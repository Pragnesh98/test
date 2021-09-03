package requesthandler

import (
	"context"
	"net/http"

	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"bitbucket.org/yellowmessenger/asterisk-ari/core/listencall"
	"github.com/labstack/echo"
)

type ListenCallHandler struct{}

func (handler ListenCallHandler) Any(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodPost:
		return handler.Create(c)
	case http.MethodDelete:
		return handler.Delete(c)
	}
	return RawResponse(c, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

func (ListenCallHandler) Create(c echo.Context) error {
	var response *contracts.CreateListenCallResponse
	ctx := context.WithValue(nil, "RequestID", "CreateListenCall")
	var ba contracts.BasicAuthCreds
	err := ba.ExtractFromHTTP(c)
	if err != nil {
		response = new(contracts.CreateListenCallResponse)
		responseData := new(contracts.SingleCreateListenCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}
	err = ba.Authenticate()
	if err != nil {
		response = new(contracts.CreateListenCallResponse)
		responseData := new(contracts.SingleCreateListenCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}
	lcReq := new(contracts.CreateListenCallRequest)
	err = lcReq.ExtractFromHTTP(c)
	if err != nil {
		response = new(contracts.CreateListenCallResponse)
		responseData := new(contracts.SingleCreateListenCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusBadRequest)
	}
	err = lcReq.Validate()
	if err != nil {
		response = new(contracts.CreateListenCallResponse)
		responseData := new(contracts.SingleCreateListenCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusBadRequest)
	}
	response, err = listencall.Create(ctx, *lcReq)
	if err != nil {
		response = new(contracts.CreateListenCallResponse)
		responseData := new(contracts.SingleCreateListenCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusInternalServerError)
	}
	return Response(c, response, http.StatusOK)
}

func (ListenCallHandler) Delete(c echo.Context) error {
	var response *contracts.DeleteListenCallResponse
	ctx := context.WithValue(nil, "RequestID", "DeleteListenCall")
	var ba contracts.BasicAuthCreds
	err := ba.ExtractFromHTTP(c)
	if err != nil {
		response = new(contracts.DeleteListenCallResponse)
		responseData := new(contracts.SingleDeleteListenCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}
	err = ba.Authenticate()
	if err != nil {
		response = new(contracts.DeleteListenCallResponse)
		responseData := new(contracts.SingleDeleteListenCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}
	lcReq := new(contracts.DeleteListenCallRequest)
	err = lcReq.ExtractFromHTTP(c)
	if err != nil {
		response = new(contracts.DeleteListenCallResponse)
		responseData := new(contracts.SingleDeleteListenCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusBadRequest)
	}
	err = lcReq.Validate()
	if err != nil {
		response = new(contracts.DeleteListenCallResponse)
		responseData := new(contracts.SingleDeleteListenCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusBadRequest)
	}
	response, err = listencall.Delete(ctx, *lcReq)
	if err != nil {
		response = new(contracts.DeleteListenCallResponse)
		responseData := new(contracts.SingleDeleteListenCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusInternalServerError)
	}
	return Response(c, response, http.StatusOK)
}
