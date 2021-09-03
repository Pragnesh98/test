package contracts

type EnqueueCall struct {
	SID         string `json:"sid"`
	CreatedTime string `json:"created_time"`
	From        string `json:"from"`
	To          string `json:"to"`
	Status      string `json:"status"`
	CallbackURL string `json:"callback_url"`
	BotID       string `json:"botId"`
	CampaignID  string `json:"campaignID"`
	Host        string `json:"host"`
}

type EnqueueCallResponse struct {
	BaseResponse
	ResponseData SingleEnqueueCallResponse `json:"response"`
}

type SingleEnqueueCallResponse struct {
	SingleResponse
	ResourceData *EnqueueCall `json:"data,omitempty"`
}
