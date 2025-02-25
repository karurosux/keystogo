package models

// ValidationResult represents the result of validating an API key
type ValidationResult struct {
	Valid  bool    `json:"valid"`
	APIKey *APIKey `json:"apiKey"`
	Error  error   `json:"error"`
}
