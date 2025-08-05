package repositories

import (
	"errors"
	"faqs-bot/config"
	"faqs-bot/models"

	"golang.org/x/crypto/bcrypt"
)

func FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := config.DB.Where("email = ?", email).First(&user)
	return &user, result.Error
}

func RegisterUser(email, password, username, phone string) (*models.User, error) {
	var existing models.User
	result := config.DB.Where("email = ?", email).First(&existing)
	if result.Error == nil {
		return nil, errors.New("email already registered")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Email:       email,
		Password:    string(hashed),
		Username:    username,
		PhoneNumber: phone,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
