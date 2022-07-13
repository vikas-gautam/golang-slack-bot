package main

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

func main() {
	//load env variables from .env files

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("err")
	}

	token := os.Getenv("SLACK_AUTH_TOKEN")
	channelID := os.Getenv("SLACK_CHANNEL_ID")

	//create a new client to slack by giving token
	//set debug to true while developing
	client := slack.New(token, slack.OptionDebug(true))

	attachment := slack.Attachment{
		Pretext: "Super Bot Message",
		Text:    "chitti reloaded",
		// Color Styles the Text, making it possible to have like Warnings etc.
		Color: "#36a64f",
		// Fields are Optional extra data!
		Fields: []slack.AttachmentField{
			{
				Title: "Date",
				Value: time.Now().String(),
			},
		},
	}

	_, timestamp, err := client.PostMessage(
		channelID,
		// uncomment the item below to add a extra Header to the message, try it out :)
		slack.MsgOptionText("New message from chitti-robot2.0 - han bhyi pandey ji kya chl rha hai", false),
		slack.MsgOptionAttachments(attachment),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Message sent at %s", timestamp)

}
