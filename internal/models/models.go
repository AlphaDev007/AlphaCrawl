package models

// SpiderRequest represents the incoming JSON payload
type SpiderRequest struct {
	URL        string `json:"url" binding:"required"`
	MaxDepth   int    `json:"depth"`
	Limit      int    `json:"limit"`
	WebhookURL string `json:"webhook_url"`
}

// Metadata represents the nested SEO, Social, and crawl-depth object
type Metadata struct {
	Title              string `json:"title"`
	Description        string `json:"description"`
	Keywords           string `json:"keywords"`
	Author             string `json:"author"`
	OGTitle            string `json:"og:title"`
	OGDescription      string `json:"og:description"`
	OGURL              string `json:"og:url"`
	OGSiteName         string `json:"og:site_name"`
	OGLocale           string `json:"og:locale"` // 🌟 Added to match result.json
	OGImage            string `json:"og:image"`
	OGImageWidth       string `json:"og:image:width"`
	OGImageHeight      string `json:"og:image:height"`
	OGImageAlt         string `json:"og:image:alt"`
	OGType             string `json:"og:type"`
	TwitterCard        string `json:"twitter:card"`
	TwitterTitle       string `json:"twitter:title"`
	TwitterDescription string `json:"twitter:description"`
	TwitterImage       string `json:"twitter:image"`
	Depth              int    `json:"depth"`
	ParentURL          string `json:"parent_url"`
}

// PageResult represents a single scraped page
type PageResult struct {
	URL              string   `json:"url"`
	Success          bool     `json:"success"`
	Markdown         string   `json:"markdown"`          // 🌟 Full Page Markdown String
	HTML             string   `json:"html"`              // Raw Source
	ExtractedContent string   `json:"extracted_content"` // Clean Article Text
	Metadata         Metadata `json:"metadata"`          // 🌟 Nested SEO/Social Object
	Error            *string  `json:"error,omitempty"`
}

// TaskResponse represents the webhook and API response payload
type TaskResponse struct {
	TaskID  string       `json:"task_id"`
	Status  string       `json:"status"`
	Results []PageResult `json:"results"`
}

// GlobalStats represents the high-level system metrics
type GlobalStats struct {
	TotalTasks        int     `json:"total_tasks"`
	CompletedTasks    int     `json:"completed_tasks"`
	ProcessingTasks   int     `json:"processing_tasks"`
	TotalPagesScraped int     `json:"total_pages_scraped"`
	TestStart         string  `json:"test_start"`
	TestEnd           string  `json:"test_end"`
	TotalLoadDuration string  `json:"total_load_test_duration"`
	AvgSecondsPerSite float64 `json:"avg_seconds_per_site"`
}

// FullDataExport represents the data export payload
type FullDataExport struct {
	TaskID   string      `json:"task_id"`
	Status   string      `json:"status"`
	URL      string      `json:"url"`
	Title    string      `json:"title"`
	Markdown string      `json:"markdown"`
	HTML     string      `json:"html"`
	Metadata interface{} `json:"metadata"`
}
