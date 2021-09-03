package contracts

type HealthResponse struct {
	QueueLength int              `json:"queue_length"`
	PipeHealth  []*GetPipeHealth `json:"pipehealth"`
}
type GetPipeHealth struct {
	Name   string     `json:"name"`
	Health SpanHealth `json:"health"`
}

// SpanHealth contains the health of a particular span
type SpanHealth struct {
	Up           bool `json:"up"`
	ChannelCount int  `json:"count"`
}

type GetPipeHealthResponse struct {
	BaseResponse
	ResponseData SingleGetPipeHealthResponse `json:"response"`
}

type SingleGetPipeHealthResponse struct {
	SingleResponse
	ResourceData *HealthResponse `json:"data,omitempty"`
}
