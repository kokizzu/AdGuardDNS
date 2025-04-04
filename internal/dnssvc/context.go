package dnssvc

import (
	"context"
	"time"

	"github.com/AdguardTeam/AdGuardDNS/internal/agd"
	"github.com/AdguardTeam/golibs/contextutil"
)

// contextConstructor is a [contextutil.Constructor] implementation that returns
// a context with the given timeout as well as a new [agd.RequestID].
type contextConstructor struct {
	timeout time.Duration
}

// newContextConstructor returns a new properly initialized *contextConstructor.
func newContextConstructor(timeout time.Duration) (c *contextConstructor) {
	return &contextConstructor{
		timeout: timeout,
	}
}

// type check
var _ contextutil.Constructor = (*contextConstructor)(nil)

// New implements the [contextutil.Constructor] interface for
// *contextConstructor.  It returns a context with a new [agd.RequestID] as well
// as its timeout and the corresponding cancelation function.
func (c *contextConstructor) New(
	parent context.Context,
) (ctx context.Context, cancel context.CancelFunc) {
	ctx, cancel = context.WithTimeout(parent, c.timeout)
	ctx = agd.WithRequestID(ctx, agd.NewRequestID())

	return ctx, cancel
}
