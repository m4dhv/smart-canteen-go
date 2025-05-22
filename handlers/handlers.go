package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"auth-website/database"
	"auth-website/models"

	"github.com/gorilla/sessions"
)

type Handler struct {
	DB    *database.DB
	Store *sessions.CookieStore
}

func NewHandler(db *database.DB, store *sessions.CookieStore) *Handler {
	return &Handler{
		DB:    db,
		Store: store,
	}
}

// Home page handler
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Feedback page handler
func (h *Handler) Feedback(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := template.ParseFiles("templates/feedback.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == "POST" {
		name := r.FormValue("name")
		email := r.FormValue("email")
		foodQualityStr := r.FormValue("food_quality")
		serviceStr := r.FormValue("service")
		comments := r.FormValue("comments")

		// Convert string ratings to integers
		foodQuality, err := strconv.Atoi(foodQualityStr)
		if err != nil || foodQuality < 1 || foodQuality > 5 {
			tmpl, _ := template.ParseFiles("templates/feedback.html")
			tmpl.Execute(w, map[string]string{"Error": "Invalid food quality rating"})
			return
		}

		service, err := strconv.Atoi(serviceStr)
		if err != nil || service < 1 || service > 5 {
			tmpl, _ := template.ParseFiles("templates/feedback.html")
			tmpl.Execute(w, map[string]string{"Error": "Invalid service rating"})
			return
		}

		// Save feedback to database
		err = h.DB.CreateFeedback(name, email, comments, foodQuality, service)
		if err != nil {
			tmpl, _ := template.ParseFiles("templates/feedback.html")
			tmpl.Execute(w, map[string]string{"Error": "Failed to submit feedback"})
			return
		}

		// Redirect with success message
		tmpl, _ := template.ParseFiles("templates/feedback.html")
		tmpl.Execute(w, map[string]string{"Success": "Thank you for your feedback!"})
	}
}

// Login page handler
func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := template.ParseFiles("templates/login.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := h.DB.ValidatePassword(username, password)
		if err != nil {
			tmpl, _ := template.ParseFiles("templates/login.html")
			tmpl.Execute(w, map[string]string{"Error": "Invalid username or password"})
			return
		}

		// Create session
		session, _ := h.Store.Get(r, "session-name")
		session.Values["user_id"] = user.ID
		session.Values["username"] = user.Username
		session.Values["role"] = user.Role
		session.Save(r, w)

		// Redirect based on role
		if user.Role == "admin" {
			http.Redirect(w, r, "/admin-dashboard", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
	}
}

// Registration page handler
func (h *Handler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := template.ParseFiles("templates/register.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		if password != confirmPassword {
			tmpl, _ := template.ParseFiles("templates/register.html")
			tmpl.Execute(w, map[string]string{"Error": "Passwords do not match"})
			return
		}

		err := h.DB.CreateUser(username, email, password)
		if err != nil {
			tmpl, _ := template.ParseFiles("templates/register.html")
			tmpl.Execute(w, map[string]string{"Error": "User already exists or invalid data"})
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

// User dashboard handler - now shows products
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "session-name")
	username, ok := session.Values["username"].(string)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	products, err := h.DB.GetAllProducts()
	if err != nil {
		http.Error(w, "Could not fetch products", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/dashboard.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Username string
		Products []models.Product
	}{
		Username: username,
		Products: products,
	}

	tmpl.Execute(w, data)
}

// Admin dashboard handler - now for product management
func (h *Handler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "session-name")
	role, ok := session.Values["role"].(string)
	if !ok || role != "admin" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	products, err := h.DB.GetAllProducts()
	if err != nil {
		http.Error(w, "Could not fetch products", http.StatusInternalServerError)
		return
	}

	users, err := h.DB.GetAllUsers()
	if err != nil {
		http.Error(w, "Could not fetch users", http.StatusInternalServerError)
		return
	}

	feedbacks, err := h.DB.GetAllFeedback()
	if err != nil {
		http.Error(w, "Could not fetch feedback", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/admin-dashboard.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Products  []models.Product
		Users     []models.User
		Feedbacks []models.Feedback
		Admin     string
	}{
		Products:  products,
		Users:     users,
		Feedbacks: feedbacks,
		Admin:     session.Values["username"].(string),
	}

	tmpl.Execute(w, data)
}

// Logout handler
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.Store.Get(r, "session-name")
	session.Values["user_id"] = nil
	session.Values["username"] = nil
	session.Values["role"] = nil
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Middleware to check authentication
func (h *Handler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := h.Store.Get(r, "session-name")
		if session.Values["user_id"] == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// Middleware to check admin role
func (h *Handler) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := h.Store.Get(r, "session-name")
		role, ok := session.Values["role"].(string)
		if !ok || role != "admin" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// Add product handler
func (h *Handler) AddProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := template.ParseFiles("templates/add-product.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == "POST" {
		name := r.FormValue("name")
		description := r.FormValue("description")
		priceStr := r.FormValue("price")
		imageURL := r.FormValue("image_url")
		stockStr := r.FormValue("stock")
		category := r.FormValue("category")

		// Convert price and stock to appropriate types
		price := 0.0
		stock := 0

		if priceStr != "" {
			if p, err := strconv.ParseFloat(priceStr, 64); err == nil {
				price = p
			}
		}

		if stockStr != "" {
			if s, err := strconv.Atoi(stockStr); err == nil {
				stock = s
			}
		}

		err := h.DB.CreateProduct(name, description, imageURL, category, price, stock)
		if err != nil {
			tmpl, _ := template.ParseFiles("templates/add-product.html")
			tmpl.Execute(w, map[string]string{"Error": "Failed to create product"})
			return
		}

		http.Redirect(w, r, "/admin-dashboard", http.StatusSeeOther)
	}
}

// Delete product handler
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		idStr := r.FormValue("product_id")
		if id, err := strconv.Atoi(idStr); err == nil {
			h.DB.DeleteProduct(id)
		}
	}
	http.Redirect(w, r, "/admin-dashboard", http.StatusSeeOther)
}

