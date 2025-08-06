package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"faqs-bot/models"
	"faqs-bot/repositories"

	"github.com/gin-gonic/gin"
)

// --- Facebook Webhook Event ---
type FBWebhookEvent struct {
	Object string `json:"object"`
	Entry  []struct {
		Messaging []struct {
			Mid    string `json:"mid"`
			Sender struct {
				ID string `json:"id"`
			} `json:"sender"`
			Message struct {
				Text string `json:"text"`
			} `json:"message"`
			Postback struct {
				Payload string `json:"payload"`
			} `json:"postback"`
		} `json:"messaging"`
	} `json:"entry"`
}

// --- Facebook Message Format ---
type FBMessage struct {
	Recipient struct {
		ID string `json:"id"`
	} `json:"recipient"`
	Message interface{} `json:"message"`
}

// --- In-Memory Storage ---
var (
	processedMessages = make(map[string]bool)
	mu                sync.Mutex
	geminiInstruction string
	companyProfile    string
)

// --- Load Text Files ---
func LoadTexts() {
	inst, err := os.ReadFile("C:/Users/User/Desktop/Programming/Go/FAQs_Bot/gemini_instruction.txt")
	if err != nil {
		log.Println("Could not read Gemini instruction:", err)
		geminiInstruction = "You are a friendly assistant for Innovation Technology. Reply casually and warmly."
	} else {
		geminiInstruction = string(inst)
	}

	profile, err := os.ReadFile("C:/Users/User/Desktop/Programming/Go/FAQs_Bot/company_profile.txt")
	if err != nil {
		log.Println("Could not read company profile:", err)
		companyProfile = "Innovation Technology - Your trusted IT Software House."
	} else {
		companyProfile = string(profile)
	}
}

// --- Verify Webhook ---
func FBWebhookVerify(c *gin.Context) {
	if c.Query("hub.verify_token") == os.Getenv("FB_VERIFY_TOKEN") {
		c.String(http.StatusOK, c.Query("hub.challenge"))
	} else {
		c.String(http.StatusForbidden, "Verification failed")
	}
}

// --- Handle Facebook Message ---
func FBWebhookReceive(c *gin.Context) {
	var event FBWebhookEvent
	if err := c.BindJSON(&event); err != nil {
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}

	for _, entry := range event.Entry {
		for _, msg := range entry.Messaging {
			mu.Lock()
			if processedMessages[msg.Mid] {
				mu.Unlock()
				c.Status(http.StatusOK)
				return
			}
			processedMessages[msg.Mid] = true
			mu.Unlock()

			senderID := msg.Sender.ID
			userInput := strings.TrimSpace(msg.Message.Text)

			// Handle Postback
			if msg.Postback.Payload != "" {
				handlePostback(senderID, msg.Postback.Payload)
				c.Status(http.StatusOK)
				return
			}

			// Handle Gemini Reply
			if userInput != "" {
				reply := getGeminiReply(userInput)
				sendText(senderID, reply)
				c.Status(http.StatusOK)
				return
			}
		}
	}
	c.Status(http.StatusOK)
}

// --- Gemini Reply Function ---
func getGeminiReply(userText string) string {
	promptText := fmt.Sprintf("%s\n\nCompany Info:\n%s\n\nCustomer Message:\n%s",
		geminiInstruction, companyProfile, userText)

	requestBody := map[string]interface{}{
		"prompt": map[string]interface{}{
			"messages": []map[string]interface{}{
				{
					"author": "user",
					"content": map[string]string{
						"text": promptText,
					},
				},
			},
		},
		"temperature":    0.7,
		"candidateCount": 1,
		"topP":           0.95,
		"topK":           40,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return "Sorry, something went wrong. 😔"
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2-0-flash:generateMessage?key=" + os.Getenv("GEMINI_API_KEY")

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Println("Gemini request error:", err)
		return "Sorry, something went wrong. 😔"
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading Gemini response body:", err)
		return "Couldn't process your request. 😔"
	}

	// Log full response body for debugging
	log.Println("Gemini raw response:", string(bodyBytes))

	var result struct {
		Candidates []struct {
			Content struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		log.Println("Gemini response decode error:", err)
		return "Couldn't process your request. 😔"
	}

	if len(result.Candidates) > 0 {
		return result.Candidates[0].Content.Text
	}

	return "Hi! 😊 How can we help you today?"
}

// --- Postback Handler ---
func handlePostback(userID, payload string) {
	if strings.HasPrefix(payload, "CATEGORY_") {
		sendText(userID, "Here are details for "+strings.TrimPrefix(payload, "CATEGORY_")+" 📦")
	} else if strings.HasPrefix(payload, "ORDER_") {
		productID := strings.TrimPrefix(payload, "ORDER_")
		product, err := repositories.GetProductByIDString(productID)
		if err == nil {
			saveOrder(userID, product.Name)
		} else {
			sendText(userID, "Product not found.")
		}
	}
}

// --- Order Handling ---
func saveOrder(userID, productName string) {
	product, err := repositories.GetProductByName(productName)
	if err != nil {
		sendText(userID, "Product not found.")
		return
	}

	customer := models.Customer{
		OrderDate:   time.Now(),
		AccountLink: userID,
	}
	repositories.CreateCustomer(&customer)

	_, orderErr := repositories.CreateOrder(customer.ID, product.ID, userID, "Messenger Order")
	if orderErr != nil {
		sendText(userID, "Failed to place order.")
		return
	}

	sendText(userID, fmt.Sprintf("Order placed for %s ✅", product.Name))
}

// --- Facebook Send Helpers ---
func sendText(userID, text string) {
	if userID == "" {
		log.Println("sendText: empty userID, skipping send")
		return
	}
	msg := FBMessage{}
	msg.Recipient.ID = userID
	msg.Message = map[string]string{"text": text}
	sendToFacebook(msg)
}

func sendGenericTemplate(userID string, elements []map[string]interface{}) {
	payload := map[string]interface{}{
		"attachment": map[string]interface{}{
			"type": "template",
			"payload": map[string]interface{}{
				"template_type": "generic",
				"elements":      elements,
			},
		},
	}
	msg := FBMessage{}
	msg.Recipient.ID = userID
	msg.Message = payload
	sendToFacebook(msg)
}

func sendToFacebook(msg FBMessage) {
	url := "https://graph.facebook.com/v16.0/me/messages?access_token=" + os.Getenv("FB_PAGE_ACCESS_TOKEN")
	jsonMsg, _ := json.Marshal(msg)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonMsg))
	if err != nil {
		log.Println("Facebook send error:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Println("Facebook send error:", string(body))
	}
}
