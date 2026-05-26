package onboarding

import (
	"time"
)

type OnboardingService struct {
	OrgName     string
	Passcode    string
	TTLPasscode time.Time
}

func NewOnboardingService(
	orgName string,
	passcode string,
	ttl time.Time,
) *OnboardingService {

	return &OnboardingService{
		OrgName:     orgName,
		Passcode:    passcode,
		TTLPasscode: ttl,
	}
}
