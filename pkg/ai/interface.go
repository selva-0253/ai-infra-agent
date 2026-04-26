package provider

import "context"

type AIProvider interface {
	InvokeModel(ctx context.Context, prompt string) (string, error)
}

type AIRequest struct {
	UserPrompt string
}

type AIResponse struct {
	Action       string `json:"action"`
	InstanceType string `json:"instance_type"`
	Count        int32  `json:"count"`
}
