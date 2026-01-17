package service

import (
	"fmt"
	"net/url"
)

// buildTargetURL builds a target URL with UTM parameters
func buildTargetURL(baseURL, utmCampaign, utmSource, utmMedium string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	q := u.Query()
	q.Set("utm_source", utmSource)
	q.Set("utm_medium", utmMedium)
	q.Set("utm_campaign", utmCampaign)
	u.RawQuery = q.Encode()

	return u.String(), nil
}
