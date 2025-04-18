package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"christville/model"

	"github.com/gin-gonic/gin"
)

// GetVerse handles the request to fetch Bible verses based on a user prompt.
func GetVerse(c *gin.Context) {
	var request model.VerseRequest

	// Validate input
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Call DeepSeek AI to process the prompt and return verses
	verses, err := getTopicOrVerseFromDeepSeek(request.Prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process prompt with AI: " + err.Error()})
		return
	}

	// Return successful response
	c.JSON(http.StatusOK, gin.H{
		"prompt": request.Prompt,
		"result": verses,
	})
}

// getTopicOrVerseFromDeepSeek queries DeepSeek AI to suggest Bible verses based on mood.
func getTopicOrVerseFromDeepSeek(prompt string) (string, error) {
	url := "https://api.deepseek.com/chat/completions"

	// Get API key from environment
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		// Log all environment variables for debugging (excluding sensitive ones)
		fmt.Printf("Environment variables available: %v\n", os.Environ())
		return "", fmt.Errorf("DEEPSEEK_API_KEY not found in environment variables. Please check your Docker environment configuration.")
	}

	// Create the JSON payload for DeepSeek
	payload := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a compassionate Christian assistant. Analyze the user's mood and respond with a brief, empathetic message followed by 2-3 relevant Bible verses. IMPORTANT: Return ONLY a JSON object with this exact structure, no markdown or additional text: {\"message\": \"brief empathetic response\", \"verses\": [{\"text\": \"verse text\", \"reference\": \"book chapter:verse\"}, ...]}",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
		"max_tokens":  300,
		"stream":      false,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to create JSON payload: %w", err)
	}

	// Create the HTTP request
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("DeepSeek API error: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("failed to parse response JSON: %w", err)
	}

	// Extract the assistant's reply (Bible verses)
	if len(result.Choices) == 0 {
		return "", errors.New("no choices in DeepSeek response")
	}

	// Return the raw JSON response
	return result.Choices[0].Message.Content, nil
}
