package main

import (
	"auth-website/database"
	"auth-website/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func main() {
	// Initialize database
	db, err := database.Initialize()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()
	// Initialize session store
	store := sessions.NewCookieStore([]byte("your-secret-key-change-this-in-production"))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 1 week
		HttpOnly: true,
	}
	// Initialize handlers
	h := handlers.NewHandler(db, store)
	// Setup router
	r := mux.NewRouter()
	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	// Public routes
	r.HandleFunc("/", h.Home).Methods("GET")
	r.HandleFunc("/home", h.Home).Methods("GET")
	r.HandleFunc("/index", h.Home).Methods("GET")
	r.HandleFunc("/login", h.LoginPage).Methods("GET", "POST")
	r.HandleFunc("/register", h.RegisterPage).Methods("GET", "POST")
	r.HandleFunc("/logout", h.Logout).Methods("GET")
	r.HandleFunc("/feedback", h.Feedback).Methods("GET", "POST")
	r.HandleFunc("/submit_feedback", h.Feedback).Methods("POST")
	// Protected routes
	r.HandleFunc("/dashboard", h.RequireAuth(h.Dashboard)).Methods("GET")
	r.HandleFunc("/admin-dashboard", h.RequireAdmin(h.AdminDashboard)).Methods("GET")
	r.HandleFunc("/add-product", h.RequireAdmin(h.AddProduct)).Methods("GET", "POST")
	r.HandleFunc("/delete-product", h.RequireAdmin(h.DeleteProduct)).Methods("POST")
	// Cart routes
	r.HandleFunc("/cart", h.RequireAuth(h.ViewCart)).Methods("GET")

	log.Println("Server starting on :8080")
	log.Println("Default admin credentials: username=admin, password=admin123")
	log.Fatal(http.ListenAndServe(":8080", r))
}
