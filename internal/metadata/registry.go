package metadata

import (
	"sync"
	"time"

	"meshsync/internal/models"
)

// FileMetadata tracks file version and ownership
type FileMetadata struct {
	FilePath     string
	Version      int
	Checksum     string
	FileSize     int64
	OwnerPeerID  string
	LastModified time.Time
	UpdatedAt    time.Time
}

// PeerFileState tracks which files a peer has
type PeerFileState struct {
	PeerID       string
	FilePath     string
	Version      int
	LastSyncedAt time.Time
}

type Registry struct {
	mu             sync.RWMutex
	Peers          map[string]*models.Peer
	FileMetadata   map[string]*FileMetadata    // key: filePath
	PeerFileStates map[string][]*PeerFileState // key: peerID
}

func NewRegistry() *Registry {

	return &Registry{
		Peers:          make(map[string]*models.Peer),
		FileMetadata:   make(map[string]*FileMetadata),
		PeerFileStates: make(map[string][]*PeerFileState),
	}
}

func (r *Registry) AddPeer(
	peer *models.Peer,
) {

	r.mu.Lock()
	defer r.mu.Unlock()

	r.Peers[peer.DeviceID] = peer
	if _, exists := r.PeerFileStates[peer.DeviceID]; !exists {
		r.PeerFileStates[peer.DeviceID] = make([]*PeerFileState, 0)
	}
}

func (r *Registry) GetPeers() []*models.Peer {

	r.mu.RLock()
	defer r.mu.RUnlock()

	peers := make([]*models.Peer, 0)

	for _, peer := range r.Peers {
		peers = append(peers, peer)
	}

	return peers
}

// UpdateFileMetadata updates or creates file metadata
func (r *Registry) UpdateFileMetadata(
	filePath string,
	ownerPeerID string,
	checksum string,
	fileSize int64,
) {

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, found := r.FileMetadata[filePath]

	if found && existing.Checksum == checksum {
		// File hasn't changed
		return
	}

	version := 1
	if found {
		version = existing.Version + 1
	}

	r.FileMetadata[filePath] = &FileMetadata{
		FilePath:     filePath,
		Version:      version,
		Checksum:     checksum,
		FileSize:     fileSize,
		OwnerPeerID:  ownerPeerID,
		LastModified: time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// GetFileMetadata retrieves file metadata
func (r *Registry) GetFileMetadata(filePath string) *FileMetadata {

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.FileMetadata[filePath]
}

// GetOutdatedPeers returns peers that don't have the latest version of a file
func (r *Registry) GetOutdatedPeers(filePath string) []*models.Peer {

	r.mu.RLock()
	defer r.mu.RUnlock()

	fileMetadata, found := r.FileMetadata[filePath]
	if !found {
		return []*models.Peer{}
	}

	outdatedPeers := make([]*models.Peer, 0)

	for peerID, peer := range r.Peers {
		// Skip the owner peer
		if peerID == fileMetadata.OwnerPeerID {
			continue
		}

		// Check if peer has the file or has outdated version
		hasCurrent := false
		for _, state := range r.PeerFileStates[peerID] {
			if state.FilePath == filePath && state.Version == fileMetadata.Version {
				hasCurrent = true
				break
			}
		}

		if !hasCurrent {
			outdatedPeers = append(outdatedPeers, peer)
		}
	}

	return outdatedPeers
}

// UpdatePeerFileState updates the sync state of a file on a peer
func (r *Registry) UpdatePeerFileState(
	peerID string,
	filePath string,
	version int,
) {

	r.mu.Lock()
	defer r.mu.Unlock()

	states, exists := r.PeerFileStates[peerID]
	if !exists {
		states = make([]*PeerFileState, 0)
	}

	// Update or add file state
	found := false
	for _, state := range states {
		if state.FilePath == filePath {
			state.Version = version
			state.LastSyncedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		states = append(states, &PeerFileState{
			PeerID:       peerID,
			FilePath:     filePath,
			Version:      version,
			LastSyncedAt: time.Now(),
		})
	}

	r.PeerFileStates[peerID] = states
}

// GetAllFiles returns all tracked files
func (r *Registry) GetAllFiles() []*FileMetadata {

	r.mu.RLock()
	defer r.mu.RUnlock()

	files := make([]*FileMetadata, 0)
	for _, metadata := range r.FileMetadata {
		files = append(files, metadata)
	}

	return files
}
