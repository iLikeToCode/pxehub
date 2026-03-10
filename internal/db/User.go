package db

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string
	Username string
	Password string
}

func CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func CreateUser(name, username, password string, db *gorm.DB) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := User{
		Name:     name,
		Username: username,
		Password: string(hashedPassword),
	}

	if err := db.Create(&user).Error; err != nil {
		return err
	}
	return nil
}

func GetUserByID(id uint, db *gorm.DB) (*User, error) {
	var user User
	if err := db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByUsername(username string, db *gorm.DB) (*User, error) {
	var user User
	if err := db.First(&user, username).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUser(id uint, name, username, password string, db *gorm.DB) error {
	var user User
	if err := db.First(&user, id).Error; err != nil {
		return err
	}

	user.Name = name
	user.Username = username

	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	}

	if err := db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

func DeleteUser(id uint, db *gorm.DB) error {
	if err := db.Delete(&User{}, id).Error; err != nil {
		return err
	}
	return nil
}
