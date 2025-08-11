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
	"time"

	"github.com/gin-gonic/gin"
)

var (
	verifyToken = "d306e0fcae497171a5511d8854ed8e3dcf7c1c0b01a84954e42280d44a6e83f1"
	pageToken   = "EAAQLPtfjYSEBPH7ZAaHPlZBxpWETJh4JXvDt0U0qnsSq7UjdQoWkzTuO3Td99dIG989kU6x8ZCky7wFCcmR8dbFKDBI6FWhleCZAZADRCZBpOkhpiTlcdDMlnh65EUGZCJXK1OJRLVijJZAPfdGAOGze0c2ric3sOF3ayX7XY7QgpYVViG8GAi56BshcfGVi1fZCjoMa7"
	geminiKey   = "AIzaSyD8nPAj0OkNZK9ilBczDqFD1ROa6ZJGC40"

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
	Message   *Message  `json:"message,omitempty"`
	Postback  *Postback `json:"postback,omitempty"`
}

type Postback struct {
	Payload string `json:"payload"`
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
func CallGeminiIntent(userMessage string) (string, string) {
	prompt := fmt.Sprintf(`You are an assistant for Aung Mobile. Your job is to understand the user's intent.

Respond using this JSON format only:

- Product list request:
  {"intent": "productlist"}

- Specific product detail:
  {"intent": "productdetail", "query": "<product name>"}

- Product availability (e.g. "When will iPhone 11 be available?"):
  {"intent": "availability", "query": "<product name>"}

- Order placement:
  {"intent": "order", "query": "<product name>"}

- FAQ questions (return, warranty, trade-in, accessories, repair):
  {"intent": "faq", "query": "<user question>"}

Do not explain anything. Only return JSON.

User message: %s`, userMessage)

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + geminiKey

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	jsonBody, _ := json.Marshal(requestBody)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println("CallGeminiIntent error:", err)
		return "", ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", ""
	}

	candidates, ok := result["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return "", ""
	}

	content := candidates[0].(map[string]interface{})["content"]
	parts := content.(map[string]interface{})["parts"].([]interface{})
	if len(parts) == 0 {
		return "", ""
	}

	text := parts[0].(map[string]interface{})["text"].(string)

	var data map[string]string
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		fmt.Println("Gemini response is not valid JSON:", text)
		return "", ""
	}

	return data["intent"], data["query"]
}
func GetAvailabilityMessage(productName, userMessage string) string {
	var product models.Inventory
	err := config.DB.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(productName)+"%").First(&product).Error
	if err != nil {
		return "❌ Product not found"
	}

	if product.Stock > 0 {
		return "" // No need to ask Gemini if it's in stock
	}

	// Prepare a friendly message
	availableTime := product.AvailableTime.Format("2006-01-02") // you can use time.RFC3339 or customize
	return CallGeminiAvailabilityReply(userMessage, availableTime)
}
func CallGeminiAvailabilityReply(userMessage, availableDate string) string {
	instruction := `
You are a smart, friendly assistant for Aung Mobile Second Phone Service.
Reply to customers naturally and casually based on when a product will be back in stock.
If the customer is asking about availability and the product is out of stock, generate a friendly short reply with emojis.
Reply in the same language the customer used.

English example:
User: When will it be available?
Reply: 📦 This item will be back in about 2 days. Please stay tuned!

Myanmar example:
User: ဘယ်တော့ရမလဲ
Reply: 📦 ဒီပစ္စည်းကို နောက် ၂ ရက်လောက်အတွင်း ပြန်ရနိုင်မယ်နော်။

Use a warm tone with emoji. Keep it short (1–2 sentences). Use relative terms like "tomorrow", "next week", "2 days", or "next month".

Today’s date: {{today}}
Available date: {{availableDate}}

Now reply to this customer message:
User: {{userMessage}}
`

	// Replace placeholders
	today := time.Now().Format("2006-01-02")
	instruction = strings.ReplaceAll(instruction, "{{today}}", today)
	instruction = strings.ReplaceAll(instruction, "{{availableDate}}", availableDate)
	instruction = strings.ReplaceAll(instruction, "{{userMessage}}", userMessage)

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + geminiKey

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": instruction},
				},
			},
		},
	}

	jsonBody, _ := json.Marshal(requestBody)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Println("Gemini error:", err)
		return "I'm having trouble checking the product availability."
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

func GenerateFilteredProductFlex(query string) map[string]interface{} {
	var inventories []models.Inventory
	config.DB.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(query)+"%").Find(&inventories)

	if len(inventories) == 0 {
		return map[string]interface{}{
			"attachment": map[string]interface{}{
				"type": "template",
				"payload": map[string]interface{}{
					"template_type": "button",
					"text":          "❌ Sorry, we couldn't find any product matching: " + query,
					"buttons": []map[string]string{
						{
							"type":    "postback",
							"title":   "Show All Products",
							"payload": "SHOW_ALL_PRODUCTS",
						},
					},
				},
			},
		}
	}

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

