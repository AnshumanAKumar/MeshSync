# MeshSync

Distributed local-first file synchronization system with centralized metadata coordination and peer-to-peer replication.

## Overview

MeshSync is a decentralized file synchronization platform designed for multi-device local network synchronization without cloud storage dependencies.

The system uses:

* UDP-based discovery
* HTTP control plane APIs
* Centralized metadata coordination
* Peer-to-peer file replication
* Event-driven runtime orchestration

The architecture is designed to be:

* hardware agnostic
* operating-system agnostic
* cloud independent
* distributed systems oriented

## Core Features

### v0.1 Features

#### Organization-Based Device Clustering

Devices join a common organization using:

* org name
* passcode

Bootstrap node manages onboarding and cluster membership.

#### UDP Discovery

Bootstrap node periodically broadcasts:

* org name
* bootstrap IP
* control port

Peer nodes listen for matching organization advertisements.

#### HTTP-Based Onboarding

Peers establish onboarding sessions using HTTP APIs.

Bootstrap validates:

* organization
* passcode
* passcode TTL

On successful onboarding:

* node identity is generated
* session identity is generated
* peer registry is updated

#### Distributed File Synchronization

All nodes are writable.

Every node:

* monitors local filesystem
* emits file events
* synchronizes metadata with bootstrap

Bootstrap coordinates replication.

Actual file transfer happens peer-to-peer.

#### Metadata Coordination

Bootstrap maintains:

* cluster topology
* peer registry
* file metadata registry
* synchronization state

Bootstrap acts as:

* control plane coordinator
* metadata authority

Bootstrap does not act as canonical file storage.

#### Peer-to-Peer File Replication

Actual file transfer occurs directly between peers.

Bootstrap only coordinates:

* replication targets
* synchronization instructions
* metadata ownership

## Architecture

### Control Plane

Managed by bootstrap node.

Responsibilities:

* discovery
* onboarding
* metadata coordination
* replication orchestration
* cluster topology management

Transport:

* UDP discovery
* HTTP APIs

### Data Plane

Managed by peer nodes.

Responsibilities:

* local file storage
* watcher service
* transfer service
* replication downloads/uploads

Transport:

* peer-to-peer transfer protocol

## Runtime Model

MeshSync uses an event-driven runtime architecture.

Each node runs:

* runtime orchestrator
* discovery service
* onboarding service
* watcher service
* transfer service

Services emit runtime events using channels.

Runtime consumes and orchestrates service interactions.

### Bootstrap Runtime

Responsibilities:

* organization initialization
* discovery broadcasting
* onboarding server
* metadata coordination
* replication orchestration
* peer registry management

### Peer Runtime

Responsibilities:

* discovery listener
* onboarding client
* watcher service
* metadata synchronization
* transfer participation

## Synchronization Model

### Multi-Writer Synchronization

All nodes can:

* create files
* modify files
* delete files

Watcher service runs on every node.

### Conflict Resolution

v0.1 uses Last Write Wins.

Latest modification timestamp becomes authoritative metadata state.

## Repository Structure

```text
meshsync/
├── cmd/
│   ├── root.go
│   └── main.go
│
├── internal/
│   ├── runtime/
│   ├── discovery/
│   ├── onboarding/
│   ├── metadata/
│   ├── watcher/
│   ├── transfer/
│   ├── org/
│   └── models/
│
├── README.md
└── DESIGN.md
```

## Running MeshSync

### Bootstrap Node

```bash
./meshsync start --bootstrap
```

### Peer Node

```bash
./meshsync start --join
```

## Future Scope

### Planned Improvements

* persistent metadata storage
* distributed metadata replication
* resumable file transfers
* heartbeat service
* peer failover
* checksum validation
* delta synchronization
* file chunking
* encrypted transfers
* distributed conflict resolution
* multi-bootstrap coordination

## DESIGN DOCUMENT

### System Overview

MeshSync is a distributed local-first synchronization platform consisting of:

* bootstrap coordinator node
* peer storage nodes
* event-driven runtimes
* centralized metadata coordination
* peer-to-peer replication

The system separates:

