package discovery

import "meshsync/internal/models"

type DiscoveryService struct {
	Org         *models.Org
	JoinRequest *models.JoinRequest
}

func NewDiscoveryService(
	org *models.Org,
	joinRequest *models.JoinRequest,
) *DiscoveryService {

	return &DiscoveryService{
		Org:         org,
		JoinRequest: joinRequest,
	}
}
