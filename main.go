package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// --- Structures ---
type AnthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

type InfraRequest struct {
	Action       string `json:"action"`
	InstanceType string `json:"instance_type"`
	Count        int32  `json:"count"`
}

// --- The Core Logic (Reusable) ---
func runInfrastructureAgent(ctx context.Context, cfg aws.Config) error {
	aiOutput := callAI(ctx, cfg)

	var anthropic AnthropicResponse
	if err := json.Unmarshal(aiOutput, &anthropic); err != nil {
		return err
	}
	rawText := anthropic.Content[0].Text
	cleanJSON := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(rawText, "```json", ""), "```", ""))

	var request InfraRequest
	if err := json.Unmarshal([]byte(cleanJSON), &request); err != nil {
		return err
	}

	if request.Action == "create_ec2" {
		return createEC2(ctx, cfg, request)
	}
	return nil
}

// --- Lambda Handler ---
func HandleRequest(ctx context.Context) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", err
	}
	err = runInfrastructureAgent(ctx, cfg)
	return "Task Complete", err
}

// --- Main Entry Point ---
func main() {
	// Check if running in Lambda environment
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		lambda.Start(HandleRequest)
	} else {
		// CLI Execution
		ctx := context.TODO()
		cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
		if err != nil {
			log.Fatal(err)
		}
		
		// Optional: Print Identity for CLI debugging
		stsClient := sts.NewFromConfig(cfg)
		identity, _ := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
		fmt.Printf("Authenticated as: %s\n", *identity.Arn)

		if err := runInfrastructureAgent(ctx, cfg); err != nil {
			log.Fatal(err)
		}
	}
}

// --- Helper Functions (Keep these as they were) ---
func createEC2(ctx context.Context, cfg aws.Config, req InfraRequest) error {
	client := ec2.NewFromConfig(cfg)
	input := &ec2.RunInstancesInput{
		ImageId:      aws.String("ami-053b0d53c279acc90"),
		InstanceType: types.InstanceType(req.InstanceType),
		MinCount:     aws.Int32(req.Count),
		MaxCount:     aws.Int32(req.Count),
	}
	result, err := client.RunInstances(ctx, input)
	if err != nil {
		return err
	}
	fmt.Printf("Success! Instance ID: %s\n", *result.Instances[0].InstanceId)
	return nil
}

func callAI(ctx context.Context, cfg aws.Config) []byte {
	client := bedrockruntime.NewFromConfig(cfg)
	input := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        200,
		"system":            "Return ONLY JSON format: {\"action\": \"create_ec2\", \"instance_type\": \"t2.micro\", \"count\": 1}",
		"messages": []map[string]string{{"role": "user", "content": "I need to launch an ec2 instance"}},
	}
	body, _ := json.Marshal(input)
	output, _ := client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String("anthropic.claude-sonnet-4-6"),
		ContentType: aws.String("application/json"),
		Body:        body,
	})
	return output.Body
}