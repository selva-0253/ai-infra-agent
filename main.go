package main

import(
	"log"
	"fmt"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)


func main(){
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err!= nil {
		log.Fatal(err)
	}

client := bedrockruntime.NewFromConfig(cfg)


systemPrompt := `You are an AWS Infrastructure Agent. 
    The user will ask for resources. You MUST return ONLY valid JSON.
    Format: {"action": "create_ec2", "instance_type": "t2.micro", "count": 1}`

userRequest := `I need to launch an ec2 instance`

input := map[string]interface{}{
        "anthropic_version": "bedrock-2023-05-31",
	"max_tokens": 200,
	"system": systemPrompt,
	"messages": []map[string]string{
		{
			"role":    "user",
			"content": userRequest,
		},
	},
}

body, _ := json.Marshal(input)

output, err := client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
	ModelId: aws.String("global.anthropic.claude-sonnet-4-6"),
	ContentType: aws.String("application/json"),
	Body:        body,
})
if err!= nil {
	log.Fatal(err)
}

fmt.Printf("AI response: %s\n", string(output.Body))

}







