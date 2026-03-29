package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"dailyApp/backend/gen/activity/v1/activityv1connect"
	"dailyApp/backend/internal/db"
	"dailyApp/backend/internal/service"
)

func getDSN() string {
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		return dsn
	}
	
	// Default fallbacks
	user := os.Getenv("NS_MARIADB_USER")
	if user == "" {
		user = os.Getenv("DB_USER")
	}
	if user == "" {
		user = "root"
	}

	pass := os.Getenv("NS_MARIADB_PASSWORD")
	if pass == "" {
		pass = os.Getenv("DB_PASS")
	}
	if pass == "" {
		pass = "root"
	}

	host := os.Getenv("NS_MARIADB_HOSTNAME")
	if host == "" {
		host = os.Getenv("DB_HOST")
	}
	if host == "" {
		host = "127.0.0.1"
	}

	port := os.Getenv("NS_MARIADB_PORT")
	if port == "" {
		port = os.Getenv("DB_PORT")
	}
	if port == "" {
		port = "3306"
	}

	dbName := os.Getenv("NS_MARIADB_DATABASE")
	if dbName == "" {
		dbName = os.Getenv("DB_NAME")
	}
	if dbName == "" {
		dbName = "dailyapp"
	}

	log.Printf("DEBUG ENV INFO -> HOST: '%s', PORT: '%s', DB: '%s', USER: '%s', PASS: (hidden)", host, port, dbName, user)

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, dbName)
}

func main() {
	dsn := getDSN()

	database, err := db.InitDB(dsn)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	mux := http.NewServeMux()
	
	activitySvc := service.NewActivityService(database)
	planSvc := service.NewPlanService(database)
	
	path, handler := activityv1connect.NewActivityServiceHandler(activitySvc)
	mux.Handle(path, handler)
	mux.Handle("/api/v1/plans", planSvc)

	// Wrap with CORS middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			// Important Connect headers
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Expose-Headers", "*")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("Starting backend server on %s", addr)
	err = http.ListenAndServe(
		addr,
		h2c.NewHandler(corsMiddleware(mux), &http2.Server{}),
	)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
