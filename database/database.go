package database

import (
	"auth-website/models"
	"database/sql"
	"log"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
}

// Initialize creates and initializes the database
func Initialize() (*DB, error) {
	db, err := sql.Open("sqlite", "./auth.db")
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	dbInstance := &DB{db}

	// Create all tables
	if err := dbInstance.createTables(); err != nil {
		return nil, err
	}

	if err := dbInstance.createCartTables(); err != nil {
		return nil, err
	}

	if err := dbInstance.createFeedbackTable(); err != nil {
		return nil, err
	}

	// Create default admin user if it doesn't exist
	if err := dbInstance.createDefaultAdmin(); err != nil {
		log.Printf("Warning: Could not create default admin: %v", err)
	}

	return dbInstance, nil
}

// createTables creates the users and products tables
func (db *DB) createTables() error {
	// Create users table
	userTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT UNIQUE NOT NULL,
        email TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL,
        role TEXT DEFAULT 'user',
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`

	// Create products table
	productTable := `
    CREATE TABLE IF NOT EXISTS products (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        description TEXT,
        price REAL NOT NULL,
        image_url TEXT,
        stock INTEGER DEFAULT 0,
        category TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`

	_, err := db.Exec(userTable)
	if err != nil {
		return err
	}

	_, err = db.Exec(productTable)
	return err
}

// createCartTables creates the cart-related tables
func (db *DB) createCartTables() error {
	// Create cart table
	cartTable := `
    CREATE TABLE IF NOT EXISTS carts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES users(id)
    )`

	// Create cart items table
	cartItemsTable := `
    CREATE TABLE IF NOT EXISTS cart_items (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        cart_id INTEGER NOT NULL,
        product_id INTEGER NOT NULL,
        quantity INTEGER NOT NULL DEFAULT 1,
        FOREIGN KEY (cart_id) REFERENCES carts(id),
        FOREIGN KEY (product_id) REFERENCES products(id)
    )`

	_, err := db.Exec(cartTable)
	if err != nil {
		return err
	}

	_, err = db.Exec(cartItemsTable)
	return err
}

// createFeedbackTable creates the feedback table
func (db *DB) createFeedbackTable() error {
	feedbackTable := `
    CREATE TABLE IF NOT EXISTS feedback (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        email TEXT NOT NULL,
        food_quality INTEGER NOT NULL CHECK(food_quality >= 1 AND food_quality <= 5),
        service INTEGER NOT NULL CHECK(service >= 1 AND service <= 5),
        comments TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`

	_, err := db.Exec(feedbackTable)
	return err
}

// createDefaultAdmin creates a default admin user if none exists
func (db *DB) createDefaultAdmin() error {
	// Check if admin exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // Admin already exists
	}

	// Create default admin
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"INSERT INTO users (username, email, password, role) VALUES (?, ?, ?, ?)",
		"admin", "admin@example.com", string(hashedPassword), "admin",
	)
	return err
}

// USER RELATED METHODS

// CreateUser creates a new user in the database
func (db *DB) CreateUser(username, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		username, email, string(hashedPassword),
	)
	return err
}

// GetUserByUsername retrieves a user by their username
func (db *DB) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := db.QueryRow(
		"SELECT id, username, email, password, role, created_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Role, &user.CreatedAt)

	if err != nil {
		return nil, err
	}
	return user, nil
}

// ValidatePassword validates a user's password
func (db *DB) ValidatePassword(username, password string) (*models.User, error) {
	user, err := db.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetAllUsers retrieves all users from the database
func (db *DB) GetAllUsers() ([]models.User, error) {
	rows, err := db.Query("SELECT id, username, email, role, created_at FROM users ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// PRODUCT RELATED METHODS

// CreateProduct creates a new product in the database
func (db *DB) CreateProduct(name, description, imageURL, category string, price float64, stock int) error {
	_, err := db.Exec(
		"INSERT INTO products (name, description, price, image_url, stock, category) VALUES (?, ?, ?, ?, ?, ?)",
		name, description, price, imageURL, stock, category,
	)
	return err
}

// GetAllProducts retrieves all products from the database
func (db *DB) GetAllProducts() ([]models.Product, error) {
	rows, err := db.Query("SELECT id, name, description, price, image_url, stock, category, created_at FROM products ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.ImageURL, &product.Stock, &product.Category, &product.CreatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

// GetProductByID retrieves a product by its ID
func (db *DB) GetProductByID(id int) (*models.Product, error) {
	product := &models.Product{}
	err := db.QueryRow(
		"SELECT id, name, description, price, image_url, stock, category, created_at FROM products WHERE id = ?",
		id,
	).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.ImageURL, &product.Stock, &product.Category, &product.CreatedAt)

	if err != nil {
		return nil, err
	}
	return product, nil
}

// UpdateProduct updates an existing product
func (db *DB) UpdateProduct(id int, name, description, imageURL, category string, price float64, stock int) error {
	_, err := db.Exec(
		"UPDATE products SET name = ?, description = ?, price = ?, image_url = ?, stock = ?, category = ? WHERE id = ?",
		name, description, price, imageURL, stock, category, id,
	)
	return err
}

// DeleteProduct deletes a product by its ID
func (db *DB) DeleteProduct(id int) error {
	_, err := db.Exec("DELETE FROM products WHERE id = ?", id)
	return err
}

// FEEDBACK RELATED METHODS

// CreateFeedback creates a new feedback entry in the database
func (db *DB) CreateFeedback(name, email, comments string, foodQuality, service int) error {
	_, err := db.Exec(
		"INSERT INTO feedback (name, email, food_quality, service, comments) VALUES (?, ?, ?, ?, ?)",
		name, email, foodQuality, service, comments,
	)
	return err
}

// GetAllFeedback retrieves all feedback from the database
func (db *DB) GetAllFeedback() ([]models.Feedback, error) {
	rows, err := db.Query("SELECT id, name, email, food_quality, service, comments, created_at FROM feedback ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedbacks []models.Feedback
	for rows.Next() {
		var feedback models.Feedback
		err := rows.Scan(&feedback.ID, &feedback.Name, &feedback.Email, &feedback.FoodQuality, &feedback.Service, &feedback.Comments, &feedback.CreatedAt)
		if err != nil {
			return nil, err
		}
		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, nil
}

// CART RELATED METHODS

// GetUserCart gets or creates a cart for a user
func (db *DB) GetUserCart(userID int) (*models.Cart, error) {
	var cartID int
	var createdAt string

	// First, check if the user already has a cart
	err := db.QueryRow("SELECT id, created_at FROM carts WHERE user_id = ?", userID).Scan(&cartID, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create a new cart for the user
			result, err := db.Exec("INSERT INTO carts (user_id) VALUES (?)", userID)
			if err != nil {
				return nil, err
			}

			newCartID, err := result.LastInsertId()
			if err != nil {
				return nil, err
			}

			cartID = int(newCartID)
			createdAt = "" // Default timestamp will be set by SQLite
		} else {
			return nil, err
		}
	}

	// Get cart items
	cart := &models.Cart{
		ID:     cartID,
		UserID: userID,
		Items:  []models.CartItem{},
	}

	// Get cart items with product details
	rows, err := db.Query(`
		SELECT ci.id, ci.product_id, ci.quantity, p.name, p.price, p.image_url
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.cart_id = ?
	`, cartID)
	if err != nil {
		return cart, nil // Return empty cart on error
	}
	defer rows.Close()

	var totalPrice float64 = 0
	for rows.Next() {
		var item models.CartItem
		var productName, imageURL string
		var productPrice float64

		err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &productName, &productPrice, &imageURL)
		if err != nil {
			continue
		}

		item.Product = models.Product{
			ID:       item.ProductID,
			Name:     productName,
			Price:    productPrice,
			ImageURL: imageURL,
		}

		// Calculate item total
		item.ItemTotal = productPrice * float64(item.Quantity)
		totalPrice += item.ItemTotal

		cart.Items = append(cart.Items, item)
	}

	cart.TotalPrice = totalPrice
	return cart, nil
}

// AddToCart adds a product to the user's cart
func (db *DB) AddToCart(userID, productID, quantity int) error {
	// First, check if product has enough stock
	var stock int
	err := db.QueryRow("SELECT stock FROM products WHERE id = ?", productID).Scan(&stock)
	if err != nil {
		return err
	}

	if stock < quantity {
		return models.ErrInsufficientStock
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get or create cart
	var cartID int
	err = tx.QueryRow("SELECT id FROM carts WHERE user_id = ?", userID).Scan(&cartID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create a new cart
			result, err := tx.Exec("INSERT INTO carts (user_id) VALUES (?)", userID)
			if err != nil {
				return err
			}

			newCartID, err := result.LastInsertId()
			if err != nil {
				return err
			}
			cartID = int(newCartID)
		} else {
			return err
		}
	}

	// Check if product already exists in cart
	var existingItemID, existingQuantity int
	err = tx.QueryRow("SELECT id, quantity FROM cart_items WHERE cart_id = ? AND product_id = ?", cartID, productID).Scan(&existingItemID, &existingQuantity)
	if err == nil {
		// Product already in cart, update quantity
		totalQuantity := existingQuantity + quantity
		if totalQuantity > stock {
			return models.ErrInsufficientStock
		}

		_, err = tx.Exec("UPDATE cart_items SET quantity = ? WHERE id = ?", totalQuantity, existingItemID)
		if err != nil {
			return err
		}
	} else if err == sql.ErrNoRows {
		// Product not in cart, add it
		_, err = tx.Exec("INSERT INTO cart_items (cart_id, product_id, quantity) VALUES (?, ?, ?)", cartID, productID, quantity)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	// Update product stock
	_, err = tx.Exec("UPDATE products SET stock = stock - ? WHERE id = ?", quantity, productID)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// UpdateCartItemQuantity updates the quantity of an item in the cart
func (db *DB) UpdateCartItemQuantity(cartItemID, newQuantity int) error {
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get current quantity and product ID
	var currentQuantity, productID int
	err = tx.QueryRow("SELECT quantity, product_id FROM cart_items WHERE id = ?", cartItemID).Scan(&currentQuantity, &productID)
	if err != nil {
		return err
	}

	// Calculate the difference in quantity
	quantityChange := newQuantity - currentQuantity

	// If decreasing quantity
	if quantityChange < 0 {
		// Return items to stock
		_, err = tx.Exec("UPDATE products SET stock = stock + ? WHERE id = ?", -quantityChange, productID)
		if err != nil {
			return err
		}
		
		// Update cart item quantity or remove if zero
		if newQuantity <= 0 {
			_, err = tx.Exec("DELETE FROM cart_items WHERE id = ?", cartItemID)
		} else {
			_, err = tx.Exec("UPDATE cart_items SET quantity = ? WHERE id = ?", newQuantity, cartItemID)
		}
		if err != nil {
			return err
		}
	} else if quantityChange > 0 {
		// If increasing quantity, check stock
		var availableStock int
		err = tx.QueryRow("SELECT stock FROM products WHERE id = ?", productID).Scan(&availableStock)
		if err != nil {
			return err
		}

		if availableStock < quantityChange {
			return models.ErrInsufficientStock
		}

		// Decrease stock
		_, err = tx.Exec("UPDATE products SET stock = stock - ? WHERE id = ?", quantityChange, productID)
		if err != nil {
			return err
		}

		// Update cart item quantity
		_, err = tx.Exec("UPDATE cart_items SET quantity = ? WHERE id = ?", newQuantity, cartItemID)
		if err != nil {
			return err
		}
	}

	// Commit transaction
	return tx.Commit()
}

// RemoveFromCart removes an item from the cart
func (db *DB) RemoveFromCart(cartItemID int) error {
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get quantity and product ID before deleting
	var quantity, productID int
	err = tx.QueryRow("SELECT quantity, product_id FROM cart_items WHERE id = ?", cartItemID).Scan(&quantity, &productID)
	if err != nil {
		return err
	}

	// Return items to stock
	_, err = tx.Exec("UPDATE products SET stock = stock + ? WHERE id = ?", quantity, productID)
	if err != nil {
		return err
	}

	// Remove item from cart
	_, err = tx.Exec("DELETE FROM cart_items WHERE id = ?", cartItemID)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// ClearCart removes all items from a user's cart
func (db *DB) ClearCart(userID int) error {
	// Get cart ID
	var cartID int
	err := db.QueryRow("SELECT id FROM carts WHERE user_id = ?", userID).Scan(&cartID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // No cart to clear
		}
		return err
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get all cart items to return stock
	rows, err := tx.Query("SELECT product_id, quantity FROM cart_items WHERE cart_id = ?", cartID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Return stock for each item
	for rows.Next() {
		var productID, quantity int
		if err := rows.Scan(&productID, &quantity); err != nil {
			return err
		}

		_, err = tx.Exec("UPDATE products SET stock = stock + ? WHERE id = ?", quantity, productID)
		if err != nil {
			return err
		}
	}

	// Delete all cart items
	_, err = tx.Exec("DELETE FROM cart_items WHERE cart_id = ?", cartID)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}