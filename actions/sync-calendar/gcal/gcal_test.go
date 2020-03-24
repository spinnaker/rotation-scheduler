package gcal

import (
	"cloud.google.com/go/httpreplay"
	"context"
	"github.com/spinnaker/rotation-scheduler/proto/schedule"
	"testing"
)

const (
	testCalendarID = "spinbot@spinnaker.io"
)

func TestSchedule(t *testing.T) {
	replayer, err := httpreplay.NewReplayer("testing/schedule.replay")
	if err != nil {
		t.Fatalf("cannot initializer HTTP replayer: %v", err)
	}
	defer replayer.Close()

	client, err := replayer.Client(context.Background())
	if err != nil {
		t.Fatalf("cannot initialize client from replayer: %v", err)
	}

	g, err := NewGCal(testCalendarID, client)
	if err != nil {
		t.Fatalf("cannot create new gcal: %v", err)
	}

	s, err := schedule.FromYAMLFile("testing/test_schedule.yaml")
	if err != nil {
		t.Fatalf("cannot read schedule from yaml file: %v", err)
	}

	stop, err := s.EstimateStopTime()
	if err != nil {
		t.Errorf("cannot calculate stop time: %v", err)
	}
	err = g.Schedule(s, stop)
	if err != nil {
		t.Errorf("error during scheduling: %v", err)
	}
}