* control plane
* data plane

for cleaner distributed coordination.

### High-Level Architecture

```text
                  ┌────────────────────┐
                  │   Bootstrap Node   │
                  │--------------------│
                  │ Discovery Service  │
                  │ Onboarding Server  │
                  │ Metadata Registry  │
                  │ Replication Coord. │
                  └─────────┬──────────┘
                            │
                  HTTP Control Plane
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        │                   │                   │
┌───────▼────────┐ ┌───────▼────────┐ ┌───────▼────────┐
│    Peer A      │ │    Peer B      │ │    Peer C      │
│----------------│ │----------------│ │----------------│
│ Watcher        │ │ Watcher        │ │ Watcher        │
│ Transfer Svc   │ │ Transfer Svc   │ │ Transfer Svc   │
│ Local Storage  │ │ Local Storage  │ │ Local Storage  │
└────────────────┘ └────────────────┘ └────────────────┘
```

### Discovery Layer

#### Purpose

Enable peer discovery of bootstrap node inside local network.

#### Transport

UDP Broadcast

#### Bootstrap Responsibilities

Broadcast:

* org name
* bootstrap IP
* control port

#### Peer Responsibilities

Listen for:

* matching org advertisements
* bootstrap control plane endpoints

### Onboarding Layer

#### Purpose

Establish authenticated cluster session between peer and bootstrap.

#### Transport

HTTP APIs

#### Endpoint

```http
POST /api/v1/onboarding/join
```

#### Request

```json
{
  "org_name": "mesh-org",
  "passcode": "ABC123",
  "device_name": "peer-a",
  "device_ip": "192.168.1.10",
  "control_port": 8081,
  "transfer_port": 9090
}
```

#### Response

```json
{
  "status": "success",
  "message": "onboarding successful",
  "node_id": "uuid",
  "session_id": "uuid",
  "heartbeat_interval": 30
}
```

### Metadata Layer

#### Purpose

Maintain authoritative cluster metadata state.

#### Owned By

Bootstrap node.

#### Responsibilities

* peer registry
* file metadata registry
* synchronization state
* replication coordination
* topology management

### Peer Registry

```go
type Peer struct {
    DeviceID     string
    SessionID    string

    DeviceName   string
    DeviceIP     string

    ControlPort  int
    TransferPort int

    JoinedAt     time.Time
    LastSeen     time.Time

    Status        PeerStatus
}
```

### Watcher Layer

#### Purpose

Monitor local filesystem mutations.

#### Deployment

Watcher runs on:

* bootstrap
* all peers

#### Events

Watcher detects:

* CREATE
* MODIFY
* DELETE
* RENAME

#### Watcher Output

```go
type FileEvent struct {
    EventType  string
    FilePath   string
    Checksum   string
    ModifiedAt time.Time
}
```

### Synchronization Flow

```text
Filesystem Mutation
        ↓
Watcher Event
        ↓
Runtime Event Loop
        ↓
Metadata Update
        ↓
Bootstrap Registry Update
        ↓
Replication Decision
        ↓
Peer-to-Peer Transfer
```

### Replication Layer

#### Purpose

Synchronize file contents between peers.

#### Architecture

Bootstrap coordinates replication.

Peers transfer files directly.

### Replication Flow

```text
Peer A modifies file
        ↓
Peer A sends metadata update
        ↓
Bootstrap updates metadata registry
        ↓
Bootstrap determines outdated peers
        ↓
Bootstrap sends replication instruction
        ↓
Peer B downloads file from Peer A
```

### Runtime Architecture

MeshSync uses event-driven runtimes.

Services communicate using channels.

Runtime orchestrates:

* discovery events
* onboarding events
* watcher events
* metadata events
* replication events
* transfer events

### Event-Driven Runtime Model

```text
Service
   ↓ emits event
Runtime
   ↓ orchestrates next action
Another Service
```

### Design Principles

* local-first synchronization
* decentralized file storage
* centralized metadata coordination
* peer-to-peer replication
* event-driven orchestration
* hardware agnostic architecture
* operating-system agnostic architecture
* cloud-independent deployment
