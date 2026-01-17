package validator

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/jonosize/affiliate-platform/pkg/adapters"
)

var (
	AllowedRedirectDomains = []string{
		"lazada.co.th",
		"www.lazada.co.th",
		"shopee.co.th",
		"www.shopee.co.th",
	}
)

// ValidateProductURL validates if a URL is from allowed marketplace domains
func ValidateProductURL(rawURL string) (adapters.Marketplace, adapters.SourceType, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL format: %w", err)
	}

	hostname := strings.ToLower(u.Hostname())

	// Check if hostname matches allowed domains
	for _, domain := range AllowedRedirectDomains {
		if hostname == domain {
			// Determine marketplace
			if strings.Contains(hostname, "lazada") {
				return adapters.MarketplaceLazada, adapters.SourceTypeURL, nil
			}
			if strings.Contains(hostname, "shopee") {
				return adapters.MarketplaceShopee, adapters.SourceTypeURL, nil
			}
		}
	}

	return "", "", fmt.Errorf("URL must be from lazada.co.th or shopee.co.th")
}

// ValidateRedirectURL validates if a redirect URL is from allowed domains
func ValidateRedirectURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	hostname := strings.ToLower(u.Hostname())
	for _, domain := range AllowedRedirectDomains {
		if hostname == domain {
			return true
		}
	}

	return false
}
