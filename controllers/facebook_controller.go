package controllers

import (
	"bytes"
	"encoding/json"
	"faqs-bot/config"
	"faqs-bot/models"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	verifyToken = "d306e0fcae497171a5511d8854ed8e3dcf7c1c0b01a84954e42280d44a6e83f1"
	pageToken   = "EAAQLPtfjYSEBPH7ZAaHPlZBxpWETJh4JXvDt0U0qnsSq7UjdQoWkzTuO3Td99dIG989kU6x8ZCky7wFCcmR8dbFKDBI6FWhleCZAZADRCZBpOkhpiTlcdDMlnh65EUGZCJXK1OJRLVijJZAPfdGAOGze0c2ric3sOF3ayX7XY7QgpYVViG8GAi56BshcfGVi1fZCjoMa7"
	geminiKey   = "AIzaSyD8nPAj0OkNZK9ilBczDqFD1ROa6ZJGC40"

	// Set your actual Facebook Page ID here to prevent message loop
	pageID = "109400371618421"
)

// Facebook webhook event structs
type FBWebhookEvent struct {
	Object string `json:"object"`
	Entry  []struct {
		Messaging []FBMessaging `json:"messaging"`
	} `json:"entry"`
}

type FBMessaging struct {
	Sender    Sender    `json:"sender"`
	Recipient Recipient `json:"recipient"`
	Message   Message   `json:"message"`
}

type Sender struct {
	ID string `json:"id"`
}

type Recipient struct {
	ID string `json:"id"`
}

type Message struct {
	Text string `json:"text"`
}

// Verify webhook GET
func VerifyWebhook(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode == "subscribe" && token == verifyToken {
		c.String(http.StatusOK, challenge)
	} else {
		c.String(http.StatusForbidden, "Verification failed")
	}
}

// Handle incoming webhook POST
func HandleMessage(c *gin.Context) {
	var event FBWebhookEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Respond 200 ASAP to stop retries
	c.Status(http.StatusOK)

	for _, entry := range event.Entry {
		for _, msg := range entry.Messaging {
			// Ignore messages sent by the page itself to prevent loop
			if msg.Sender.ID == pageID {
				continue
			}

			// Skip messages with no text (could be delivery receipts, etc.)
			if msg.Message.Text == "" {
				continue
			}

			userID := msg.Sender.ID
			userMsg := msg.Message.Text

			fmt.Println("Received message from:", userID, "text:", userMsg)

			// Simple keyword-based reply logic
			if strings.Contains(strings.ToLower(userMsg), "product") || strings.Contains(strings.ToLower(userMsg), "list") {
				go SendFlexReply(userID, GenerateProductListFlex())
			} else {
				go func(uid, message string) {
					reply := CallGemini(message)
					SendReply(uid, reply)
				}(userID, userMsg)
			}
		}
	}
}

// Call Gemini 2.0 Flash API to generate reply text
func CallGemini(userMessage string) string {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + geminiKey

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": userMessage},
				},
			},
		},
	}

	jsonBody, _ := json.Marshal(requestBody)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println("CallGemini error:", err)
		return "I'm having trouble replying right now."
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "Sorry, I couldn't understand the response."
	}

	candidates, ok := result["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return "No response generated."
	}

	content := candidates[0].(map[string]interface{})["content"]
	parts := content.(map[string]interface{})["parts"].([]interface{})
	if len(parts) > 0 {
		return parts[0].(map[string]interface{})["text"].(string)
	}

	return "Sorry, no meaningful response."
}

// Generate product list as Facebook Generic Template
func GenerateProductListFlex() map[string]interface{} {
	var inventories []models.Inventory
	config.DB.Find(&inventories)

	var elements []map[string]interface{}
	for _, item := range inventories {
		stockStatus := "✅ In Stock"
		availability := ""

		if item.Stock <= 0 {
			stockStatus = "❌ Out of Stock"
			availability = "Available at: " + item.AvailableTime.Format("Jan 2, 2006 15:04")
		}

		subtitle := fmt.Sprintf("💵 %.0f MMK\n%s\n%s", item.Price, stockStatus, availability)

		elements = append(elements, map[string]interface{}{
			"title":     item.Name,
			"image_url": item.ImageURL,
			"subtitle":  subtitle,
			"buttons": []map[string]string{
				{
					"type":    "postback",
					"title":   "Order 📦",
					"payload": fmt.Sprintf("ORDER_%d", item.ID),
				},
			},
		})
	}

	// Final Facebook message payload
	return map[string]interface{}{
		"attachment": map[string]interface{}{
			"type": "template",
			"payload": map[string]interface{}{
				"template_type": "generic",
				"elements":      elements,
			},
		},
	}
}

// Send simple text message via Facebook Send API
func SendReply(userID, message string) {
	url := fmt.Sprintf("https://graph.facebook.com/v17.0/me/messages?access_token=%s", pageToken)

	body := map[string]interface{}{
		"recipient": map[string]string{"id": userID},
		"message":   map[string]string{"text": message},
	}

	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println("SendReply error:", err)
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println("SendReply response:", string(respBody))
}

// Send Flex (generic template) message via Facebook Send API
func SendFlexReply(userID string, message map[string]interface{}) {
	url := fmt.Sprintf("https://graph.facebook.com/v17.0/me/messages?access_token=%s", pageToken)

	body := map[string]interface{}{
		"recipient": map[string]string{"id": userID},
		"message":   message,
	}

	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println("SendFlexReply error:", err)
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println("SendFlexReply response:", string(respBody))
}
