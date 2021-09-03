package contracts

type CreateCall struct {
	SID         string `json:"sid"`
	CreatedTime string `json:"created_time"`
	From        string `json:"from"`
	To          string `json:"to"`
	Status      string `json:"status"`
	CallbackURL string `json:"callback_url"`
	ChannelID   string `json:"channel_id"`
	ServerHost  string `json:"host"`
}

type CreateCallResponse struct {
	BaseResponse
	ResponseData SingleCreateCallResponse `json:"response"`
}

type SingleCreateCallResponse struct {
	SingleResponse
	ResourceData *CreateCall `json:"data,omitempty"`
}
