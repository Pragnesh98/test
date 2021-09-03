package requesthandler

import (
	"context"
	"errors"
	"net/http"

	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"bitbucket.org/yellowmessenger/asterisk-ari/core/createcall"
	"github.com/labstack/echo"
	"golang.org/x/time/rate"
)

const CreateCallAPIRequestsPerSecond = 5
const CreateCallAPIRequestsPerMinute = 300

var createCallLimiter = rate.NewLimiter(CreateCallAPIRequestsPerSecond, CreateCallAPIRequestsPerMinute)

type CreateCallHandler struct{}

func (handler CreateCallHandler) Any(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodPost:
		return handler.Create(c)
	}
	return RawResponse(c, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

func (CreateCallHandler) Create(c echo.Context) error {
	var response *contracts.CreateCallResponse
	ctx := context.WithValue(nil, "RequestID", "CreateCall")
	if createCallLimiter.Allow() == false {
		response = new(contracts.CreateCallResponse)
		responseData := new(contracts.SingleCreateCallResponse)
		err := errors.New("Making more than allowed requests")
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusTooManyRequests)
	}
	var ba contracts.BasicAuthCreds
	err := ba.ExtractFromHTTP(c)
	if err != nil {
		response = new(contracts.CreateCallResponse)
		responseData := new(contracts.SingleCreateCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}
	err = ba.Authenticate()
	if err != nil {
		response = new(contracts.CreateCallResponse)
		responseData := new(contracts.SingleCreateCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusUnauthorized)
	}

	ccReq := new(contracts.CreateCallRequest)
	err = ccReq.ExtractFromHTTP(c)
	if err != nil {
		response = new(contracts.CreateCallResponse)
		responseData := new(contracts.SingleCreateCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusBadRequest)
	}
	err = ccReq.Validate()
	if err != nil {
		response = new(contracts.CreateCallResponse)
		responseData := new(contracts.SingleCreateCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusBadRequest)
	}

	dndActive, err := createcall.CheckDNDStatus(*ccReq.To)
	if dndActive || err != nil {
		response = new(contracts.CreateCallResponse)
		responseData := new(contracts.SingleCreateCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusForbidden)
	}

	response, err = createcall.Create(ctx, *ccReq)
	if err != nil {
		response = new(contracts.CreateCallResponse)
		responseData := new(contracts.SingleCreateCallResponse)
		responseData.SingleResponse.SetErrorData(err)
		response.ResponseData = *responseData
		return Response(c, response, http.StatusInternalServerError)
	}
	return Response(c, response, http.StatusOK)
}
