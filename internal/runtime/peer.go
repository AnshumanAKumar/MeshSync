package runtime

import (
	"context"
	"fmt"
	"time"

	"meshsync/internal/discovery"
	"meshsync/internal/models"
	"meshsync/internal/onboarding"
	"meshsync/internal/transfer"
	"meshsync/internal/watcher"
)

func (r *Runtime) startPeer(
	ctx context.Context,
) error {

	fmt.Println(
		"[PEER] initializing peer runtime",
	)

	// =========================================================
	// USER INPUT
	// =========================================================

	var orgName string

	fmt.Print("Enter org Name: ")

	fmt.Scanln(&orgName)

	var passcode string

	fmt.Print("Enter passcode for the org: ")

	fmt.Scanln(&passcode)

	// =========================================================
	// JOIN REQUEST
	// =========================================================

	joinRequest := &models.JoinRequest{
		OrgName:  orgName,
		Passcode: passcode,
	}

	// =========================================================
	// LOCAL PEER CONFIGURATION
	// =========================================================

	localPeer := &models.Peer{
		DeviceName: "laptop-b",
		DeviceIP:   "192.168.1.10",

		ControlPort:  8081,
		TransferPort: 9090,
	}

	// =========================================================
	// EVENT CHANNELS
	// =========================================================

	discoveryEvents := make(
		chan *models.DiscoveryEvent,
		32,
	)

	watcherEvents := make(
		chan *models.FileEvent,
		128,
	)

	metadataEvents := make(
		chan *models.MetadataEvent,
		128,
	)

	replicationEvents := make(
		chan *models.ReplicationEvent,
		128,
	)

	transferEvents := make(
		chan *models.TransferEvent,
		128,
	)

	// =========================================================
	// DISCOVERY SERVICE
	// =========================================================

	discoveryService := discovery.NewDiscoveryService(
		nil,
		joinRequest,
	)

	go discoveryService.StartListener(
		ctx,
		discoveryEvents,
	)

	fmt.Println(
		"[PEER] runtime started",
	)

	onboarded := false

	// =========================================================
	// RUNTIME EVENT LOOP
	// =========================================================

	for {

		select {

		// =====================================================
		// DISCOVERY EVENTS
		// =====================================================

		case event := <-discoveryEvents:

			if onboarded {
				continue
			}

			fmt.Printf(
				"[PEER] bootstrap discovered ip=%s\n",
				event.BootstrapIP,
			)

			// =================================================
			// ONBOARDING CLIENT
			// =================================================

			onboardingService := onboarding.NewOnboardingService(
				event.OrgName,
				joinRequest.Passcode,
				time.Now(),
			)

			response, err := onboardingService.StartClient(
				event.BootstrapIP,
				8080,
				&models.OnboardingRequest{
					OrgName:  event.OrgName,
					Passcode: joinRequest.Passcode,

					DeviceName: localPeer.DeviceName,
					DeviceIP:   localPeer.DeviceIP,

					ControlPort:  localPeer.ControlPort,
					TransferPort: localPeer.TransferPort,
				},
			)

			if err != nil {

				fmt.Printf(
					"[PEER] onboarding failed: %v\n",
					err,
				)

				continue
			}

			fmt.Printf(
				"[PEER] onboarding response status=%s node=%s session=%s\n",
				response.Status,
				response.NodeID,
				response.SessionID,
			)

			if response.Status != "success" {
				continue
			}

			// =================================================
			// UPDATE LOCAL SESSION STATE
			// =================================================

			localPeer.DeviceID = response.NodeID
			localPeer.SessionID = response.SessionID

			localPeer.JoinedAt = time.Now()
			localPeer.LastSeen = time.Now()

			localPeer.Status = models.PeerStatusOnline

			onboarded = true

			fmt.Println(
				"[PEER] successfully onboarded to cluster",
			)

			// =================================================
			// WATCHER SERVICE
			// =================================================

			watcherService := watcher.NewWatcherService(
				"/meshsync",
			)

			go func() {

				if err := watcherService.StartWatcher(
					ctx,
					"/meshsync",
					watcherEvents,
				); err != nil {

					fmt.Printf(
						"[WATCHER] error: %v\n",
						err,
					)
				}
			}()

			// =================================================
			// TRANSFER SERVICE
			// =================================================

			transferService := transfer.NewTransferService(
				"/meshsync",
				localPeer.TransferPort,
			)

			go func() {

				if err := transferService.StartServer(
					ctx,
					localPeer.TransferPort,
					transferEvents,
				); err != nil && err != context.Canceled {

					fmt.Printf(
						"[TRANSFER] server error: %v\n",
						err,
					)
				}
			}()

			// =================================================
			// HEARTBEAT SERVICE
			// =================================================

			/*
				heartbeatService := heartbeat.NewHeartbeatService()

				go heartbeatService.StartHeartbeatLoop(
					ctx,
					localPeer,
				)
			*/

		// =====================================================
		// WATCHER EVENTS
		// =====================================================

		case event := <-watcherEvents:

			fmt.Printf(
				"[WATCHER] file event=%s path=%s size=%d\n",
				event.EventType,
				event.FilePath,
				event.FileSize,
			)

			// Future:
			// send metadata update to bootstrap
			// POST /api/v1/metadata/update

		// =====================================================
		// METADATA EVENTS
		// =====================================================

		case event := <-metadataEvents:

			fmt.Printf(
				"[METADATA] sync event path=%s version=%d\n",
				event.FilePath,
				event.Version,
			)

			// Future:
			// process metadata updates
			// determine local sync state

		// =====================================================
		// REPLICATION EVENTS
		// =====================================================

		case event := <-replicationEvents:

			fmt.Printf(
				"[REPLICATION] source=%s file=%s\n",
				event.SourcePeerID,
				event.FilePath,
			)

			// Future:
			// download file from source peer

			/*
				go transferService.DownloadFile(
					sourcePeerIP,
					sourcePeerPort,
					event.FilePath,
					localPeer.DeviceID,
				)
			*/

		// =====================================================
		// TRANSFER EVENTS
		// =====================================================

		case event := <-transferEvents:

			fmt.Printf(
				"[TRANSFER] completed file=%s size=%d\n",
				event.FilePath,
				event.FileSize,
			)

			// Future:
			// notify bootstrap replication completed
			// update local sync metadata

		// =====================================================
		// SHUTDOWN
		// =====================================================

		case <-ctx.Done():

			fmt.Println(
				"[PEER] shutting down runtime",
			)

			return nil
		}
	}
}
