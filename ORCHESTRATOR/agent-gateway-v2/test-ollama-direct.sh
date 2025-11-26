#!/bin/bash

# Direct test of Ollama function calling
curl -s http://192.168.1.232:11434/api/chat -d '{
  "model": "llama3.1:8b",
  "messages": [
    {"role": "user", "content": "What time is it?"}
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "time",
        "description": "Returns the current date and time",
        "parameters": {
          "type": "object",
          "properties": {},
          "required": []
        }
      }
    }
  ],
  "stream": false
}' | jq .
