/*
	Package openai provides functionality to interact with OpenAI's GPT API

specifically for generating LinkedIn connection messages based on profile data.
The package handles the communication with OpenAI's API and processes profile
information to create personalized connection requests.

Basic usage:

	profile := scraper.Profile{
	    Name: "John Doe",
	    Experience: []scraper.Experience{...},
	    Posts: []scraper.Post{...},
	}

	message, err := openai.GetMessage(profile, "your-api-key")
	if err != nil {
	    log.Fatal(err)
	}
	fmt.Println(message)
*/
package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hemantsharma1498/segwise-assignment/pkg/scraper"
)

/*
	OpenAIReq represents the request structure for OpenAI's chat completion API.

It includes the model to be used and an array of messages with roles and content.
*/
type OpenAIReq struct {
	Model    string       `json:"model"`    // The GPT model to be used
	Messages []OpenAIRole `json:"messages"` // Array of messages with roles
}

/*
	OpenAIRole represents a single message in the conversation with OpenAI's API.

Each message has a role (system, user, or assistant) and content.
*/
type OpenAIRole struct {
	Role    string `json:"role"`    // Role of the message sender (system/user/assistant)
	Content string `json:"content"` // Content of the message
}

/*
	OpenAIResponse represents the response structure from OpenAI's API.

It contains an array of choices, each containing a message.
*/
type OpenAIResponse struct {
	Choices []Choice `json:"choices"` // Array of possible responses
}

/*
	Choice represents a single response option from OpenAI's API.

It contains the message generated by the model.
*/
type Choice struct {
	Message Message `json:"message"` // The generated message
}

/*
	Message represents the content of a response from OpenAI's API.

It includes the role of the responder and the message content.
*/
type Message struct {
	Role    string `json:"role"`    // Role of the message sender
	Content string `json:"content"` // Content of the generated message
}

/*
	GetMessage generates a personalized LinkedIn connection message based on a user's profile data.

It processes the profile information and uses OpenAI's GPT model to create a contextual
connection request. The function prioritizes different aspects of the profile in the following order:
posts, experience, education, about section, name, and geography.

Parameters:
  - userData: A scraper.Profile struct containing the LinkedIn profile information
  - apiKey: OpenAI API key for authentication

Returns:
  - string: The generated connection message
  - error: Any error encountered during the API request or response processing

Example:

	profile := scraper.Profile{
	    Name: "John Doe",
	    Experience: []scraper.Experience{
	        {Company: "Tech Corp", Title: "Software Engineer"},
	    },
	}
	message, err := GetMessage(profile, "your-api-key")
*/
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
