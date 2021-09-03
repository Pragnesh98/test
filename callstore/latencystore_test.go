package callstore

import (
	"fmt"
	"testing"
)

func TestGetLatenciesAlterSlice(t *testing.T) {

	t.Run("GetLatency", func(t *testing.T) {
		lstore := LatencyStore{}
		lstore.AddNewStep("testSid", "testStep")
		lstore.RecordLatency("testSid", "testStep", 1, 20)
		before := lstore.GetLatencies("testSid")
		before[0] = LatencyParameter{StepIdentifier: "new"}

		after := lstore.GetLatencies("testSid")
		
		if before[0].StepIdentifier == after[0].StepIdentifier {
			t.Errorf("Changes in before %v are reflected in After %v", before, after)
		}
	})
}
func TestGetLatenciesDeleteSlice(t *testing.T) {

	t.Run("GetLatency", func(t *testing.T) {
		lstore := LatencyStore{}
		lstore.AddNewStep("testSid", "testStep")
		lstore.RecordLatency("testSid", "testStep", 1, 20)
		before := lstore.GetLatencies("testSid")
		before = nil

		after := lstore.GetLatencies("testSid")
		if after == nil {
			t.Errorf("Changes in before %v are reflected in After %v", before, after)
		}
	})
}
