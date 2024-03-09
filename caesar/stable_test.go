package caesar

import (
	"conalg/model"
	"testing"

	gs "github.com/deckarep/golang-set/v2"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/stretchr/testify/assert"
)

func TestBreakLoop(t *testing.T) {
	// Create a new Caesar instance
	caesar := &Caesar{}

	// Create a mock history with some requests
	history := cmap.New[model.Request]()
	history.Set("reqID1", model.Request{ID: "reqID1", Status: model.STABLE, Timestamp: 30, Pred: gs.NewSet[string]("predID1", "predID2")})
	history.Set("predID1", model.Request{ID: "predID1", Status: model.STABLE, Timestamp: 10, Pred: gs.NewSet[string]("reqID1")})
	history.Set("predID2", model.Request{ID: "predID2", Status: model.STABLE, Timestamp: 40, Pred: gs.NewSet[string]()})
	caesar.History = history

	err := caesar.breakLoop("reqID1")
	assert.NoError(t, err)

	// Verify the updated history
	expectedHistory := cmap.New[model.Request]()
	expectedHistory.Set("reqID1", model.Request{ID: "reqID1", Timestamp: 30, Status: model.STABLE, Pred: gs.NewSet[string]("predID1")})
	expectedHistory.Set("predID1", model.Request{ID: "predID1", Status: model.STABLE, Timestamp: 10, Pred: gs.NewSet[string]()})
	expectedHistory.Set("predID2", model.Request{ID: "predID2", Status: model.STABLE, Timestamp: 40, Pred: gs.NewSet[string]()})

	assert.Equal(t, expectedHistory.Items(), caesar.History.Items())
}
