package onboarding

import (
	"context"
	"encoding/json"
	"log"
	"meshsync/internal/models"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (s *OnboardingService) StartServer(
	ctx context.Context,
	events chan<- *models.OnboardingEvent,
) error {

	mux := http.NewServeMux()

	mux.HandleFunc(
		"/api/v1/onboarding/join",
		func(w http.ResponseWriter, r *http.Request) {

			if r.Method != http.MethodPost {

				http.Error(
					w,
					"method not allowed",
					http.StatusMethodNotAllowed,
				)

				return
			}

			var req models.OnboardingRequest

			err := json.NewDecoder(r.Body).Decode(&req)

			if err != nil {

				http.Error(
					w,
					"invalid request body",
					http.StatusBadRequest,
				)

				return
			}

			log.Printf(
				"[ONBOARDING] join request org=%s device=%s",
				req.OrgName,
				req.DeviceName,
			)

			// Validate org name and passcode
			if req.OrgName != s.OrgName ||
				req.Passcode != s.Passcode {

				response := models.OnboardingResponse{
					Status:  "error",
					Message: "invalid org name or passcode",
				}

				w.WriteHeader(http.StatusUnauthorized)

				json.NewEncoder(w).Encode(response)

				return
			}

			// Validate passcode TTL
			if time.Now().After(s.TTLPasscode) {

				response := models.OnboardingResponse{
					Status:  "error",
					Message: "passcode expired",
				}

				w.WriteHeader(http.StatusUnauthorized)

				json.NewEncoder(w).Encode(response)

				return
			}

			// Generate cluster identities
			nodeID := uuid.NewString()

			sessionID := uuid.NewString()

			// Create peer descriptor
			peer := models.Peer{
				DeviceID:  nodeID,
				SessionID: sessionID,

				DeviceName: req.DeviceName,
				DeviceIP:   req.DeviceIP,

				ControlPort:  req.ControlPort,
				TransferPort: req.TransferPort,

				JoinedAt: time.Now(),
				LastSeen: time.Now(),

				Status: models.PeerStatusOnline,
			}

			// Build onboarding response
			response := models.OnboardingResponse{
				Status:  "success",
				Message: "onboarding successful",

				NodeID:    nodeID,
				SessionID: sessionID,

				HeartbeatInterval: 30,
			}

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			json.NewEncoder(w).Encode(response)

			// Emit onboarding event
			events <- &models.OnboardingEvent{
				OrgName:   req.OrgName,
				Onboarded: true,
				Peer:      peer,
			}

			log.Printf(
				"[ONBOARDING] peer onboarded device=%s session=%s",
				req.DeviceName,
				sessionID,
			)
		},
	)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {

		<-ctx.Done()

		log.Println(
			"[ONBOARDING] shutting down onboarding server",
		)

		server.Shutdown(context.Background())
	}()

	log.Println(
		"[ONBOARDING] HTTP onboarding server listening on :8080",
	)

	return server.ListenAndServe()
}
