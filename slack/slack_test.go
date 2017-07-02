package slack_test

import (
	"testing"

	. "github.com/Emyrk/LendingBot/slack"
)

func Test_message_alert(t *testing.T) {
	err := SendMessage(":+1:", "testBot", "tests", "@channel test")
	if err != nil {
		for _, e := range *err {
			t.Errorf("Sending slack: %s\n", e.Error())
		}
	}
}
