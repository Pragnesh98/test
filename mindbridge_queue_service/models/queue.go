package models

type QueueApiResponse struct {
	Status       string `json:",omitempty"`
	Msg          string `json:",omitempty"`
	ResponseCode int    `json:",omitempty"`
}

type Body struct {
	ConferenceUUID string `json:"conference_uuid"`
	ConferenceName string `json:"conference_name"`
}

type Entry struct {
	EntryAnnounce  bool   `json:"conference_entry"`
	ConferenceUUID string `json:"conference_uuid"`
	ConferenceName string `json:"conference_name"`
}
