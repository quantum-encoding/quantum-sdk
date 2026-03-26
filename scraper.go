package qai

import "context"

// ---------------------------------------------------------------------------
// Scrape
// ---------------------------------------------------------------------------

// ScrapeTarget is a single scrape target.
type ScrapeTarget struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Type      string `json:"type,omitempty"`
	Selector  string `json:"selector,omitempty"`
	Content   string `json:"content,omitempty"`
	Notebook  string `json:"notebook,omitempty"`
	Recursive bool   `json:"recursive,omitempty"`
	MaxPages  int    `json:"max_pages,omitempty"`
	Delay     int    `json:"delay_ms,omitempty"`
	Ingest    string `json:"ingest,omitempty"`
}

// ScrapeRequest is the request body for submitting a scrape job.
type ScrapeRequest struct {
	Targets []ScrapeTarget `json:"targets"`
}

// ScrapeResponse is the response from submitting a scrape job.
type ScrapeResponse struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	Targets   int    `json:"targets"`
	RequestID string `json:"request_id"`
}

// ---------------------------------------------------------------------------
// Screenshot
// ---------------------------------------------------------------------------

// ScreenshotURL is a single URL to screenshot.
type ScreenshotURL struct {
	URL      string `json:"url"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	FullPage bool   `json:"full_page,omitempty"`
	Delay    int    `json:"delay_ms,omitempty"`
}

// ScreenshotRequest is the request body for taking screenshots.
type ScreenshotRequest struct {
	URLs []ScreenshotURL `json:"urls"`
}

// ScreenshotResult is a single screenshot result.
type ScreenshotResult struct {
	URL    string `json:"url"`
	Base64 string `json:"base64"`
	Format string `json:"format"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Error  string `json:"error,omitempty"`
}

// ScreenshotResponse is the response from the screenshot endpoint.
type ScreenshotResponse struct {
	Screenshots []ScreenshotResult `json:"screenshots"`
	Count       int                `json:"count"`
}

// ScreenshotJobResponse is the response from async screenshot job submission.
type ScreenshotJobResponse struct {
	JobID     string `json:"job_id"`
	Status    string `json:"status"`
	URLs      int    `json:"urls"`
	RequestID string `json:"request_id"`
}

// ---------------------------------------------------------------------------
// Client methods
// ---------------------------------------------------------------------------

// Scrape submits a doc-scraping job. Returns a job ID for polling.
func (c *Client) Scrape(ctx context.Context, req *ScrapeRequest) (*ScrapeResponse, error) {
	var resp ScrapeResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/scraper/scrape", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Screenshot takes screenshots of URLs. For <=5 URLs, returns results inline.
// For >5, returns a job response for async processing.
func (c *Client) Screenshot(ctx context.Context, req *ScreenshotRequest) (*ScreenshotResponse, error) {
	var resp ScreenshotResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/scraper/screenshot", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
