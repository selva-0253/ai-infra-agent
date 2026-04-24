package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// 1. Structures to handle the Anthropic API response
type AnthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

// 2. Your actual Infrastructure request
type InfraRequest struct {
	Action       string `json:"action"`
	InstanceType string `json:"instance_type"`
	Count        int32  `json:"count"`
}

func main() {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatal(err)
	}
	stsClient := sts.NewFromConfig(cfg)
    identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
    if err != nil {
        log.Fatal("Could not verify identity:", err)
    }
    fmt.Printf("Authenticated as: %s\n", *identity.Arn)
	// --- STEP 1: Call AI ---
	aiOutput := callAI(ctx, cfg)

	// --- STEP 2: Parse Wrapper ---
	var anthropic AnthropicResponse
	if err := json.Unmarshal(aiOutput, &anthropic); err != nil {
		log.Fatal(err)
	}
	rawText := anthropic.Content[0].Text

	// --- STEP 3: Clean Markdown ---
	cleanJSON := strings.ReplaceAll(rawText, "```json", "")
	cleanJSON = strings.ReplaceAll(cleanJSON, "```", "")
	cleanJSON = strings.TrimSpace(cleanJSON)

	// --- STEP 4: Parse Logic ---
	var request InfraRequest
	json.Unmarshal([]byte(cleanJSON), &request)

	fmt.Printf("Parsed Action: %s, Type: %s\n", request.Action, request.InstanceType)

	// --- STEP 5: Execute Infrastructure ---
	if request.Action == "create_ec2" {
		createEC2(ctx, cfg, request)
	}
}

func createEC2(ctx context.Context, cfg aws.Config, req InfraRequest) {
	client := ec2.NewFromConfig(cfg)

	// NOTE: Update this AMI ID for your region!
	// This is Amazon Linux 2023 AMI for us-east-1
	input := &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-053b0d53c279acc90"), 
		InstanceType: types.InstanceType(req.InstanceType),
		MinCount:     aws.Int32(req.Count),
		MaxCount:     aws.Int32(req.Count),
	}

	result, err := client.RunInstances(ctx, input)
	if err != nil {
		log.Fatalf("AWS API Error: %v", err)
	}
	fmt.Printf("Success! Instance ID: %s\n", *result.Instances[0].InstanceId)
}

func callAI(ctx context.Context, cfg aws.Config) []byte {
	client := bedrockruntime.NewFromConfig(cfg)
	input := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        200,
		"system":            "Return ONLY JSON format: {\"action\": \"create_ec2\", \"instance_type\": \"t2.micro\", \"count\": 1}",
		"messages": []map[string]string{
			{"role": "user", "content": "I need to launch an ec2 instance"},
		},
	}
	body, _ := json.Marshal(input)

	output, _ := client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String("anthropic.claude-sonnet-4-6"), // Update if needed
		ContentType: aws.String("application/json"),
		Body:        body,
	})
	return output.Body
}