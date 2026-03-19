package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"alphacrawl/internal/api" // Adjust "alphacrawl" to your actual module name

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func initDB() *sql.DB {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://alpha:password@db:5432/crawlerdb?sslmode=disable"
	}
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("❌ DB Connect: %v", err)
	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(100)
	if err = db.Ping(); err != nil {
		log.Fatalf("❌ DB Ping: %v", err)
	}
	fmt.Println("🗄️ Database Connected Successfully!")
	return db
}

func main() {
	db := initDB()
	defer db.Close()

	// Initialize the API handler with dependencies
	h := &api.Handler{DB: db}

	// Use Gin's release mode for production (optional, uncomment for launch)
	// gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Setup Routes
	protected := r.Group("/")
	protected.Use(api.AuthMiddleware())
	{
		protected.GET("/stats", h.GetStats)
		protected.GET("/leads", h.GetLeads)
		protected.POST("/spider", h.PostSpider)
		protected.GET("/task/:task_id", h.GetTask)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 1. Create a custom HTTP server with timeouts!
	srv := &http.Server{
		Addr:         "0.0.0.0:" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 2. Run the server in a goroutine so it doesn't block
	go func() {
		fmt.Printf("🚀 AlphaCrawl API running on port %s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Listen error: %s\n", err)
		}
	}()

	// 3. Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT (Ctrl+C)
	// kill -9 is syscall.SIGKILL (but can't be caught, so don't add it)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\n🛑 Shutting down server gracefully...")

	// 4. Give outstanding requests 10 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("❌ Server forced to shutdown: ", err)
	}

	fmt.Println("✅ AlphaCrawl exited properly.")
}
