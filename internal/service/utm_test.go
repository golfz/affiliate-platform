package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// UTMTestSuite is the test suite for UTM utility functions
type UTMTestSuite struct {
	suite.Suite
}

func (suite *UTMTestSuite) SetupTest() {
	// No setup needed for utility functions
}

func (suite *UTMTestSuite) TearDownTest() {
	// No teardown needed
}

// TestBuildTargetURL tests the buildTargetURL function
func (suite *UTMTestSuite) TestBuildTargetURL() {
	tests := []struct {
		name         string
		baseURL      string
		utmCampaign  string
		utmSource    string
		utmMedium    string
		wantErr      bool
		errContains  string
		wantContains []string // URLs should contain these query params
	}{
		{
			name:         "success with simple URL",
			baseURL:      "https://example.com/product",
			utmCampaign:  "summer_2025",
			utmSource:    "affiliate",
			utmMedium:    "affiliate",
			wantErr:      false,
			wantContains: []string{"utm_source=affiliate", "utm_medium=affiliate", "utm_campaign=summer_2025"},
		},
		{
			name:         "success with URL that already has query params",
			baseURL:      "https://example.com/product?id=123",
			utmCampaign:  "summer_2025",
			utmSource:    "affiliate",
			utmMedium:    "affiliate",
			wantErr:      false,
			wantContains: []string{"utm_source=affiliate", "utm_medium=affiliate", "utm_campaign=summer_2025", "id=123"},
		},
		{
			name:         "success with special characters in UTM campaign",
			baseURL:      "https://example.com/product",
			utmCampaign:  "summer_deal_2025",
			utmSource:    "affiliate",
			utmMedium:    "affiliate",
			wantErr:      false,
			wantContains: []string{"utm_campaign=summer_deal_2025"},
		},
		{
			name:        "error with invalid URL",
			baseURL:     "://invalid-url",
			utmCampaign: "summer_2025",
			utmSource:   "affiliate",
			utmMedium:   "affiliate",
			wantErr:     true,
			errContains: "invalid base URL",
		},
		{
			name:         "success with empty UTM values",
			baseURL:      "https://example.com/product",
			utmCampaign:  "",
			utmSource:    "",
			utmMedium:    "",
			wantErr:      false,
			wantContains: []string{"utm_source=", "utm_medium=", "utm_campaign="},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result, err := buildTargetURL(tt.baseURL, tt.utmCampaign, tt.utmSource, tt.utmMedium)

			if tt.wantErr {
				assert.Error(suite.T(), err)
				if tt.errContains != "" {
					assert.Contains(suite.T(), err.Error(), tt.errContains)
				}
				assert.Empty(suite.T(), result)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotEmpty(suite.T(), result)
				for _, want := range tt.wantContains {
					assert.Contains(suite.T(), result, want)
				}
			}
		})
	}
}

func TestUTMTestSuite(t *testing.T) {
	suite.Run(t, new(UTMTestSuite))
}
