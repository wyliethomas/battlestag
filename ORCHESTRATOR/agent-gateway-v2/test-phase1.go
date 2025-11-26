package main

import (
	"agent-gateway/gateway"
	"agent-gateway/llm"
	"fmt"
	"log"
	"time"
)

func main() {
	fmt.Println("=== Agent Gateway v2 - Phase 1 Test ===\n")

	// Initialize registry
	fmt.Println("Loading programs from ./programs directory...")
	registry, err := gateway.NewRegistry("./programs")
	if err != nil {
		log.Fatalf("Failed to create registry: %v", err)
	}
	fmt.Printf("✓ Loaded %d programs\n", registry.Count())

	// List loaded programs
	fmt.Println("\nRegistered programs:")
	for name, program := range registry.GetAllPrograms() {
		fmt.Printf("  - %s: %s\n", name, program.Manifest.Description)
		fmt.Printf("    Executable: %s\n", program.ExecutablePath)
	}

	// Initialize executor
	fmt.Println("\n✓ Executor initialized (30s timeout)")
	executor := gateway.NewExecutor(30 * time.Second)

	// Initialize Ollama client
	ollamaURL := "http://localhost:11434"
	ollamaModel := "llama3.2:latest"
	fmt.Printf("\n✓ Ollama client initialized (%s)\n", ollamaModel)
	llmClient := llm.NewOllamaClient(ollamaURL, ollamaModel)

	// Initialize gateway
	fmt.Println("\n✓ Gateway initialized")
	gw := gateway.NewGateway(registry, executor, llmClient)

	// Test: Direct tool execution
	fmt.Println("\n=== Testing Direct Tool Execution ===")
	echoProgram, exists := registry.GetProgram("echo")
	if !exists {
		log.Fatal("Echo program not found")
	}

	echoInput := []byte(`{"message": "Hello from the gateway!"}`)
	fmt.Printf("Executing: echo %s\n", string(echoInput))
	result := executor.Execute(echoProgram, echoInput)
	if result.Success {
		fmt.Printf("✓ Result: %s", result.Output)
	} else {
		fmt.Printf("✗ Error: %s\n", result.Error)
	}

	// Test: Time tool
	timeProgram, exists := registry.GetProgram("time")
	if !exists {
		log.Fatal("Time program not found")
	}

	fmt.Printf("\nExecuting: time\n")
	result = executor.Execute(timeProgram, []byte(`{}`))
	if result.Success {
		fmt.Printf("✓ Result: %s", result.Output)
	} else {
		fmt.Printf("✗ Error: %s\n", result.Error)
	}

	// Test: Ollama tools format conversion
	fmt.Println("\n=== Testing Ollama Tools Format ===")
	tools := registry.ToOllamaTools()
	fmt.Printf("Converted %d programs to Ollama tools:\n", len(tools))
	for _, tool := range tools {
		fmt.Printf("  - %s (%s)\n", tool.Function.Name, tool.Type)
	}

	fmt.Println("\n=== Phase 1 Tests Complete ===")
	fmt.Println("\nGateway is ready. To test full LLM integration:")
	fmt.Println("1. Ensure Ollama is running: ollama serve")
	fmt.Println("2. Ensure model is available: ollama pull llama3.2:latest")
	fmt.Println("3. Use the gateway.ProcessChat() method with test messages")

	// Suppress unused variable warning
	_ = gw
}
