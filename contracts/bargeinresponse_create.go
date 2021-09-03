package contracts

type CreateBargeINCall struct {
	SID              string `json:"sid"`
	CreatedTime      string `json:"created_time"`
	From             string `json:"from"`
	To               string `json:"to"`
	Status           string `json:"status"`
	ChannelID        string `json:"channel_id"`
	BargeINChannelID string `json:"barge_in_channel_id"`
	ServerHost       string `json:"server_host"`
}

type CreateBargeINCallResponse struct {
	BaseResponse
	ResponseData SingleCreateBargeINCallResponse `json:"response"`
}

type SingleCreateBargeINCallResponse struct {
	SingleResponse
	ResourceData *CreateBargeINCall `json:"data,omitempty"`
}
