package main

import (
	"bitbucket.org/yellowmessenger/asterisk-ari/requesthandler"

	"github.com/labstack/echo"
)

// AddRoutes defines the routes and the handlers
func AddRoutes(e *echo.Echo) {
	e.Any("/createcall", requesthandler.CreateCallHandler{}.Any)
	e.Any("/enqueuecall", requesthandler.EnqueueCallHandler{}.Any)
	e.Any("/listen", requesthandler.ListenCallHandler{}.Any)
	e.Any("/bargein", requesthandler.BargeINHandler{}.Any)
	e.Any("/sip/register", requesthandler.SIPRegisterHandler{}.Any)
	e.Any("/pipehealth", requesthandler.PipeHealthHandler{}.Any)
}
