package controllers

import (
	"faqs-bot/config"
	"faqs-bot/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetFaqs(c *gin.Context) {
	var faqs []models.Faq
	config.DB.Find(&faqs)
	c.JSON(http.StatusOK, faqs)
}

func CreateFaq(c *gin.Context) {
	var faq models.Faq
	if err := c.ShouldBindJSON(&faq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	config.DB.Create(&faq)
	c.JSON(http.StatusOK, faq)
}