// Cart related handlers

// AddToCart handler processes the form submission to add a product to the cart
func (h *Handler) AddToCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// Get user ID from session
	session, _ := h.Store.Get(r, "session-name")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get product ID from form
	productIDStr := r.FormValue("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	// Get quantity from form (default to 1 if not specified)
	quantityStr := r.FormValue("quantity")
	quantity := 1
	if quantityStr != "" {
		if q, err := strconv.Atoi(quantityStr); err == nil && q > 0 {
			quantity = q
		}
	}

	// Add product to cart
	err = h.DB.AddToCart(userID, productID, quantity)
	if err != nil {
		if err == models.ErrInsufficientStock {
			// Redirect back with error message
			http.Redirect(w, r, "/dashboard?error=insufficient_stock", http.StatusSeeOther)
			return
		}
		http.Error(w, "Failed to add product to cart", http.StatusInternalServerError)
		return
	}

	// Redirect to cart page
	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

// ViewCart handler displays the user's cart
func (h *Handler) ViewCart(w http.ResponseWriter, r *http.Request) {
	// Get user ID from session
	session, _ := h.Store.Get(r, "session-name")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get cart from database
	cart, err := h.DB.GetUserCart(userID)
	if err != nil {
		http.Error(w, "Failed to load cart", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/cart.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Username string
		Cart     *models.Cart
	}{
		Username: session.Values["username"].(string),
		Cart:     cart,
	}

	tmpl.Execute(w, data)
}

// UpdateCartItem handler updates the quantity of an item in the cart
func (h *Handler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/cart", http.StatusSeeOther)
		return
	}

	// Get user ID from session
	session, _ := h.Store.Get(r, "session-name")
	_, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get cart item ID from form
	itemIDStr := r.FormValue("item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	// Get new quantity from form
	quantityStr := r.FormValue("quantity")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil || quantity < 0 {
		http.Error(w, "Invalid quantity", http.StatusBadRequest)
		return
	}

	// Update cart item
	err = h.DB.UpdateCartItemQuantity(itemID, quantity)
	if err != nil {
		if err == models.ErrInsufficientStock {
			http.Redirect(w, r, "/cart?error=insufficient_stock", http.StatusSeeOther)
			return
		}
		http.Error(w, "Failed to update cart", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

// RemoveCartItem handler removes an item from the cart
func (h *Handler) RemoveCartItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/cart", http.StatusSeeOther)
		return
	}

	// Get user ID from session
	session, _ := h.Store.Get(r, "session-name")
	_, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get cart item ID from form
	itemIDStr := r.FormValue("item_id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	// Remove item from cart
	err = h.DB.RemoveFromCart(itemID)
	if err != nil {
		http.Error(w, "Failed to remove item from cart", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

// ClearCart handler removes all items from the user's cart
func (h *Handler) ClearCart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/cart", http.StatusSeeOther)
		return
	}

	// Get user ID from session
	session, _ := h.Store.Get(r, "session-name")
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Clear the cart
	err := h.DB.ClearCart(userID)
	if err != nil {
		http.Error(w, "Failed to clear cart", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}
