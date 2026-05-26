package models

type FileEvent struct {
	EventType string `json:"event_type"`
	FilePath  string `json:"file_path"`
	FileSize  int64  `json:"file_size"`

	Timestamp int64 `json:"timestamp"`
}

type MetadataEvent struct {
	FilePath string `json:"file_path"`
	FileSize int64  `json:"file_size"`
	Version  int    `json:"version"`

	Timestamp int64 `json:"timestamp"`
}

type ReplicationEvent struct {
	SourcePeerID string `json:"source_peer_id"`
	TargetPeerID string `json:"target_peer_id"`
	FilePath     string `json:"file_path"`

	Timestamp int64 `json:"timestamp"`
}

type TransferEvent struct {
	PeerID   string `json:"peer_id"`
	FilePath string `json:"file_path"`
	FileSize int64  `json:"file_size"`

	Status string `json:"status"`

	Timestamp int64 `json:"timestamp"`
}
