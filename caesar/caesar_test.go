package caesar

import (
	"conalg/config"
	"conalg/model"
	"testing"
	"time"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/gookit/slog"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/stretchr/testify/assert"
)

type SampleApp struct {
	conalg Conalg
}

func (s *SampleApp) DetermineConflict(c1, c2 []byte) bool {
	slog.Debugf("Conflict determined: %s, %s = %v", c1, c2, len(c1) == len(c2))
	return len(c1) == len(c2)
}

func (s *SampleApp) Execute(c []byte) {
	slog.Infof(" ... doing whatever i want with %s", c)
}

func (s *SampleApp) SetConalgModule(m Conalg) {
	s.conalg = m
}

func TestComputePred_EmptyWhitelist(t *testing.T) {
	app := &SampleApp{}
	caesar := NewCaesar(config.Config{}, nil, app)
	app.SetConalgModule(caesar)

	// Define test inputs
	reqID := "testReqID"
	payload := []byte("payload1")
	timestamp := uint64(32)
	whitelist := gs.NewSet[string]()

	// Create a mock history with some requests
	history := cmap.New[model.Request]()
	history.Set("reqID1", model.Request{ID: "reqID1", Payload: []byte("payload1"), Timestamp: 34, Status: model.SLOW_PEND})
	history.Set("reqID2", model.Request{ID: "reqID2", Payload: []byte("payload2"), Timestamp: 30, Status: model.ACC})
	history.Set("reqID3", model.Request{ID: "reqID3", Payload: []byte("payload34"), Timestamp: 21, Status: model.STABLE})

	// Set the mock history in the Caesar instance
	caesar.History = history

	// Call the computePred function
	pred := caesar.computePred(reqID, payload, timestamp, whitelist)

	// Assert the expected result
	expectedPred := gs.NewSet[string]()
	expectedPred.Add("reqID2")
	assert.Equal(t, expectedPred, pred)
}

func TestComputePred_NilWhitelist(t *testing.T) {
	app := &SampleApp{}
	caesar := NewCaesar(config.Config{}, nil, app)
	app.SetConalgModule(caesar)

	// Define test inputs
	reqID := "testReqID"
	payload := []byte("payload1")
	timestamp := uint64(32)

	// Create a mock history with some requests
	history := cmap.New[model.Request]()
	history.Set("reqID1", model.Request{ID: "reqID1", Payload: []byte("payload1"), Timestamp: 34, Status: model.SLOW_PEND})
	history.Set("reqID2", model.Request{ID: "reqID2", Payload: []byte("payload2"), Timestamp: 30, Status: model.ACC})
	history.Set("reqID3", model.Request{ID: "reqID3", Payload: []byte("payload34"), Timestamp: 21, Status: model.STABLE})

	// Set the mock history in the Caesar instance
	caesar.History = history

	// Call the computePred function
	pred := caesar.computePred(reqID, payload, timestamp, nil)

	// Assert the expected result
	expectedPred := gs.NewSet[string]()
	expectedPred.Add("reqID2")
	assert.Equal(t, expectedPred, pred)
}

func TestComputePred_WithWhitelist(t *testing.T) {
	app := &SampleApp{}
	caesar := NewCaesar(config.Config{}, nil, app)
	app.SetConalgModule(caesar)

	// Define test inputs
	reqID := "testReqID"
	payload := []byte("payload1")
	timestamp := uint64(32)
	whitelist := gs.NewSet[string]()
	whitelist.Add("reqID1")

	// Create a mock history with some requests
	history := cmap.New[model.Request]()
	history.Set("reqID1", model.Request{ID: "reqID1", Payload: []byte("payload1"), Timestamp: 34, Status: model.SLOW_PEND})
	history.Set("reqID2", model.Request{ID: "reqID2", Payload: []byte("payload2"), Timestamp: 30, Status: model.ACC})
	history.Set("reqID2.5", model.Request{ID: "reqID2.5", Payload: []byte("payload2"), Timestamp: 31, Status: model.SLOW_PEND})
	history.Set("reqID3", model.Request{ID: "reqID3", Payload: []byte("payload34"), Timestamp: 21, Status: model.STABLE})
	history.Set("reqID4", model.Request{ID: "reqID4", Payload: []byte("payload6"), Timestamp: 22, Status: model.PRE_FAST_PEND})

	// Set the mock history in the Caesar instance
	caesar.History = history

	// Call the computePred function
	pred := caesar.computePred(reqID, payload, timestamp, whitelist)

	// Assert the expected result
	expectedPred := gs.NewSet[string]()
	expectedPred.Add("reqID2")
	expectedPred.Add("reqID2.5")
	expectedPred.Add("reqID1")
	assert.Equal(t, expectedPred, pred)
}
func TestComputeWaitlist_NoError(t *testing.T) {
	app := &SampleApp{}
	caesar := NewCaesar(config.Config{}, nil, app)
	app.SetConalgModule(caesar)

	// Define test inputs
	reqID := "testReqID"
	payload := []byte("payload1")
	timestamp := uint64(32)

	// Create a mock history with some requests
	history := cmap.New[model.Request]()
	history.Set("reqID1", model.Request{ID: "reqID1", Payload: []byte("payload1"), Timestamp: 34, Status: model.SLOW_PEND, Pred: gs.NewSet[string]()})
	history.Set("reqID2", model.Request{ID: "reqID2", Payload: []byte("payload25"), Timestamp: 30, Status: model.ACC, Pred: gs.NewSet[string]()})
	history.Set("reqID3", model.Request{ID: "reqID3", Payload: []byte("payload34"), Timestamp: 21, Status: model.STABLE, Pred: gs.NewSet[string]()})
	history.Set("reqID4", model.Request{ID: "reqID4", Payload: []byte("payload1"), Timestamp: 37, Status: model.FAST_PEND, Pred: gs.NewSet[string]("testReqID")})

	// Set the mock history in the Caesar instance
	caesar.History = history

	// Call the computeWaitgroup function
	waitgroup, err := caesar.computeWaitlist(reqID, payload, timestamp)

	// Assert the expected result
	expectedWaitgroup := gs.NewSet[string]()
	expectedWaitgroup.Add("reqID1")
	assert.NoError(t, err)
	assert.Equal(t, expectedWaitgroup, waitgroup)
}

