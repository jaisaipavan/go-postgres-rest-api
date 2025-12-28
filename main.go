package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

var db *sql.DB

type Product struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Available bool    `json:"available"`
}

func main() {

	// 1️ Database connection
	connStr := "host=localhost port=5432 user=postgres password=562007 dbname=postgres sslmode=disable"

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); 
	err != nil {
		log.Fatal(err)
	}

	fmt.Println(" Connected to database")

	// 2️ Create table
	createProductTable(db)

	// 3️ Insert one product (only for demo)
	product := Product{Name: "Book", Price: 15.55, Available: true}
	insertProduct(db, product)

	// 4️ REST API routes
	http.HandleFunc("/product", getProductHandler)

	fmt.Println(" Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createProductTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS product (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		price NUMERIC(6,2) NOT NULL,
		available BOOLEAN,
		created TIMESTAMP DEFAULT NOW()
	)`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func insertProduct(db *sql.DB, product Product) {
	query := `INSERT INTO product (name, price, available)
	          VALUES ($1, $2, $3)`

	_, err := db.Exec(query, product.Name, product.Price, product.Available)
	if err != nil {
		log.Fatal(err)
	}
}

func getProductHandler(w http.ResponseWriter, r *http.Request) {

	// Read id from URL: /product?id=1
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var product Product

	query := `SELECT id, name, price, available
	          FROM product
	          WHERE id = $1`

	err = db.QueryRow(query, id).
		Scan(&product.ID, &product.Name, &product.Price, &product.Available)

	if err == sql.ErrNoRows {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}
