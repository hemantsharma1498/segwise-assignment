package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hemantsharma1498/segwise-assignment/pkg/scraper"
)

type OpenAIReq struct {
	Model    string       `json:"model"`
	Messages []OpenAIRole `json:"messages"`
}

type OpenAIRole struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func GetMessage(userData scraper.Profile, apiKey string) (string, error) {
	apiURL := "https://api.openai.com/v1/chat/completions"
	jsonProfile, err := json.Marshal(userData)
	if err != nil {
		return "", err
	}

	systemMessage := OpenAIRole{
		Role: "system",
		Content: "You will be provided with a JSON containing slices and strings of posts, experience, education, about, name, and geography for a LinkedIn user. " +
			"Create a connect message of maximum two lines. Prioritize the content of the message by posts, experience, education, about, name, and geography. " +
			"If nothing is present, send a sample connect message.",
	}
	userMessage := OpenAIRole{
		Role:    "user",
		Content: string(jsonProfile),
	}

	reqBody := OpenAIReq{
		Model:    "gpt-4o-mini",
		Messages: []OpenAIRole{systemMessage, userMessage},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return "", err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return "", err
	}
	defer resp.Body.Close()

	var message string
	if resp.StatusCode == http.StatusOK {
		response := &OpenAIResponse{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			fmt.Println("Error decoding response:", err)
			return "", err
		}
		message = response.Choices[0].Message.Content
	} else {
		fmt.Printf("Request failed with status code: %d\n", resp.StatusCode)
	}
	return message, nil
}
