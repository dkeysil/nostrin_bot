package nostrutils

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
)

func QueryPoolSync(ctx context.Context, pool *nostr.SimplePool, filters nostr.Filters) []*nostr.Event {
	relays := make([]string, 0, len(pool.Relays))
	for url := range pool.Relays {
		relays = append(relays, url)
	}

	events := pool.SubManyEose(ctx, relays, filters)

	var evs []*nostr.Event
	for ev := range events {
		evs = append(evs, ev)
	}

	return evs
}
