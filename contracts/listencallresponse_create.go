package contracts

type CreateListenCall struct {
	SID             string `json:"sid"`
	CreatedTime     string `json:"created_time"`
	From            string `json:"from"`
	To              string `json:"to"`
	Status          string `json:"status"`
	ChannelID       string `json:"channel_id"`
	ListenChannelID string `json:"listen_channel_id"`
	ServerHost      string `json:"server_host"`
}

type CreateListenCallResponse struct {
	BaseResponse
	ResponseData SingleCreateListenCallResponse `json:"response"`
}

type SingleCreateListenCallResponse struct {
	SingleResponse
	ResourceData *CreateListenCall `json:"data,omitempty"`
}
