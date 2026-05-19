package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestE2E(t *testing.T) {
	ctx := context.Background()

	// 1. Build the server
	t.Log("Building server...")
	buildCmd := exec.Command("go", "build", "-o", "../titles-mcp.exe", "../main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatal("Failed to build server: ", err)
	}

	// 2. Create a new client
	client := mcp.NewClient(&mcp.Implementation{Name: "mcp-client", Version: "v1.0.0"}, nil)

	// 3. Connect to a server over stdin/stdout with USE_SQLITE=true
	cmd := exec.CommandContext(ctx, "./../titles-mcp.exe")
	cmd.Env = append(os.Environ(), "USE_SQLITE=true")
	transport := &mcp.CommandTransport{Command: cmd}
	
	t.Log("Connecting to server...")
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatal("Session creation failed: ", err)
	}
	defer session.Close()

	// Scenarios
	runAddTitles(t, ctx, session)
	summaries := runListAll(t, ctx, session)
	if len(summaries) > 0 {
		id := uint(summaries[0]["id"].(float64))
		runMarkAsWatched(t, ctx, session, id)
		runSearch(t, ctx, session, "Anorrrra")
		runRemoveFromWatched(t, ctx, session, id)
		runMarkAsWished(t, ctx, session, id)
		runRemoveFromWished(t, ctx, session, id)
	}

	t.Log("E2E Tests completed successfully")
}

func runAddTitles(t *testing.T, ctx context.Context, session *mcp.ClientSession) {
	t.Log("Scenario: Add Titles")
	params := &mcp.CallToolParams{
		Name: "add_titles",
		Arguments: map[string]any{
			"titles": []map[string]any{
				{
					"name":         "Anorrrra",
					"release_year": 2024,
					"genres":        []string{"Drama", "Comedy", "Romance"},
					"imdb_rating":   7.1,
					"imdb_id":       "tt2860845951",
					"tmdb_id":       "10646613",
				},
			},
		},
	}
	res, err := session.CallTool(ctx, params)
	if err != nil {
		t.Fatalf("AddTitles failed: %v", err)
	}
	checkResult(t, res)
}

func runListAll(t *testing.T, ctx context.Context, session *mcp.ClientSession) []map[string]any {
	t.Log("Scenario: List All Titles")
	params := &mcp.CallToolParams{
		Name:      "list_all_titles",
		Arguments: map[string]any{},
	}
	res, err := session.CallTool(ctx, params)
	if err != nil {
		t.Fatalf("ListAllTitles failed: %v", err)
	}
	
	// Parse response to get an ID for subsequent tests
	var output struct {
		Titles []map[string]any `json:"titles"`
	}
	for _, c := range res.Content {
		if textContent, ok := c.(*mcp.TextContent); ok {
			json.Unmarshal([]byte(textContent.Text), &output)
		}
	}
	checkResult(t, res)
	return output.Titles
}

func runMarkAsWatched(t *testing.T, ctx context.Context, session *mcp.ClientSession, id uint) {
	t.Log(fmt.Sprintf("Scenario: Mark as Watched (ID: %d)", id))
	params := &mcp.CallToolParams{
		Name: "mark_title_as_watched",
		Arguments: map[string]any{
			"title_id": id,
		},
	}
	res, err := session.CallTool(ctx, params)
	if err != nil {
		t.Fatalf("MarkTitleAsWatched failed: %v", err)
	}
	checkResult(t, res)
}

func runSearch(t *testing.T, ctx context.Context, session *mcp.ClientSession, name string) {
	t.Log(fmt.Sprintf("Scenario: Search Titles (Name: %s)", name))
	params := &mcp.CallToolParams{
		Name: "search_titles",
		Arguments: map[string]any{
			"title_names": []string{name},
		},
	}
	res, err := session.CallTool(ctx, params)
	if err != nil {
		t.Fatalf("SearchTitles failed: %v", err)
	}
	checkResult(t, res)
}

func runRemoveFromWatched(t *testing.T, ctx context.Context, session *mcp.ClientSession, id uint) {
	t.Log(fmt.Sprintf("Scenario: Remove from Watched (ID: %d)", id))
	params := &mcp.CallToolParams{
		Name: "remove_from_watched",
		Arguments: map[string]any{
			"title_id": id,
		},
	}
	res, err := session.CallTool(ctx, params)
	if err != nil {
		t.Fatalf("RemoveFromWatched failed: %v", err)
	}
	checkResult(t, res)
}

func runMarkAsWished(t *testing.T, ctx context.Context, session *mcp.ClientSession, id uint) {
	t.Log(fmt.Sprintf("Scenario: Mark as Wished (ID: %d)", id))
	params := &mcp.CallToolParams{
		Name: "mark_title_as_wished",
		Arguments: map[string]any{
			"title_id": id,
		},
	}
	res, err := session.CallTool(ctx, params)
	if err != nil {
		t.Fatalf("MarkTitleAsWished failed: %v", err)
	}
	checkResult(t, res)
}

func runRemoveFromWished(t *testing.T, ctx context.Context, session *mcp.ClientSession, id uint) {
	t.Log(fmt.Sprintf("Scenario: Remove from Wished (ID: %d)", id))
	params := &mcp.CallToolParams{
		Name: "remove_from_wished",
		Arguments: map[string]any{
			"title_id": id,
		},
	}
	res, err := session.CallTool(ctx, params)
	if err != nil {
		t.Fatalf("RemoveFromWished failed: %v", err)
	}
	checkResult(t, res)
}

func checkResult(t *testing.T, res *mcp.CallToolResult) {
	for _, c := range res.Content {
		if textContent, ok := c.(*mcp.TextContent); ok {
			log.Print("Result: ", textContent.Text)
			var status struct {
				Status string `json:"status"`
			}
			json.Unmarshal([]byte(textContent.Text), &status)
			if status.Status == "error" {
				t.Errorf("Tool returned error status: %s", textContent.Text)
			}
		}
	}
}
