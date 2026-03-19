package crawler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"alphacrawl/internal/models"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/go-readability"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

var conv *converter.Converter

func init() {
	conv = converter.NewConverter(
		converter.WithPlugins(base.NewBasePlugin(), commonmark.NewCommonmarkPlugin()),
	)
}

// Helper to extract meta tags by name or property
func getMeta(doc *goquery.Document, key string) string {
	var val string
	// Try property first (common for OpenGraph)
	val, _ = doc.Find(fmt.Sprintf("meta[property='%s']", key)).Attr("content")
	if val == "" {
		// Try name (common for standard SEO/Twitter)
		val, _ = doc.Find(fmt.Sprintf("meta[name='%s']", key)).Attr("content")
	}
	return strings.TrimSpace(val)
}

func StartCrawl(db *sql.DB, taskID string, req models.SpiderRequest) {
	parsedURL, err := url.Parse(strings.TrimSpace(req.URL))
	if err != nil {
		log.Printf("❌ Task %s: Invalid URL %s", taskID, req.URL)
		return
	}

	seedStr := parsedURL.String()
	domain := strings.TrimPrefix(parsedURL.Host, "www.")

	limit := int32(req.Limit)
	if limit <= 0 {
		limit = 10
	}
	maxDepth := req.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 2
	}

	startTime := time.Now()
	var pagesFound int32 = 0

	collector := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(maxDepth),
		colly.AllowedDomains(domain, "www."+domain),
	)

	extensions.RandomUserAgent(collector)
	extensions.Referer(collector)

	collector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.5")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
	})

	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       1 * time.Second,
	})

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if atomic.LoadInt32(&pagesFound) < limit {
			link := e.Request.AbsoluteURL(e.Attr("href"))
			e.Request.Ctx.Put("parent_url", e.Request.URL.String())
			e.Request.Visit(link)
		}
	})

	collector.OnResponse(func(r *colly.Response) {
		if atomic.LoadInt32(&pagesFound) >= limit || !strings.Contains(r.Headers.Get("Content-Type"), "text/html") {
			return
		}

		if r.Request.URL.String() == seedStr {
			db.Exec("UPDATE website_leads SET status='success', error_code=200 WHERE task_id=$1", taskID)
		}

		htmlContent := string(r.Body)
		doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))

		mdMeta := models.Metadata{
			Title:              doc.Find("title").Text(),
			Description:        getMeta(doc, "description"),
			Keywords:           getMeta(doc, "keywords"),
			Author:             getMeta(doc, "author"),
			OGTitle:            getMeta(doc, "og:title"),
			OGDescription:      getMeta(doc, "og:description"),
			OGURL:              getMeta(doc, "og:url"),
			OGSiteName:         getMeta(doc, "og:site_name"),
			OGLocale:           getMeta(doc, "og:locale"), // 🌟 FIX 1: Added OGLocale
			OGImage:            getMeta(doc, "og:image"),
			OGImageWidth:       getMeta(doc, "og:image:width"),
			OGImageHeight:      getMeta(doc, "og:image:height"),
			OGImageAlt:         getMeta(doc, "og:image:alt"),
			OGType:             getMeta(doc, "og:type"),
			TwitterCard:        getMeta(doc, "twitter:card"),
			TwitterTitle:       getMeta(doc, "twitter:title"),
			TwitterDescription: getMeta(doc, "twitter:description"),
			TwitterImage:       getMeta(doc, "twitter:image"),
			Depth:              r.Request.Depth,
			ParentURL:          r.Ctx.Get("parent_url"),
		}

		// 🌟 FIX 2: Generate Full Page Markdown
		md, _ := conv.ConvertString(htmlContent)

		// Generate Extracted Content (Clean Article Text)
		article, err := readability.FromReader(bytes.NewReader(r.Body), r.Request.URL)
		extractedContent := ""
		if err == nil {
			extractedContent = article.TextContent
		}

		metaJSON, _ := json.Marshal(mdMeta)

		// Persistence with proper ON CONFLICT updates
		_, dbErr := db.Exec(`INSERT INTO crawled_pages (task_id, url, parent_url, depth, title, markdown_content, extracted_content, html_content, metadata, status)
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
             ON CONFLICT (url) DO UPDATE SET 
             status='success', 
             html_content=EXCLUDED.html_content,
             markdown_content=EXCLUDED.markdown_content,
             extracted_content=EXCLUDED.extracted_content,
             metadata=EXCLUDED.metadata`,
			taskID, r.Request.URL.String(), mdMeta.ParentURL, mdMeta.Depth, mdMeta.Title,
			md, extractedContent, htmlContent, metaJSON, "success")

		if dbErr == nil {
			atomic.AddInt32(&pagesFound, 1)
		}
	})

	collector.OnError(func(r *colly.Response, err error) {
		errMsg := err.Error()
		if r.Request.URL.String() == seedStr {
			db.Exec("UPDATE website_leads SET status='failed', error_code=$1, error_message=$2 WHERE task_id=$3", r.StatusCode, errMsg, taskID)
		}
		db.Exec(`INSERT INTO crawled_pages (task_id, url, parent_url, depth, status, error_message)
                 VALUES ($1, $2, $3, $4, $5, $6)
                 ON CONFLICT (url) DO UPDATE SET status='failed', error_message=EXCLUDED.error_message, crawled_at=CURRENT_TIMESTAMP 
                 WHERE crawled_pages.status != 'success'`,
			taskID, r.Request.URL.String(), r.Request.Ctx.Get("parent_url"), r.Request.Depth, "failed", errMsg)
	})

	// 🌟 FIX 3: Assign the error correctly
	err = collector.Visit(seedStr)
	if err != nil {
		log.Printf("❌ Task %s: Failed to start collector: %v", taskID, err)
		return
	}

	collector.Wait()

	finalCount := int(atomic.LoadInt32(&pagesFound))
	duration := time.Since(startTime).Seconds()

	db.Exec("UPDATE spider_tasks SET status='completed', completed_at=CURRENT_TIMESTAMP, duration_seconds=$1, page_count=$2 WHERE task_id=$3", duration, finalCount, taskID)

	if req.WebhookURL != "" {
		sendWebhook(db, taskID, req.WebhookURL)
	}
}

func sendWebhook(db *sql.DB, taskID, webhookURL string) {
	var status string
	var total int
	db.QueryRow("SELECT status FROM spider_tasks WHERE task_id=$1", taskID).Scan(&status)
	db.QueryRow("SELECT COUNT(*) FROM crawled_pages WHERE task_id=$1", taskID).Scan(&total)

	// Sending a lightweight webhook payload to prevent OOM
	payload := map[string]interface{}{
		"task_id":     taskID,
		"status":      status,
		"total_count": total,
		"message":     "Crawl complete. Use GET /task/" + taskID + " to retrieve data.",
	}

	jsonData, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		fmt.Printf("✅ Webhook delivered for task %s | Status: %d\n", taskID, resp.StatusCode)
	} else {
		fmt.Printf("❌ Webhook failed for task %s: %v\n", taskID, err)
	}
}
