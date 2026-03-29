package main

import (
	"log"
	"net/http"
	"os"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"dailyApp/backend/gen/activity/v1/activityv1connect"
	"dailyApp/backend/internal/db"
	"dailyApp/backend/internal/service"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/dailyapp?charset=utf8mb4&parseTime=True&loc=Local"
	}

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

	addr := ":8080"
	log.Printf("Starting backend server on %s", addr)
	err = http.ListenAndServe(
		addr,
		h2c.NewHandler(corsMiddleware(mux), &http2.Server{}),
	)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
