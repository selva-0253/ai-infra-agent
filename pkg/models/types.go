package models

//Extract from current main.go

type InfraRequest struct {
	Action       string `json:"action"`
	InstanceType string `json:"instance_type"`
	Count        int32  `json:"count"`
}

type VertexResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}
