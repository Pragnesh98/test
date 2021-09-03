package contracts

type CreateSIPRegisterResponse struct {
	BaseResponse
	ResponseData SingleCreateSIPRegisterResponse `json:"response"`
}

type SingleCreateSIPRegisterResponse struct {
	SingleResponse
}