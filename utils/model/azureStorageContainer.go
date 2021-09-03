package model

type BotBlobInfo struct {

	Mode			string `json:"mode"`
	BotId			string `json:"botId"`
	AccountName		string `json:"account_name"`
	AccountKey		string `json:"account_key"`
	ContainerName	string `json:"container_name"`
	SASToken		string `json:"sas_token"`
}

type BotBlobMapping struct {
	BotMapings []BotBlobInfo `json:"storage_info"`
}
