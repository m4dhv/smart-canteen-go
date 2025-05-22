package models

import (
	"errors"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Don't include password in JSON responses
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	ImageURL    string    `json:"image_url"`
	Stock       int       `json:"stock"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
}

// Cart related models
type Cart struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	Items      []CartItem `json:"items"`
	TotalPrice float64    `json:"total_price"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CartItem struct {
	ID        int     `json:"id"`
	ProductID int     `json:"product_id"`
	Product   Product `json:"product"`
	Quantity  int     `json:"quantity"`
	ItemTotal float64 `json:"item_total"`
}

// Feedback model
type Feedback struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	FoodQuality int       `json:"food_quality"`
	Service     int       `json:"service"`
	Comments    string    `json:"comments"`
	CreatedAt   time.Time `json:"created_at"`
}

// Custom errors
var (
	ErrInsufficientStock = errors.New("insufficient stock")
)