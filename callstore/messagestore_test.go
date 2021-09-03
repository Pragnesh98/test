package callstore

import (
	"fmt"
	"testing"
)

func TestGetMessagesAlterSlice(t *testing.T) {

	t.Run("GetLatency", func(t *testing.T) {
		mstore := MessageStore{}
		mstore.AddNewMessage("testSid_1", "test_step_1", "test message 1", Bot, "test_url_1")
		mstore.AddNewMessage("testSid_2", "test_step_2", "test message 2", Bot, "test_url_2")

		before := mstore.GetMessages("test_sid")
		before[0] = Message{StepIdentifier: "new"}

		after := mstore.GetMessages("test_sid")
		fmt.Printf("Message: [%#v]", after)
		if before[0].StepIdentifier == after[0].StepIdentifier {
			t.Errorf("Changes in before %v are reflected in After %v", before, after)
		}
	})
}
// func TestGetLatenciesDeleteSlice(t *testing.T) {

// 	t.Run("GetLatency", func(t *testing.T) {
// 		lstore := LatencyStore{}
// 		lstore.AddNewStep("testStep")
// 		lstore.RecordLatency("testStep", 1, 20)
// 		before := lstore.GetLatencies()
// 		before = nil

// 		after := lstore.GetLatencies()
// 		if after == nil {
// 			t.Errorf("Changes in before %v are reflected in After %v", before, after)
// 		}
// 	})
// }
