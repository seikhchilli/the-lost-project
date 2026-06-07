package clients

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"
	"titles-mcp/config"

	"google.golang.org/genai"
)

var llmClient LLM

func TestMain(m *testing.M) {
	config.LoadConfig()
	ctx := context.Background()
	llmClient = NewLLM(ctx)

	m.Run()
}

func Test_GenerateContent(t *testing.T) {
	ctx := context.Background()
	prompt := "Suggest 5 popular unique movie name and its release year of genre comedy. Return in json format array [{movie_name: nameOfTheMovie, release_year: ReleaseYear}]. There should be no text before opening `[` and after closing `]`. Do not include these movies: [`Interstellar`, `Weapons`, `Hangover`, `Superbad`]"
	result, err := llmClient.GenerateContent(ctx, prompt, nil)
	if err != nil {
		t.Errorf("generate content test failed with error: %v", err)
	}
	if result == "" {
		t.Error("Expected result, found empty string")
	}
	log.Print("result: ", result)
}

func Test_GenerateContentWithConfig(t *testing.T) {
	t.Skip()
	ctx := context.Background()
	responseSchema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"movie_name": {
				Type: genai.TypeString,
			},
			"release_year": {
				Type: genai.TypeString,
			},
		},
		Required: []string{"movie_name", "release_year"},
	}
	prompt := fmt.Sprintf("Suggest a completely random comedy movie with imdb rating greater than 7 and its release year. Use this random seed to ensure a unique choice: %v", time.Now().UnixNano())
	temperature := float32(2)
	topP := float32(0.95) // 0.0 to 1.0. Higher = wider variety of words considered
	topK := float32(64)
	promptConfig := genai.GenerateContentConfig{Temperature: &temperature, ResponseMIMEType: "application/json", ResponseSchema: responseSchema, TopP: &topP, TopK: &topK}
	result, err := llmClient.GenerateContent(ctx, prompt, &promptConfig)
	if err != nil {
		t.Errorf("generate content test failed with error: %v", err)
	}
	if result == "" {
		t.Error("Expected result, found empty string")
	}
	log.Print("result: ", result)

}

func Test_GenerateContentWithConfigInLoop(t *testing.T) {
	t.Skip()
	ctx := context.Background()
	responseSchema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"movies": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"movie_name": {
							Type: genai.TypeString,
						},
						"release_year": {
							Type: genai.TypeString,
						},
					},
					Required: []string{"movie_name", "release_year"},
				},
			},
		},
		Required: []string{"movies"},
	}
	temperature := float32(1)

	topP := float32(0.8) // 0.0 to 1.0. Higher = wider variety of words considered
	topK := float32(64)

	prompt := "Suggest 5 completely random, obscure comedy movie with imdb rating greater than 7 and its release year."
	promptConfig := genai.GenerateContentConfig{Temperature: &temperature, ResponseMIMEType: "application/json", ResponseSchema: responseSchema, TopP: &topP, TopK: &topK}
	result, err := llmClient.GenerateContent(ctx, prompt, &promptConfig)
	if err != nil {
		t.Errorf("generate content test failed with error: %v", err)
	}
	if result == "" {
		t.Error("Expected result, found empty string")
	}
	log.Print("result: ", result)

}

func Test_ListAllModels(t *testing.T) {
	t.Skip()
	llmClient.ListAllModels(context.Background())
}
