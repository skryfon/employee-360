package eventing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skryfon/employee360/backend/internal/domain/event"
	"github.com/skryfon/employee360/backend/internal/infrastructure/database"
)

func TestRiverPublisher_RequiresTransaction(t *testing.T) {
	p := NewRiverPublisher(nil, nil)
	err := p.Publish(context.Background(), event.Event{EventType: event.EventTypeUserInvited})
	require.ErrorIs(t, err, database.ErrNoSQLTx)
}

func TestRiverPublisher_NoEventsIsNoop(t *testing.T) {
	require.NoError(t, NewRiverPublisher(nil, nil).Publish(context.Background()))
}
