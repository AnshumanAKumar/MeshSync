package runtime

import (
	"context"
	"fmt"
	"time"

	"meshsync/internal/discovery"
	"meshsync/internal/metadata"
	"meshsync/internal/models"
	"meshsync/internal/onboarding"
	"meshsync/internal/org"
	"meshsync/internal/transfer"
	"meshsync/internal/watcher"
)

func (r *Runtime) startBootstrap(
	ctx context.Context,
) error {

	fmt.Println(
		"[BOOTSTRAP] initializing bootstrap runtime",
	)

	// =========================================================
	// ORGANIZATION INITIALIZATION
	// =========================================================

	var orgName string

	fmt.Print("Enter org Name: ")

	fmt.Scanln(&orgName)

	orgService := org.New()

	orgModel, err := orgService.CreateOrg(
		orgName,
		"",
	)

	if err != nil {
		return err
	}

	fmt.Printf(
		"[BOOTSTRAP] org created passcode=[%s]\n",
		orgModel.Passcode,
	)

	// =========================================================
	// CONSTANTS
	// =========================================================

	const bootstrapPeerID = "bootstrap"

	// =========================================================
	// METADATA REGISTRY
	// =========================================================

	registry := metadata.NewRegistry()

	// =========================================================
	// EVENT CHANNELS
	// =========================================================

	onboardingEvents := make(
		chan *models.OnboardingEvent,
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
		orgModel,
		nil,
	)

	go discoveryService.StartBroadcaster(ctx)

	// =========================================================
	// ONBOARDING SERVICE
	// =========================================================

	onboardingService := onboarding.NewOnboardingService(
		orgModel.Name,
		orgModel.Passcode,
		orgModel.ExpiresAt,
	)

	go onboardingService.StartServer(
		ctx,
		onboardingEvents,
	)

	// =========================================================
	// WATCHER SERVICE
	// =========================================================

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

	// =========================================================
	// TRANSFER SERVICE
	// =========================================================

	transferService := transfer.NewTransferService(
		"/meshsync",
		9090,
	)

	go func() {

		if err := transferService.StartServer(
			ctx,
			9090,
			transferEvents,
		); err != nil && err != context.Canceled {

			fmt.Printf(
				"[TRANSFER] server error: %v\n",
				err,
			)
		}
	}()

	fmt.Println(
		"[BOOTSTRAP] runtime started",
	)

	// =========================================================
	// RUNTIME EVENT LOOP
	// =========================================================

	for {

		select {

		// =====================================================
		// ONBOARDING EVENTS
		// =====================================================

		case event := <-onboardingEvents:

			fmt.Printf(
				"[BOOTSTRAP] peer onboarded device=%s id=%s ip=%s\n",
				event.Peer.DeviceName,
				event.Peer.DeviceID,
				event.Peer.DeviceIP,
			)

			registry.AddPeer(
				&event.Peer,
			)

			fmt.Printf(
				"[BOOTSTRAP] cluster topology updated peers=%d\n",
				len(registry.GetPeers()),
			)

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

			checksum := fmt.Sprintf(
				"%d-%d",
				event.FileSize,
				event.Timestamp,
			)

			registry.UpdateFileMetadata(
				event.FilePath,
				bootstrapPeerID,
				checksum,
				event.FileSize,
			)

			fileMetadata := registry.GetFileMetadata(
				event.FilePath,
			)

			if fileMetadata == nil {
				continue
			}

			fmt.Printf(
				"[METADATA] updated path=%s version=%d\n",
				event.FilePath,
				fileMetadata.Version,
			)

			outdatedPeers := registry.GetOutdatedPeers(
				event.FilePath,
			)

			for _, peer := range outdatedPeers {

				select {

				case replicationEvents <- &models.ReplicationEvent{
					SourcePeerID: bootstrapPeerID,
					TargetPeerID: peer.DeviceID,
					FilePath:     event.FilePath,
					Timestamp:    time.Now().UnixMilli(),
				}:

				case <-ctx.Done():
					return nil
				}
			}

		// =====================================================
		// METADATA EVENTS
		// =====================================================

		case event := <-metadataEvents:

			fmt.Printf(
				"[METADATA] processing path=%s version=%d\n",
				event.FilePath,
				event.Version,
			)

		// =====================================================
		// REPLICATION EVENTS
		// =====================================================

		case event := <-replicationEvents:

			fmt.Printf(
				"[REPLICATION] source=%s target=%s file=%s\n",
				event.SourcePeerID,
				event.TargetPeerID,
				event.FilePath,
			)

			var targetPeer *models.Peer

			for _, peer := range registry.GetPeers() {

				if peer.DeviceID == event.TargetPeerID {
					targetPeer = peer
					break
				}
			}

			if targetPeer == nil {

				fmt.Printf(
					"[REPLICATION] target peer not found=%s\n",
					event.TargetPeerID,
				)

				continue
			}

			fileMetadata := registry.GetFileMetadata(
				event.FilePath,
			)

			if fileMetadata == nil {

				fmt.Printf(
					"[REPLICATION] metadata missing file=%s\n",
					event.FilePath,
				)

				continue
			}

			fmt.Printf(
				"[REPLICATION] instruct peer=%s download file=%s version=%d\n",
				targetPeer.DeviceName,
				event.FilePath,
				fileMetadata.Version,
			)

			// future:
			// send replication HTTP request to peer

		// =====================================================
		// TRANSFER EVENTS
		// =====================================================

		case event := <-transferEvents:

			fmt.Printf(
				"[TRANSFER] completed peer=%s file=%s size=%d\n",
				event.PeerID,
				event.FilePath,
				event.FileSize,
			)

			fileMetadata := registry.GetFileMetadata(
				event.FilePath,
			)

			if fileMetadata == nil {
				continue
			}

			registry.UpdatePeerFileState(
				event.PeerID,
				event.FilePath,
				fileMetadata.Version,
			)

			fmt.Printf(
				"[METADATA] peer=%s synced version=%d\n",
				event.PeerID,
				fileMetadata.Version,
			)

		// =====================================================
		// SHUTDOWN
		// =====================================================

		case <-ctx.Done():

			fmt.Println(
				"[BOOTSTRAP] shutting down runtime",
			)

			return nil
		}
	}
}
