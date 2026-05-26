package discovery

import (
	"context"
	"encoding/json"
	"log"
	"meshsync/internal/models"
	"net"
)

// StartListener listens for incoming UDP packets on port 9999 and unmarshals them as Advertisement packets.
// It discards any packets that are not valid Advertisement structs.
// The listener runs in an infinite loop until the context is cancelled.
func (s *DiscoveryService) StartListener(ctx context.Context, events chan<- *models.DiscoveryEvent) *models.DiscoveryEvent {
	addr := net.UDPAddr{
		Port: 9999,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Printf("[DISCOVERY] Failed to start listener: %v\n", err)
		return &models.DiscoveryEvent{}
	}
	defer conn.Close()

	log.Println("[DISCOVERY] listener started on port 9999")

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	buffer := make([]byte, 4096)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("[DISCOVERY] listener shutting down: %v\n", err)
			return &models.DiscoveryEvent{}
		}

		var advertisement Advertisement
		err = json.Unmarshal(buffer[:n], &advertisement)
		if err != nil {
			// Discard invalid packets
			log.Printf("[DISCOVERY] invalid advertisement packet received, discarding\n")
			continue
		}

		// Validate that it's an org advertisement
		if advertisement.Type != "ORG_ADVERTISEMENT" {
			log.Printf("[DISCOVERY] received non-advertisement packet, discarding\n")
			continue
		}

		if advertisement.OrgName == s.JoinRequest.OrgName && advertisement.Passcode == s.JoinRequest.Passcode {
			log.Printf("[DISCOVERY] advertisement received from org=%s, bootstrap_ip=%s, control_port=%d\n",
				advertisement.OrgName,
				advertisement.BootstrapIP,
				advertisement.ControlPort,
			)
		}

		events <- &models.DiscoveryEvent{
			OrgName:     advertisement.OrgName,
			BootstrapIP: advertisement.BootstrapIP,
			Passcode:    advertisement.Passcode,
		}
	}
}
