package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Facebook webhook event structs
type FBWebhookEvent struct {
	Object string `json:"object"`
	Entry  []struct {
		Messaging []struct {
			Sender struct {
				ID string `json:"id"`
			} `json:"sender"`
			Message struct {
				Text string `json:"text"`
			} `json:"message"`
		} `json:"messaging"`
	} `json:"entry"`
}

type FBMessage struct {
	Recipient struct {
		ID string `json:"id"`
	} `json:"recipient"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
}

// GET /webhook - Facebook webhook verification
func FBWebhookVerify(c *gin.Context) {
	token := c.Query("hub.verify_token")
	fbVerifyToken := os.Getenv("FB_VERIFY_TOKEN")

	if token == "" || fbVerifyToken == "" {
		c.String(http.StatusBadRequest, "Missing verify token")
		return
	}

	if token == fbVerifyToken {
		challenge := c.Query("hub.challenge")
		c.String(http.StatusOK, challenge)
		return
	}

	c.String(http.StatusForbidden, "Verification failed")
}

// POST /webhook - Receive messages from Facebook Messenger
func FBWebhookReceive(c *gin.Context) {
	var event FBWebhookEvent
	if err := c.BindJSON(&event); err != nil {
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}

	for _, entry := range event.Entry {
		for _, msg := range entry.Messaging {
			senderID := msg.Sender.ID
			text := msg.Message.Text
			log.Printf("Received message from %s: %s\n", senderID, text)

			reply, err := askGemini(text)
			if err != nil {
				log.Println("Gemini error:", err)
				reply = "Sorry, I couldn't process your question right now."
			}

			if err := sendFBMessage(senderID, reply); err != nil {
				log.Println("Failed to send message:", err)
			}
		}
	}

	c.Status(http.StatusOK)
}

// askGemini sends the user question to Gemini API and returns AI-generated reply
func askGemini(question string) (string, error) {
	geminiApiKey := os.Getenv("GEMINI_API_KEY")
	if geminiApiKey == "" {
		return "", fmt.Errorf("missing GEMINI_API_KEY")
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=" + geminiApiKey

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": question},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var gemResp struct {
		Candidates []struct {
			Content struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(bodyBytes, &gemResp); err != nil {
		return "", err
	}

	if len(gemResp.Candidates) == 0 {
		return "I don't have an answer for that.", nil
	}

	return gemResp.Candidates[0].Content.Text, nil
}

// sendFBMessage sends a text message back to the Facebook user
func sendFBMessage(userID, message string) error {
	fbPageAccessToken := os.Getenv("FB_PAGE_ACCESS_TOKEN")
	if fbPageAccessToken == "" {
		return fmt.Errorf("missing FB_PAGE_ACCESS_TOKEN")
	}

	msg := FBMessage{}
	msg.Recipient.ID = userID
	msg.Message.Text = message

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	url := "https://graph.facebook.com/v16.0/me/messages?access_token=" + fbPageAccessToken

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonMsg))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Facebook send error: %s", string(body))
	}

	return nil
}
