package qai

import (
	"context"
	"fmt"
	"net/url"
)

// ---------------------------------------------------------------------------
// Scan requests
// ---------------------------------------------------------------------------

// SecurityScanURLRequest is the request body for scanning a URL for prompt injection.
type SecurityScanURLRequest struct {
	// URL to scan.
	URL string `json:"url"`
}

// SecurityScanHTMLRequest is the request body for scanning raw HTML content.
type SecurityScanHTMLRequest struct {
	// Raw HTML to scan.
	HTML string `json:"html"`

	// Rendered visible text (for structural analysis).
	VisibleText string `json:"visible_text,omitempty"`

	// Source URL (for context).
	URL string `json:"url,omitempty"`
}

// SecurityReportRequest is the request body for reporting a suspicious URL.
type SecurityReportRequest struct {
	// URL to report.
	URL string `json:"url"`

	// Description of the suspected threat.
	Description string `json:"description,omitempty"`

	// Category of the suspected threat.
	Category string `json:"category,omitempty"`
}

// ---------------------------------------------------------------------------
// Response types
// ---------------------------------------------------------------------------

// SecurityScanResponse is the response from a security scan.
type SecurityScanResponse struct {
	// Full threat assessment.
	Assessment SecurityAssessment `json:"assessment"`

	// Request identifier.
	RequestID string `json:"request_id"`
}

// SecurityAssessment is the threat assessment for a scanned page.
type SecurityAssessment struct {
	// Source URL.
	URL string `json:"url"`

	// Overall threat level: "none", "low", "medium", "high", "critical".
	ThreatLevel string `json:"threat_level"`

	// Numeric threat score (0.0 - 100.0).
	ThreatScore float64 `json:"threat_score"`

	// Individual findings.
	Findings []SecurityFinding `json:"findings"`

	// Length of hidden text content detected.
	HiddenTextLength int32 `json:"hidden_text_length"`

	// Length of visible text content.
	VisibleTextLength int32 `json:"visible_text_length"`

	// Ratio of hidden to total content.
	HiddenRatio float64 `json:"hidden_ratio"`

	// Human-readable summary.
	Summary string `json:"summary"`
}

// SecurityFinding is a single detected injection pattern.
type SecurityFinding struct {
	// Category: "instruction_override", "role_impersonation", "data_exfiltration",
	// "hidden_text", "hidden_comment", "model_targeting", "encoded_payload",
	// "structural_anomaly", "meta_injection", "safety_override".
	Category string `json:"category"`

	// Pattern that matched.
	Pattern string `json:"pattern"`

	// Offending content (truncated).
	Content string `json:"content"`

	// Location in the page.
	Location string `json:"location"`

	// Threat level for this finding.
	Threat string `json:"threat"`

	// Detection confidence (0.0 - 1.0).
	Confidence float64 `json:"confidence"`

	// Human-readable description.
	Description string `json:"description"`
}

// SecurityCheckResponse is the response from checking a URL against the registry.
type SecurityCheckResponse struct {
	// URL that was checked.
	URL string `json:"url"`

	// Whether the URL is blocked.
	Blocked bool `json:"blocked"`

	// Threat level (if blocked).
	ThreatLevel string `json:"threat_level,omitempty"`

	// Threat score (if blocked).
	ThreatScore *float64 `json:"threat_score,omitempty"`

	// Detection categories (if blocked).
	Categories []string `json:"categories,omitempty"`

	// First seen timestamp.
	FirstSeen string `json:"first_seen,omitempty"`

	// Last seen timestamp.
	LastSeen string `json:"last_seen,omitempty"`

	// Number of reports.
	ReportCount *int32 `json:"report_count,omitempty"`

	// Registry status: "confirmed", "suspected".
	Status string `json:"status,omitempty"`

	// Human-readable message.
	Message string `json:"message,omitempty"`
}

// SecurityBlocklistResponse is the response from the blocklist feed.
type SecurityBlocklistResponse struct {
	// Blocklist entries.
	Entries []SecurityBlocklistEntry `json:"entries"`

	// Total count.
	Count int32 `json:"count"`

	// Filter status used.
	Status string `json:"status"`
}

// SecurityBlocklistEntry is a single blocklist entry.
type SecurityBlocklistEntry struct {
	// Entry identifier.
	ID string `json:"id,omitempty"`

	// Blocked URL.
	URL string `json:"url"`

	// Registry status.
	Status string `json:"status"`

	// Threat level.
	ThreatLevel string `json:"threat_level"`

	// Threat score.
	ThreatScore float64 `json:"threat_score"`

	// Detection categories.
	Categories []string `json:"categories"`

	// Number of findings.
	FindingsCount int32 `json:"findings_count"`

	// Hidden content ratio.
	HiddenRatio float64 `json:"hidden_ratio"`

	// First seen timestamp.
	FirstSeen string `json:"first_seen,omitempty"`

	// Summary.
	Summary string `json:"summary"`
}

// SecurityReportResponse is the response from reporting a URL.
type SecurityReportResponse struct {
	// URL that was reported.
	URL string `json:"url"`

	// Report status: "existing" or "suspected".
	Status string `json:"status"`

	// Message.
	Message string `json:"message"`

	// Threat level (if already in registry).
	ThreatLevel string `json:"threat_level,omitempty"`
}

// ---------------------------------------------------------------------------
// Client methods
// ---------------------------------------------------------------------------

// SecurityScanURL scans a URL for prompt injection attacks.
func (c *Client) SecurityScanURL(ctx context.Context, scanURL string) (*SecurityScanResponse, error) {
	req := &SecurityScanURLRequest{URL: scanURL}
	var resp SecurityScanResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/security/scan-url", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SecurityScanHTML scans raw HTML content for prompt injection.
func (c *Client) SecurityScanHTML(ctx context.Context, req *SecurityScanHTMLRequest) (*SecurityScanResponse, error) {
	var resp SecurityScanResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/security/scan-html", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SecurityCheck checks a URL against the injection registry.
func (c *Client) SecurityCheck(ctx context.Context, checkURL string) (*SecurityCheckResponse, error) {
	path := fmt.Sprintf("/qai/v1/security/check?url=%s", url.QueryEscape(checkURL))
	var resp SecurityCheckResponse
	_, err := c.doJSON(ctx, "GET", path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SecurityBlocklist returns the injection blocklist feed.
// Pass an empty string for status to get all entries.
func (c *Client) SecurityBlocklist(ctx context.Context, status string) (*SecurityBlocklistResponse, error) {
	path := "/qai/v1/security/blocklist"
	if status != "" {
		path = fmt.Sprintf("/qai/v1/security/blocklist?status=%s", url.QueryEscape(status))
	}
	var resp SecurityBlocklistResponse
	_, err := c.doJSON(ctx, "GET", path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// SecurityReport reports a suspicious URL.
func (c *Client) SecurityReport(ctx context.Context, req *SecurityReportRequest) (*SecurityReportResponse, error) {
	var resp SecurityReportResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/security/report", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
