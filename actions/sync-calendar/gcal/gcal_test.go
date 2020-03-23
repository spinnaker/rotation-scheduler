package gcal

import (
	"context"
	"io/ioutil"
	"testing"

	"cloud.google.com/go/httpreplay"
	"github.com/ghodss/yaml"
	"github.com/spinnaker/rotation-scheduler/schedule"
)

const (
	testCalendarID = "spinbot@spinnaker.io"
)

func TestSchedule(t *testing.T) {
	replayer, err := httpreplay.NewReplayer("testing/schedule.replay")
	if err != nil {
		t.Fatalf("cannot initializer HTTP replayer: %v", err)
	}

	client, err := replayer.Client(context.Background())
	if err != nil {
		t.Fatalf("cannot initialize client from replayer: %v", err)
	}

	gcalUnderTest, err := NewGCal(testCalendarID, client)
	if err != nil {
		t.Fatalf("cannot create new gcal: %v", err)
	}

	testSchedBytes, err := ioutil.ReadFile("testing/test_schedule.yaml")
	if err != nil {
		t.Fatalf("cannot read test schedule: %v", err)
	}

	testSched := &schedule.Schedule{}
	err = yaml.Unmarshal(testSchedBytes, testSched)
	if err != nil {
		t.Fatalf("cannot read schedule from yaml file: %v", err)
	}

	err = gcalUnderTest.Schedule(testSched)
	if err != nil {
		t.Errorf("error during scheduling: %v", err)
	}
}