func TestComputeWaitlist_ErrorAutoNack(t *testing.T) {
	app := &SampleApp{}
	caesar := NewCaesar(config.Config{}, nil, app)
	app.SetConalgModule(caesar)

	// Define test inputs
	reqID := "testReqID"
	payload := []byte("payload1")
	timestamp := uint64(32)

	// Create a mock history with some requests
	history := cmap.New[model.Request]()
	history.Set("reqID1", model.Request{ID: "reqID1", Payload: []byte("payload1"), Timestamp: 34, Status: model.SLOW_PEND, Pred: gs.NewSet[string]()})
	history.Set("reqID2", model.Request{ID: "reqID2", Payload: []byte("payload25"), Timestamp: 30, Status: model.ACC, Pred: gs.NewSet[string]()})
	history.Set("reqID3", model.Request{ID: "reqID3", Payload: []byte("payload34"), Timestamp: 21, Status: model.STABLE, Pred: gs.NewSet[string]()})
	history.Set("reqID4", model.Request{ID: "reqID4", Payload: []byte("payload1"), Timestamp: 37, Status: model.STABLE, Pred: gs.NewSet[string]()})

	// Set the mock history in the Caesar instance
	caesar.History = history

	// Call the computeWaitgroup function
	waitgroup, err := caesar.computeWaitlist(reqID, payload, timestamp)

	// Assert the expected result
	assert.Error(t, err)
	assert.Nil(t, waitgroup)
}
func TestWait_Succesful(t *testing.T) {
	app := &SampleApp{}
	caesar := NewCaesar(config.Config{}, nil, app)
	app.SetConalgModule(caesar)

	// Define test inputs
	id := "testReqID"
	payload := []byte("payload1")
	timestamp := uint64(32)

	// Create a mock history
	history := cmap.New[model.Request]()
	history.Set("reqID1", model.Request{ID: "reqID1", Payload: []byte("payload1"), Timestamp: 34, Status: model.SLOW_PEND, Pred: gs.NewSet[string]()})
	history.Set("reqID2", model.Request{ID: "reqID2", Payload: []byte("payload25"), Timestamp: 30, Status: model.ACC, Pred: gs.NewSet[string]()})
	history.Set("reqID3", model.Request{ID: "reqID3", Payload: []byte("payload34"), Timestamp: 21, Status: model.STABLE, Pred: gs.NewSet[string]()})
	history.Set("reqID4", model.Request{ID: "reqID4", Payload: []byte("payload1"), Timestamp: 37, Status: model.FAST_PEND, Pred: gs.NewSet[string]("testReqID")})
	caesar.History = history

	go func() {
		time.Sleep(1 * time.Second)
		caesar.Publisher.Publish(model.StatusUpdate{RequestID: "reqID1", Status: model.STABLE, Pred: gs.NewSet[string]("testReqID")})
	}()

	// Call the wait function
	result := caesar.wait(id, payload, timestamp)

	// Assert the expected result
	assert.True(t, result)
}

func TestWait_Unsuccessful(t *testing.T) {
	app := &SampleApp{}
	caesar := NewCaesar(config.Config{}, nil, app)
	app.SetConalgModule(caesar)

	// Define test inputs
	id := "testReqID"
	payload := []byte("payload1")
	timestamp := uint64(32)

	// Create a mock history
	history := cmap.New[model.Request]()
	history.Set("reqID1", model.Request{ID: "reqID1", Payload: []byte("payload1"), Timestamp: 34, Status: model.SLOW_PEND, Pred: gs.NewSet[string]()})
	history.Set("reqID2", model.Request{ID: "reqID2", Payload: []byte("payload25"), Timestamp: 30, Status: model.ACC, Pred: gs.NewSet[string]()})
	history.Set("reqID3", model.Request{ID: "reqID3", Payload: []byte("payload34"), Timestamp: 21, Status: model.STABLE, Pred: gs.NewSet[string]()})
	history.Set("reqID4", model.Request{ID: "reqID4", Payload: []byte("payload1"), Timestamp: 37, Status: model.FAST_PEND, Pred: gs.NewSet[string]("testReqID")})
	caesar.History = history

	go func() {
		time.Sleep(1 * time.Second)
		caesar.Publisher.Publish(model.StatusUpdate{RequestID: "reqID1", Status: model.STABLE, Pred: gs.NewSet[string]("blablalba")})
	}()

	// Call the wait function
	result := caesar.wait(id, payload, timestamp)

	// Assert the expected result
	assert.False(t, result)
}
