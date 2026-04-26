package main

import (
	"ai-infra-agent/pkg/ai/vertex"
	"ai-infra-agent/pkg/models"
	"ai-infra-agent/pkg/provider"
	"ai-infra-agent/pkg/provider/gcp"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	projectID := flag.String("project", os.Getenv("PROJECT_ID"), "GCP Project ID")
	region := flag.String("region", "asia-south1", "GCP Region")
	zone := flag.String("zone", "asia-south1-a", "GCP Zone")
	flag.Parse()

	if *projectID == "" {
		log.Fatal("PROJECT_ID is required")
	}

	ctx := context.Background()

	gcpProvider, err := gcp.NewGCPProvider(ctx, *projectID, *region, *zone)
	if err != nil {
		log.Fatalf("Failed to initialize GCP Provider: %v", err)
	}

	aiClient, err := vertex.NewVertexAIClient(ctx, *projectID)
	if err != nil {
		log.Fatalf("Failed to initialize Vertex AI Client: %v", err)
	}

	if err := runAgent(ctx, gcpProvider, aiClient); err != nil {
		log.Fatalf("Agent failed: %v", err)
	}
}

func runAgent(ctx context.Context, gcpProvider *gcp.GCPProvider, aiClient *vertex.VertexAIClient) error {

	aiOutput, err := aiClient.InvokeModel(ctx, "I need to launch a compute instance")
	if err != nil {
		return fmt.Errorf("AI call failed: %w", err)
	}

	// safer cleanup (works even without ```json)
	cleanJSON := strings.Trim(aiOutput, "` \n")

	var req models.InfraRequest
	if err := json.Unmarshal([]byte(cleanJSON), &req); err != nil {
		return fmt.Errorf("failed to parse AI response: %w | raw=%s", err, aiOutput)
	}

	switch req.Action {

	case "create_gcp_instance":

		name := fmt.Sprintf("gcp-instance-%d", time.Now().Unix())

		resp, err := gcpProvider.CreateInstance(ctx, &provider.InstanceRequest{
			Name:        name,
			MachineType: req.InstanceType,
			Count:       req.Count,
		})
		if err != nil {
			return fmt.Errorf("failed to create instance: %w", err)
		}

		fmt.Printf("Instance created: %s\n", resp.ID)

	default:
		return fmt.Errorf("unsupported action: %s", req.Action)
	}

	return nil
}
