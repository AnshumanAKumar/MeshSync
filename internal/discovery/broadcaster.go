package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

// StartBroadcaster starts a goroutine that periodically broadcasts the org advertisement over UDP.
// It listens for the context cancellation to gracefully shut down the broadcaster.
// The broadcaster sends a JSON-encoded Advertisement packet containing the org details to a predefined broadcast address and port.
// The Advertisement struct includes the org name, passcode, bootstrap IP, and control port, which can be used by peer nodes to discover and join the org.
// The broadcaster runs in an infinite loop, sending the advertisement every 3 seconds until the context is cancelled.
// Note: In a real implementation, the bootstrap IP and control port would likely be dynamically determined rather than hardcoded.
func (s *DiscoveryService) StartBroadcaster(ctx context.Context) {

	//UDP broadcast logic here
	broadcastAddr := &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 9999,
	}

	conn, err := net.DialUDP("udp", nil, broadcastAddr)
	if err != nil {
		fmt.Printf("Failed to start broadcaster: %v\n", err)
		return
	}
	defer conn.Close()

	packet := Advertisement{
		Type:        "ORG_ADVERTISEMENT",
		OrgName:     s.Org.Name,
		Passcode:    s.Org.Passcode,
		BootstrapIP: "192.168.1.5",
		ControlPort: 8080,
	}

	fmt.Printf("[DISCOVERY] broadcasting org=%s",
		s.Org.Name,
	)

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():

			log.Println("[DISCOVERY] broadcaster shutting down")
			return
		case <-ticker.C:
			payload, err := json.Marshal(packet)
			if err != nil {
				log.Println(err)
				continue
			}

			_, err = conn.Write(payload)

			if err != nil {
				log.Println(err)
				continue
			}

			log.Printf(
				"[DISCOVERY] advertisement sent",
			)
		}
	}
}
