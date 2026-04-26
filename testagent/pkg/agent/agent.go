package agent

import (
	"encoding/json"
	"fmt"

	"github.com/0xMeechie/Aranea/pkg/sdk"
)

const systemPrompt = `You are a capable AI agent with access to tools that let you interact with the real world.
You can read and write files, make HTTP requests, and run shell commands.

Important guidelines:
- Think step by step before acting
- Use the minimum number of tool calls needed to complete the task
- When writing files, always confirm what you wrote
- When running shell commands, prefer safe read-only commands unless explicitly asked to make changes
- Always explain what you are doing and why`

// Agent is the main agent struct
type Agent struct {
	client   *client
	tools    []ToolDefinition
	messages []Message
	ts       *sdk.ToolSet
}

// New creates a new Agent
func New(apiKey string, ts *sdk.ToolSet) *Agent {
	return &Agent{
		client:   newClient(apiKey),
		tools:    toolDefinitions,
		messages: []Message{},
		ts:       ts,
	}
}

// Run takes a user message, runs the agentic loop, and returns the final response
func (a *Agent) Run(userInput string) (string, error) {
	// append the user message to conversation history
	a.messages = append(a.messages, Message{
		Role:    "user",
		Content: userInput,
	})

	// agentic loop — keep going until Claude stops calling tools
	for {
		resp, err := a.client.send(a.messages, a.tools)
		if err != nil {
			return "", fmt.Errorf("API call failed: %w", err)
		}

		// append Claude's response to history
		a.messages = append(a.messages, Message{
			Role:    "assistant",
			Content: resp.Content,
		})

		// if Claude is done (no more tool calls), return the text response
		if resp.StopReason == "end_turn" {
			return extractText(resp.Content), nil
		}

		// if Claude wants to use tools, execute them and feed results back
		if resp.StopReason == "tool_use" {
			toolResults, err := a.executeTools(resp.Content)
			if err != nil {
				return "", fmt.Errorf("tool execution failed: %w", err)
			}

			// append tool results as a user message (Anthropic's required format)
			a.messages = append(a.messages, Message{
				Role:    "user",
				Content: toolResults,
			})

			// loop again — Claude will process the results and either call more tools or respond
			continue
		}

		// unexpected stop reason
		return extractText(resp.Content), nil
	}
}

// executeTools runs all tool calls in a response and returns the results
func (a *Agent) executeTools(blocks []ContentBlock) ([]ToolResultContent, error) {
	var results []ToolResultContent

	for _, block := range blocks {
		if block.Type != "tool_use" {
			continue
		}

		result := dispatchTool(a.ts, block.Name, block.Input)

		results = append(results, ToolResultContent{
			Type:      "tool_result",
			ToolUseID: block.ID,
			Content:   result,
		})
	}

	return results, nil
}

// extractText pulls the text content from a response
func extractText(blocks []ContentBlock) string {
	for _, block := range blocks {
		if block.Type == "text" {
			return block.Text
		}
	}
	return ""
}

// Reset clears conversation history (start a fresh session)
func (a *Agent) Reset() {
	a.messages = []Message{}
}

// needed for marshaling tool results
var _ = json.Marshal
