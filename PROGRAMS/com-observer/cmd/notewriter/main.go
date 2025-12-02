package main

import (
	"flag"
	"log"

	"github.com/battlestag/com-observer/internal/config"
	"github.com/battlestag/com-observer/internal/extractor"
	"github.com/battlestag/com-observer/internal/messages"
	"github.com/battlestag/com-observer/internal/writers"
)

const (
	batchSize = 20 // Process messages in batches of 20
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize message queue
	queue, err := messages.NewQueue(cfg.Queue.Path)
	if err != nil {
		log.Fatalf("Failed to create message queue: %v", err)
	}

	// Check queue size
	queueCount, err := queue.Count()
	if err != nil {
		log.Fatalf("Failed to get queue count: %v", err)
	}

	if queueCount == 0 {
		log.Println("No messages in queue")
		return
	}

	log.Printf("Processing %d messages from queue", queueCount)

	// Initialize task extractor
	taskExtractor := extractor.NewExtractor(&cfg.LLM)

	// Initialize note writer
	writer, err := writers.NewWriter(cfg)
	if err != nil {
		log.Fatalf("Failed to create writer: %v", err)
	}

	// Process queue in batches
	totalTasks := 0
	totalMessages := 0

	for {
		// Pop a batch of messages
		msgs, err := queue.Pop(batchSize)
		if err != nil {
			log.Fatalf("Failed to pop messages from queue: %v", err)
		}

		if len(msgs) == 0 {
			break // Queue is empty
		}

		totalMessages += len(msgs)
		log.Printf("Processing batch of %d messages", len(msgs))

		// Extract tasks using LLM
		tasks, err := taskExtractor.Extract(cfg.User.Name, msgs)
		if err != nil {
			log.Printf("Error extracting tasks: %v", err)
			continue
		}

		log.Printf("Extracted %d tasks from batch", len(tasks))

		// Write each task
		for _, task := range tasks {
			// Check if task already exists (by permalink)
			if writer.TaskExists(task.Permalink) {
				log.Printf("Task already exists (skipping): %s", task.Description)
				continue
			}

			// Write task
			if err := writer.WriteTask(task); err != nil {
				log.Printf("Error writing task: %v", err)
				continue
			}

			totalTasks++
			log.Printf("Wrote task: %s", task.Description)
		}
	}

	log.Printf("Processing complete: %d messages processed, %d tasks written", totalMessages, totalTasks)
}
