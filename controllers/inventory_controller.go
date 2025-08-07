package controllers

import (
	"faqs-bot/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllProductsHandler(c *gin.Context) {
	products, err := repositories.GetAllProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}
