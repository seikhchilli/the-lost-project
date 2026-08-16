package clients

import (
	"context"
	"log"
	"os"
	"titles-mcp/config"

	"google.golang.org/api/iterator"
	"google.golang.org/genai"
)

type LLM interface {
	GenerateContent(ctx context.Context, prompt string, promptConfig *genai.GenerateContentConfig) (string, error)
	ListAllModels(ctx context.Context)
}

type gemini struct {
	client genai.Client
}

func NewLLM(ctx context.Context) LLM {
	apiKey := config.AppConfig.LLMConfig.GeminiAPIKey
	if apiKey == "" {
		log.Fatal("LLM API key not found.")
	}
	geminiClient, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		log.Fatalf("Failed to create gemini client: %v", err)
	}

	return &gemini{
		client: *geminiClient,
	}
}

func (g *gemini) ListAllModels(ctx context.Context) {
	iter := g.client.Models.All(ctx)

	log.Println("Available Models:")
	for modelInfo, err := range iter {
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to list models: %v", err)
		}
		log.Printf("- %s (Description: %s)\n", modelInfo.Name, modelInfo.Description)
	}
}

func (g *gemini) GenerateContent(ctx context.Context, prompt string, promptConfig *genai.GenerateContentConfig) (string, error) {
	log.Print("Generating content using llm")
	file, err := os.OpenFile("llm-logs/prompt.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open prompt log file with err: %v", err)
	} else {
		file.Write([]byte(prompt))
		file.Close()
	}
	result, err := g.client.Models.GenerateContent(
		ctx, "gemini-flash-lite-latest", genai.Text(prompt), promptConfig,
	)
	if err != nil {
		log.Printf("Failed to generate content: %v", err)
		return "", err
	}
	log.Print("Generated conetnt using llm")
	return result.Text(), nil
}
