package contracts

type DeleteListenCallResponse struct {
	BaseResponse
	ResponseData SingleDeleteListenCallResponse `json:"response"`
}

type SingleDeleteListenCallResponse struct {
	SingleResponse
}
