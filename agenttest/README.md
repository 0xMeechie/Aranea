# AgentGuard — Go AI Agent

A minimal AI agent built in Go using the Anthropic API directly — no frameworks, no magic.
The agent has three real tools: filesystem access, HTTP requests, and shell commands.

## Project structure

```
.
├── main.go              # Entry point — REPL loop
├── go.mod
└── agent/
    ├── agent.go         # Core agentic loop
    ├── client.go        # Anthropic API client
    ├── tools.go         # Tool implementations
    └── types.go         # Request/response types
```

## Setup

### 1. Get an Anthropic API key

Sign up at https://console.anthropic.com and create an API key.

### 2. Set your API key

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

### 3. Run the agent

```bash
go run main.go
```

## Example prompts to try

Once the agent is running, try these to see the tools in action:

**Filesystem:**
```
Write a file called hello.txt with the content "Hello from the agent"
```
```
Read the file hello.txt
```

**Shell:**
```
What is the current working directory and what files are in it?
```
```
How much disk space is available on this machine?
```

**HTTP:**
```
Fetch https://httpbin.org/get and tell me what it returns
```

**Multi-step:**
```
Check the current date using a shell command, then write a file called today.txt containing today's date
```

## How the agentic loop works

This is the core pattern you need to understand:

1. User sends a message
2. Claude responds — either with text (done) or with tool_use blocks (needs to act)
3. If tool_use: execute the tools, send results back to Claude as a user message
4. Claude processes the results and responds again
5. Repeat until Claude sends stop_reason: "end_turn"

That loop in agent/agent.go is the heart of every AI agent, regardless of framework.

## What to notice (AgentGuard context)

As you use the agent, pay attention to:

- The agent has NO restrictions — it will read any file, run any command, call any URL
- There is NO audit log — you have no record of what it did
- There is NO rate limiting — it will make as many API calls as it wants
- There is NO identity — you don't know which agent instance did what

These are exactly the gaps AgentGuard is designed to fill.
