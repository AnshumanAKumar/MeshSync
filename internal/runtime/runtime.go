package runtime

import (
	"context"
	"fmt"
)

type Role string

const (
	BootstrapRole Role = "bootstrap"
	PeerRole      Role = "peer"
)

type Runtime struct {
	Ctx        context.Context
	Role       Role
	CancelFunc context.CancelFunc
}

// New creates a new Runtime instance with the specified role and initializes the context.
func New(role Role) *Runtime {
	ctx, cancel := context.WithCancel(context.Background())
	return &Runtime{
		Ctx:        ctx,
		Role:       role,
		CancelFunc: cancel,
	}
}

// Start initializes the runtime based on the specified role (bootstrap or peer).
func (r *Runtime) Start() error {

	fmt.Printf("[RUNTIME] starting node in %s mode\n", r.Role)

	switch r.Role {

	case BootstrapRole:
		return r.startBootstrap(r.Ctx)

	case PeerRole:
		return r.startPeer(r.Ctx)

	default:
		return fmt.Errorf("invalid runtime role")
	}
}
