// backend/main.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// This struct matches the request body we expect from the Chrome extension
type SummarizeRequest struct {
	Text string `json:"text"`
}

// This struct matches the response body we'll send back to the extension
type SummarizeResponse struct {
	Summary string `json:"summary"`
}

func main() {
	// Get the API key from an environment variable for security
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable not set")
	}

	http.HandleFunc("/summarize", summarizeHandler(apiKey))

	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func summarizeHandler(apiKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers to allow requests from the extension
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight OPTIONS request for CORS
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		// Decode the request body from the extension
		var reqPayload SummarizeRequest
		if err := json.NewDecoder(r.Body).Decode(&reqPayload); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if reqPayload.Text == "" {
			http.Error(w, "Text field is required", http.StatusBadRequest)
			return
		}

		// Call the Gemini API
		summary, err := callGeminiAPI(apiKey, reqPayload.Text)
		if err != nil {
			log.Printf("Error calling Gemini API: %v", err)
			http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
			return
		}

		// Prepare and send the response
		respPayload := SummarizeResponse{Summary: summary}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(respPayload)
	}
}

// callGeminiAPI handles the actual communication with Google's API
func callGeminiAPI(apiKey, textToSummarize string) (string, error) {
	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + apiKey

	// Construct the request payload for Gemini
	prompt := fmt.Sprintf("Summarize the following text in 200 words:\n\n%s", textToSummarize)
	requestBody, err := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": prompt,
					},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Make the POST request to Gemini
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to make request to Gemini API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the Gemini response
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", fmt.Errorf("failed to decode Gemini response: %w", err)
	}

	// Extract the summary text
	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no summary found in Gemini response")
}
