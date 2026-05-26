package onboarding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"meshsync/internal/models"
	"net/http"
)

// StartClient sends onboarding request to bootstrap node.
func (s *OnboardingService) StartClient(
	host string,
	port int,
	req *models.OnboardingRequest,
) (*models.OnboardingResponse, error) {

	url := fmt.Sprintf(
		"http://%s:%d/api/v1/onboarding/join",
		host,
		port,
	)

	payload, err := json.Marshal(req)

	if err != nil {
		return nil, err
	}

	response, err := http.Post(
		url,
		"application/json",
		bytes.NewBuffer(payload),
	)

	if err != nil {

		log.Printf(
			"[ONBOARDING] failed sending onboarding request: %v",
			err,
		)

		return nil, err
	}

	defer response.Body.Close()

	var resp models.OnboardingResponse

	err = json.NewDecoder(response.Body).Decode(&resp)

	if err != nil {
		return nil, err
	}

	log.Printf(
		"[ONBOARDING] onboarding response status=%s",
		resp.Status,
	)

	return &resp, nil
}
