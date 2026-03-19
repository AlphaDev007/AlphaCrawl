# 🕷️ AlphaCrawl (by AlphaDev)

> **A high-performance, asynchronous web crawler built in Go, designed specifically for extracting clean, LLM-ready Markdown.**

AlphaCrawl isn't just a scraper; it's an intelligent data pipeline. It traverses websites, strips away intrusive UI elements (ads, navbars, footers) using Readability algorithms, and converts the core content into perfectly formatted Markdown—ready to be ingested by your Retrieval-Augmented Generation (RAG) pipelines or LLM training sets.

## ✨ Features

- **LLM-Ready Markdown:** Automatically converts messy HTML into clean Markdown.
- **Smart Content Extraction:** Identifies the "main article" and ignores boilerplate UI.
- **Asynchronous & Fast:** Powered by Colly for highly concurrent, depth-controlled scraping.
- **Webhook Notifications:** Receive real-time payloads the moment a crawl job finishes.
- **API-First & Secure:** Simple REST API secured via API keys.

## 🚀 Quick Start (Docker)

Get AlphaCrawl and its PostgreSQL database running in seconds:

1. Clone the repo:
   ```bash
   git clone [https://github.com/AlphaDev007/alphacrawl.git](https://github.com/AlphaDev007/alphacrawl.git)
   cd alphacrawl
   ```
2. Start the stack:
   ```bash
   docker-compose up -d
   ```
3. Test the API:
   ```bash
   curl -X POST http://localhost:8080/spider \
    -H "X-API-KEY: AlphaDev_Super_Secret_2026" \
    -H "Content-Type: application/json" \
    -d '{"task_id": "job_123", "url": "[https://example.com](https://example.com)", "depth": 1, "limit": 5}'
   ```
