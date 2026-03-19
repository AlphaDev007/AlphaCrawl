-- 1. Track the overall crawl jobs
CREATE TABLE IF NOT EXISTS spider_tasks (
    task_id VARCHAR(255) PRIMARY KEY,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_seconds DOUBLE PRECISION DEFAULT 0.0,
    page_count INTEGER DEFAULT 0
);

-- 2. Track the initial entry points (The "Leads")
CREATE TABLE IF NOT EXISTS website_leads (
    id SERIAL PRIMARY KEY,
    task_id VARCHAR(255) REFERENCES spider_tasks(task_id) ON DELETE CASCADE,
    website_url TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_code INTEGER,
    error_message TEXT,
    last_crawled_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Store the actual scraped data
CREATE TABLE IF NOT EXISTS crawled_pages (
    id SERIAL PRIMARY KEY,
    task_id VARCHAR(255) REFERENCES spider_tasks(task_id) ON DELETE CASCADE,
    url TEXT NOT NULL UNIQUE,
    parent_url TEXT,
    depth INTEGER DEFAULT 0,
    title TEXT,
    markdown_content TEXT,    -- Scanned into the "markdown" JSON key
    extracted_content TEXT,   -- Scanned into the "extracted_content" JSON key
    html_content TEXT,        -- Scanned into the "html" JSON key
    metadata JSONB,           -- Scanned into the "metadata" JSON object
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    crawled_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 4. Performance Indexes
CREATE INDEX IF NOT EXISTS idx_crawled_pages_task_id ON crawled_pages(task_id);
CREATE INDEX IF NOT EXISTS idx_metadata_gin ON crawled_pages USING GIN (metadata);