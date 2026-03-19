# 🕷️ AlphaCrawl v1.0

AlphaCrawl is a high-performance, asynchronous web crawler built in Go. It is specifically engineered to power LLM (Large Language Model) data pipelines and RAG (Retrieval-Augmented Generation) applications by extracting noise-free text, full-page Markdown, and deep SEO/Social metadata.

## ✨ Core Features

* **LLM-Ready Extraction:** Dual-layer content parsing. Retrieves the full page structure as Markdown, and uses a Readability engine to extract the clean, boilerplate-free article text.
* **Deep Metadata Parsing:** Automatically extracts 18+ data points per page, including standard SEO tags, OpenGraph (`og:title`, `og:locale`, etc.), and Twitter Cards.
* **Anti-Blocking Architecture:** Built-in Random User-Agent rotation, automatic Referer header injection, and human-like request headers to bypass standard bot protections.
* **Asynchronous Processing:** Built on `gocolly/colly`, supporting concurrent scraping with customizable politeness delays and depth limits.
* **Robust Persistence:** Fully containerized PostgreSQL backend with `JSONB` metadata indexing for lightning-fast retrieval.

## 🚀 Quick Start

### Prerequisites
* Docker
* Docker Compose

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/AlphaDev007/alphacrawl.git
   cd alphacrawl
   ```
2. Start the engine (this spins up the Go API and the PostgreSQL database):
```bash
docker-compose up -d --build
```
The API will be available at http://localhost:8080.

## API Reference
Authentication requires the X-API-KEY header (default: your_secret_key_here).

### 1. Start a Crawl Task 
```POST /spider```

### Payload:

```json
{
  "url": "[https://news.ycombinator.com](https://news.ycombinator.com)",
  "depth": 1,
  "limit": 10,
  "webhook_url": "https://your-webhook.com/callback"
}
```

*Returns an auto-generated ```task_id```*

### 2. Retrieve Task Data
```GET /task/:task_id```

### Sample Response Segment:

```json
{
  "task_id": "e3a1257f-a2ad-4987-8d6a-4924ac4c8678",
  "status": "completed",
  "total_count": 10,
  "results": [
    {
      "url": "[https://example.com/article](https://example.com/article)",
      "success": true,
      "markdown": "# Article Title\n\nFull page content...",
      "html": "<!DOCTYPE html>...",
      "extracted_content": "The core noise-free text of the article goes here...",
      "metadata": {
        "title": "Article Title",
        "og:locale": "en_US",
        "og:image": "[https://example.com/image.png](https://example.com/image.png)",
        "depth": 1,
        "parent_url": "[https://example.com](https://example.com)"
      }
    }
  ]
}
```

### 3. System Metrics
```GET /stats```
Returns system-wide metrics, including total pages scraped, active tasks, and average processing time.

### 🛠️ Tech Stack

* Language: Go 1.21+

* Framework: Gin

* Scraping Engine: Colly & Goquery

* Data Processing: JohannesKaufmann/html-to-markdown, go-shiori/go-readability

* Database: PostgreSQL (pgx driver)

## 📄 License
### MIT License
