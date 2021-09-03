package contracts

import "net/http"

type BaseResponse struct {
	Method   string `json:"method"`
	HTTPCode int    `json:"http_code"`
	HTTPText string `json:"http_text"`
}

type SingleResponse struct {
	Msg    string `json:"msg"`
	Status string `json:"status"`
}

func (res *SingleResponse) SetErrorData(err error) *SingleResponse {
	if err == nil {
		res.Msg = "success"
		res.Status = "success"
	}
	res.Status = "failure"
	res.Msg = err.Error()
	return res
}

// SetHTTPCode stets the http code
func (res *BaseResponse) SetHTTPCode(code int) Response {
	res.HTTPCode = code
	return res
}

// SetHTTPText stets the http code
func (res *BaseResponse) SetHTTPText(code int) Response {
	res.HTTPText = http.StatusText(code)
	return res
}

// SetMethod  sets http method
func (res *BaseResponse) SetMethod(method string) Response {
	res.Method = method
	return res
}

type Response interface {
	SetHTTPCode(int) Response
	SetHTTPText(int) Response
	SetMethod(string) Response
}
