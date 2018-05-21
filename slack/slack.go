package slack

import (
	"github.com/ashwanthkumar/slack-go-webhook"
)

func SendMessage(iconEmoji string, user string, channel string, message string) *[]error {
	webhookUrl := "https://hooks.slack.com/services/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"

	payload := slack.Payload(message, user, iconEmoji, channel, []slack.Attachment{})
	// payload := slack.Payload{
	// 	Text:        message,
	// 	Username:    user,
	// 	IconEmoji:   iconEmoji,
	// 	Channel:     channel,
	// 	Attachments: []slack.Attachment{},
	// }
	// payload :=  slack.Payload{
	// 	Text:        message,
	// 	Username:    user,
	// 	Channel:     channel,
	// 	IconEmoji:   iconEmoji,
	// 	Attachments: []slack.Attachment{},
	// 	LinkNames:   "1",
	// }
	err := slack.Send(webhookUrl, "", payload)
	if len(err) > 0 {
		return &err
	}
	return nil
}