// Handle incoming webhook POST
func HandleMessage(c *gin.Context) {
	var event FBWebhookEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)

	for _, entry := range event.Entry {
		for _, msg := range entry.Messaging {
			if msg.Sender.ID == pageID {
				continue
			}

			userID := msg.Sender.ID
			var userMsg string

			// 🟡 Support both text messages and postback payloads
			if msg.Message != nil && msg.Message.Text != "" {
				userMsg = msg.Message.Text
			} else if msg.Postback != nil && msg.Postback.Payload != "" {
				userMsg = msg.Postback.Payload
			} else {
				continue // skip if no valid message or payload
			}

			fmt.Println("Received message from:", userID, "text:", userMsg)

			go func(uid, message string) {
				// 🟢 Directly handle ORDER postback payloads
				if strings.HasPrefix(message, "ORDER_") {
					SendReply(uid, "📦 မှာစာအတွက်ကျေးဇူးတင်ပါတယ်နော်။ ငွေပေးချေမှုအတွက် admin ဆက်သွယ်ပါလိမ့်မယ် 🙏")
					return
				}

				// 🔍 Intent detection with Gemini
				intent, query := CallGeminiIntent(message)

				switch intent {
				case "productlist":
					SendFlexReply(uid, GenerateProductListFlex())
				case "productdetail":
					if strings.Contains(strings.ToLower(message), "available") {
						availabilityMsg := GetAvailabilityMessage(query, message)
						SendReply(uid, availabilityMsg)
					} else {
						SendFlexReply(uid, GenerateFilteredProductFlex(query))
					}
				case "availability":
					availabilityMsg := GetAvailabilityMessage(query, message)
					SendReply(uid, availabilityMsg)
				case "order":
					SendReply(uid, "📦 မှာစာအတွက်ကျေးဇူးတင်ပါတယ်နော်။ ငွေပေးချေမှုအတွက် admin ဆက်သွယ်ပါလိမ့်မယ် 🙏")
				case "faq":
					reply := CallGeminiWithCompanyProfile(query)
					SendReply(uid, reply)
				default:
					reply := CallGeminiWithCompanyProfile(message)
					SendReply(uid, reply)
				}
			}(userID, userMsg)
		}
	}
}

// Call Gemini with company profile and instruction
func CallGeminiWithCompanyProfile(userMessage string) string {
	instruction := `
You are a smart, friendly assistant for Aung Mobile Second Phone Service.
Your job is to reply to customers with short, warm, casual messages using emojis.
Only send customer-facing replies – no explanations or extra context.
Use clear, casual, positive language, suitable for chat or SMS.
Always respond in the same language the customer uses. If the user types in Burmese (Myanmar), reply in Burmese.

🟡 If the customer asks "When will this product be available again?" and the product is out of stock, check the available time from the database and answer accordingly.

  - Example in English: "🚚 This item will be back in stock on August 8, 2025 📦"
  - Example in Myanmar: "📅 ဒီပစ္စည်း August 8, 2025 မှာပြန်ရမယ်နော် 📦"

🟢 If the customer places an order, thank them and tell them to wait for admin to contact for the payment.

  - Example in English: "🙏 Thanks for your order! Please wait while our admin helps you with payment 💵"
  - Example in Myanmar: "📦 မှာစာအတွက်ကျေးဇူးတင်ပါတယ်နော်။ ငွေပေးချေမှုအတွက် admin ဆက်သွယ်ပါလိမ့်မယ် 🙏"

🔁 If the customer asks about **returns, refunds, or exchanges** like "ဝယ်ပြီးပြန်လဲလို့ရလား", explain politely that second-hand phones can't be returned but are fully tested.

  - Myanmar example: "🙏 ဝယ်ပြီးပြန်လဲတာ မရနိုင်ပါဘူးနော်၊ ဒါပေမဲ့ ဝယ်ဖို့ကောင်းအောင် စစ်ဆေးပြီးပေးတယ် 📱"

  📱 If the user asks for phone suggestions like "Which phone should I buy?" or compares brands like "iPhone vs Samsung", give a friendly, short suggestion based on general preferences:
- Recommend iPhone if user prefers performance, camera, and ecosystem.
- Recommend Samsung if user wants flexibility, customization, and better value for money.

Examples:
English: "📱 If you like smooth performance and camera, go for iPhone. But if you want customization and better value, Samsung is a great choice too 😎"
Myanmar: "📱 မိမိအတွက် အမြန်ဆန်ပြီး ကင်မရာကောင်းတဲ့ဖုန်းလိုချင်ရင် iPhone ကောင်းပါတယ်၊ မူလတန်းအရ လှုပ်ရှားနိုင်မှုနဲ့ စျေးနှုန်းလည်း ပြေလည်ချင်ရင်တော့ Samsung ကောင်းပါတယ်နော် 😄"

❓ Common FAQ patterns and replies:
- Warranty: "🛡️ စက်တွေမှာ စမ်းသပ်ပြီးဖြစ်ပြီး အာမခံနည်းနည်းရှိပါတယ်နော်"
- Accessories included? "🎧 ဖုန်းနဲ့အတူ အခြားအစိတ်အပိုင်းတွေပါဝင်မလားဆိုတာ Modelပေါ် မူတည်ပါတယ်နော်"
- Trade-in available? "🔁 ပစ္စည်းအဟောင်းပြန်လဲဝယ်တာလည်း ရပါတယ်။ ဝင်ရောက်မေးမြန်းလို့ရပါတယ်နော်"
- Repair service? "🔧 ဖုန်းပြင်ခန်းဝန်ဆောင်မှုလည်း ရပါတယ်နော်"

Mention services like:
📱 Second-hand smartphones
🔧 Repairs & maintenance
🎧 Accessories
🔁 Trade-in offers
📦 Promotions

Keep replies 1–2 short sentences with a friendly tone and emojis.
`

	companyProfile := `
Company: Aung Mobile Second Phone Service
Owner: Min Thway Khaing
Phone: 0650125735
Overview: We sell quality second-hand phones, offer repairs and accessories.
All devices are tested and include limited warranties.
Follow our Facebook Page for new arrivals!`

	prompt := instruction + "\n\n" + companyProfile + "\n\nCustomer: " + userMessage

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + geminiKey

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
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
