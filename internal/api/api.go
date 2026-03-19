package api

import (
	"alphacrawl/internal/crawler"
	"alphacrawl/internal/models"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler holds dependencies for the API
type Handler struct {
	DB *sql.DB
}

// AuthMiddleware secures the endpoints
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		secret := os.Getenv("API_SECRET_KEY")
		if secret == "" {
			secret = "AlphaDev_Super_Secret_2026"
		}
		if apiKey != secret {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (h *Handler) GetStats(c *gin.Context) {
	var s models.GlobalStats
	err := h.DB.QueryRow(`SELECT COUNT(*), COUNT(*) FILTER (WHERE status='completed'), COUNT(*) FILTER (WHERE status='processing'), 
                 COALESCE(SUM(page_count), 0), COALESCE(MIN(started_at)::text, ''), COALESCE(MAX(completed_at)::text, ''),
                 COALESCE((MAX(completed_at) - MIN(started_at))::text, '0'), COALESCE(AVG(duration_seconds), 0) FROM spider_tasks`).Scan(
		&s.TotalTasks, &s.CompletedTasks, &s.ProcessingTasks, &s.TotalPagesScraped, &s.TestStart, &s.TestEnd, &s.TotalLoadDuration, &s.AvgSecondsPerSite)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) GetLeads(c *gin.Context) {
	rows, err := h.DB.Query(`SELECT website_url, status, error_code, COALESCE(error_message, ''), last_crawled_at FROM website_leads ORDER BY last_crawled_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close() // Safely deferred AFTER error check

	var results []interface{}
	for rows.Next() {
		var url, status, err_msg, time_str string
		var code sql.NullInt32
		rows.Scan(&url, &status, &code, &err_msg, &time_str)
		results = append(results, gin.H{"url": url, "status": status, "code": code.Int32, "error": err_msg, "time": time_str})
	}
	c.JSON(http.StatusOK, results)
}

func (h *Handler) PostSpider(c *gin.Context) {
	var req models.SpiderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsedURL, err := url.Parse(strings.TrimSpace(req.URL))
	if err != nil || parsedURL.Host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL provided"})
		return
	}
	taskID := uuid.New().String()
	h.DB.Exec(`INSERT INTO spider_tasks (task_id, status, started_at) VALUES ($1, $2, $3)
             ON CONFLICT (task_id) DO UPDATE SET status='processing', started_at=EXCLUDED.started_at`,
		taskID, "processing", time.Now())

	h.DB.Exec(`INSERT INTO website_leads (task_id, website_url, status) VALUES ($1, $2, $3)`,
		taskID, parsedURL.String(), "processing")

	// Launch crawler in background
	go crawler.StartCrawl(h.DB, taskID, req)

	c.JSON(http.StatusAccepted, gin.H{"task_id": taskID, "status": "processing"})
}

func (h *Handler) GetTask(c *gin.Context) {
	taskID := c.Param("task_id")

	// 🌟 FIX 1: Bring back pagination to protect your server's RAM
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var status string
	var total int
	if err := h.DB.QueryRow("SELECT status FROM spider_tasks WHERE task_id=$1", taskID).Scan(&status); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	h.DB.QueryRow("SELECT COUNT(*) FROM crawled_pages WHERE task_id=$1", taskID).Scan(&total)

	// 🌟 FIX 1 (Cont.): Add LIMIT and OFFSET to the query
	rows, err := h.DB.Query(`SELECT url, status, markdown_content, extracted_content, html_content, metadata, error_message 
                         FROM crawled_pages WHERE task_id=$1 ORDER BY id ASC LIMIT $2 OFFSET $3`, taskID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task data"})
		return
	}
	defer rows.Close()

	var results []models.PageResult
	for rows.Next() {
		var pr models.PageResult
		var metaBytes []byte
		var pageStatus string
		var errStr sql.NullString

		// 🌟 FIX 3: Catch scan errors
		scanErr := rows.Scan(
			&pr.URL,
			&pageStatus,
			&pr.Markdown,
			&pr.ExtractedContent,
			&pr.HTML,
			&metaBytes,
			&errStr,
		)

		if scanErr != nil {
			continue // Skip corrupted rows instead of crashing or returning blank data
		}

		pr.Success = (pageStatus == "success")
		if errStr.Valid {
			pr.Error = &errStr.String
		}

		// Unmarshal the metadata object from DB into the struct
		json.Unmarshal(metaBytes, &pr.Metadata)

		results = append(results, pr)
	}

	c.JSON(http.StatusOK, models.TaskResponse{
		TaskID:  taskID,
		Status:  status,
		Results: results, // Note: You might want to add TotalCount to this struct later so the UI knows how many pages exist!
	})
}
