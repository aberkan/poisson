package utils

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateURL validates that a string is a well-formed URL with http or https scheme.
func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsed.Scheme == "" {
		return fmt.Errorf("URL must include a scheme (http:// or https://)")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https, got: %s", parsed.Scheme)
	}

	if parsed.Host == "" {
		return fmt.Errorf("URL must include a host")
	}

	return nil
}

// ValidateRSSURL validates that a string is a well-formed URL suitable for RSS feeds.
// It performs the same validation as ValidateURL but can be extended with RSS-specific checks.
func ValidateRSSURL(urlStr string) error {
	if err := ValidateURL(urlStr); err != nil {
		return fmt.Errorf("invalid RSS feed URL: %w", err)
	}

	// Optional: Check if URL looks like an RSS feed
	// This is a soft check - many RSS feeds don't have .xml or /rss in the path
	urlLower := strings.ToLower(urlStr)
	if !strings.Contains(urlLower, ".xml") &&
		!strings.Contains(urlLower, "/rss") &&
		!strings.Contains(urlLower, "/feed") &&
		!strings.Contains(urlLower, "atom") {
		// This is just a warning, not an error - many valid RSS feeds don't follow this pattern
		// We'll let the RSS parser handle actual validation
	}

	return nil
}

