package contracts

type DeleteBargeINCallResponse struct {
	BaseResponse
	ResponseData SingleDeleteBargeINCallResponse `json:"response"`
}

type SingleDeleteBargeINCallResponse struct {
	SingleResponse
}
