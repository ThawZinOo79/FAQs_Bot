package controllers

import (
	"faqs-bot/config"
	"faqs-bot/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateCustomer(c *gin.Context) {
	var input models.Customer

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.OrderDate = time.Now()

	if err := config.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func GetCustomers(c *gin.Context) {
	// Pagination params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.DefaultQuery("search", "")

	var customers []models.Customer
	offset := (page - 1) * limit
	query := config.DB.Model(&models.Customer{})

	if search != "" {
		query = query.Where("LOWER(account_link) LIKE ?", "%"+strings.ToLower(search)+"%")
	}

	if err := query.Offset(offset).Limit(limit).Find(&customers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customers"})
		return
	}

	c.JSON(http.StatusOK, customers)
}

func UpdateCustomer(c *gin.Context) {
	id := c.Param("id")
	var customer models.Customer

	if err := config.DB.First(&customer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	var input models.Customer
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer.AccountLink = input.AccountLink
	customer.OrderDate = time.Now()

	config.DB.Save(&customer)

	c.JSON(http.StatusOK, customer)
}

func DeleteCustomer(c *gin.Context) {
	id := c.Param("id")
	var customer models.Customer

	if err := config.DB.First(&customer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	config.DB.Delete(&customer)
	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted"})
}
