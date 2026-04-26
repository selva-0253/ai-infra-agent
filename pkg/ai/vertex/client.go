package vertex

import (
	"context"
	"encoding/json"
	"fmt"

	"ai-infra-agent/pkg/models"

	"github.com/google/generative-ai-go/genai"
)

type VertexAIClient struct {
	client    *genai.Client
	projectID string
}

func NewVertexAIClient(ctx context.Context, projectID string) (*VertexAIClient, error) {
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  projectID,
		Location: "asia-south1",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI client: %w", err)
	}

	return &VertexAIClient{
		client:    c,
		projectID: projectID,
	}, nil
}

func (v *VertexAIClient) InvokeModel(ctx context.Context, prompt string) (string, error) {
	model := v.client.GenerativeModel("gemini-1.5-pro")

	// Correct system instruction (valid JSON only)
	systemPrompt := `You are an infrastructure automation agent.
Respond ONLY in valid JSON:
{"action":"create_gcp_instance","instance_type":"e2-micro","count":1}`

	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{{Text: systemPrompt}},
	}

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to invoke model: %w", err)
	}

	if len(resp.Candidates) == 0 ||
		len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty model response")
	}

	text := resp.Candidates[0].Content.Parts[0].Text

	// Optional: validate JSON format by unmarshalling into struct
	var result models.InfraRequest
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal model response: %w", err)
	}

	return text, nil
}
