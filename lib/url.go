package lib

import (
	"strings"
)

// NormalizeURL normalizes a URL by removing the protocol (http:// or https://) and query parameters.
func NormalizeURL(url string) string {
	// Remove http:// or https:// from the front
	normalized := strings.TrimPrefix(url, "https://")
	normalized = strings.TrimPrefix(normalized, "http://")

	// Remove query parameters (everything after ?)
	if idx := strings.Index(normalized, "?"); idx != -1 {
		normalized = normalized[:idx]
	}

	return normalized
}

// AddProtocol adds https:// to a normalized URL if it doesn't already have a protocol.
// This is used when making HTTP requests with normalized URLs.
// For localhost and 127.0.0.1, it uses http:// instead of https://.
func AddProtocol(normalizedURL string) string {
	if strings.HasPrefix(normalizedURL, "http://") || strings.HasPrefix(normalizedURL, "https://") {
		return normalizedURL
	}
	// Use http:// for localhost and 127.0.0.1, https:// for everything else
	if strings.HasPrefix(normalizedURL, "localhost") || strings.HasPrefix(normalizedURL, "127.0.0.1") {
		return "http://" + normalizedURL
	}
	return "https://" + normalizedURL
}
