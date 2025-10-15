package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	v2 "github.com/ca-gip/artifactory-operator/internal/types/v2"
	"github.com/ca-gip/artifactory-operator/internal/utils"
	"github.com/rs/zerolog"
)

// ExternalAPIService handles communication with the external API
type ExternalAPIService struct {
	logger   zerolog.Logger
	endpoint string
	token    string
	client   *http.Client
}

// NewExternalAPIService creates a new ExternalAPIService
func NewExternalAPIService(endpoint string) *ExternalAPIService {
	logger := utils.Log.With().Str("service", "externalapi").Logger()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	// Get token from environment variable
	token := os.Getenv("ARTI_OP_EXTERNAL_API_TOKEN")
	if token == "" {
		logger.Warn().Msg("External API token not found in environment variables")
	}

	return &ExternalAPIService{
		logger:   logger,
		endpoint: endpoint,
		token:    token,
		client:   client,
	}
}

// CallExternalAPI makes a call to the external API with tenant and project values
func (s *ExternalAPIService) CallExternalAPI(tenant, project string) error {
	s.logger.Info().Msgf("Calling external API for tenant: %s, project: %s", tenant, project)

	// Create request payload
	requestData := v2.ExternalAPIRequest{
		Tenant:  tenant,
		Project: project,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		s.logger.Error().Msgf("Failed to marshal request data: %v", err)
		return err
	}

	// Create request
	req, err := http.NewRequest("POST", s.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		s.logger.Error().Msgf("Failed to create request: %v", err)
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Add authorization token if available
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
		s.logger.Debug().Msg("Added authorization token to request")
	}

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error().Msgf("Failed to send request: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("API call failed with status: %d", resp.StatusCode)
		s.logger.Error().Msgf("%v", err)
		return err
	}

	// Parse response
	var apiResponse v2.ExternalAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		s.logger.Error().Msgf("Failed to decode response: %v", err)
		return err
	}

	// Check if the API call was successful
	if !apiResponse.Success {
		err := fmt.Errorf("API call returned error: %s", apiResponse.Message)
		s.logger.Error().Msgf("%v", err)
		return err
	}

	s.logger.Info().Msgf("External API call successful: %s", apiResponse.Message)
	return nil
}
