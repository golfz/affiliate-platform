package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ShortCodeTestSuite is the test suite for short code generation
type ShortCodeTestSuite struct {
	suite.Suite
}

func (suite *ShortCodeTestSuite) SetupTest() {
	// No setup needed for utility functions
}

func (suite *ShortCodeTestSuite) TearDownTest() {
	// No teardown needed
}

// TestGenerateShortCode tests the generateShortCode function
func (suite *ShortCodeTestSuite) TestGenerateShortCode() {
	tests := []struct {
		name        string
		setupMock   func()
		wantErr     bool
		errContains string
		validate    func(code string) bool // Custom validation function
	}{
		{
			name:      "success - generates valid short code",
			setupMock: func() {},
			wantErr:   false,
			validate: func(code string) bool {
				// Check length is between 8-12
				if len(code) < 8 || len(code) > 12 {
					return false
				}
				// Check all characters are alphanumeric
				for _, char := range code {
					if !((char >= 'A' && char <= 'Z') ||
						(char >= 'a' && char <= 'z') ||
						(char >= '0' && char <= '9')) {
						return false
					}
				}
				return true
			},
		},
		{
			name:      "success - generates multiple unique codes",
			setupMock: func() {},
			wantErr:   false,
			validate: func(code string) bool {
				// Just check it's valid format
				return len(code) >= 8 && len(code) <= 12
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tt.setupMock()

			// Generate multiple codes to test uniqueness
			codes := make(map[string]bool)
			for i := 0; i < 10; i++ {
				code, err := generateShortCode()

				if tt.wantErr {
					assert.Error(suite.T(), err)
					if tt.errContains != "" {
						assert.Contains(suite.T(), err.Error(), tt.errContains)
					}
					assert.Empty(suite.T(), code)
				} else {
					assert.NoError(suite.T(), err)
					assert.NotEmpty(suite.T(), code)

					if tt.validate != nil {
						assert.True(suite.T(), tt.validate(code), "Generated code '%s' failed validation", code)
					}

					// Check uniqueness (for multiple generation test)
					if tt.name == "success - generates multiple unique codes" {
						assert.False(suite.T(), codes[code], "Duplicate code generated: %s", code)
						codes[code] = true
					}
				}
			}
		})
	}
}

func TestShortCodeTestSuite(t *testing.T) {
	suite.Run(t, new(ShortCodeTestSuite))
}
