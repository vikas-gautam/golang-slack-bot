package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func main() {
	//load env variables from .env files

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("err")
	}

	token := os.Getenv("SLACK_AUTH_TOKEN")
	//	channelID := os.Getenv("SLACK_CHANNEL_ID")
	appToken := os.Getenv("SLACK_APP_TOKEN")

	//create a new client to slack by giving token
	//set debug to true while developing
	client := slack.New(token, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))

	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(true),
		// Option to set a custom logger
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)
	// Create a context that can be used to cancel goroutine
	ctx, cancel := context.WithCancel(context.Background())
	// Make this cancel called properly in a real program , graceful shutdown etc
	defer cancel()
	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
		// Create a for loop that selects either the context cancellation or the events incomming
		for {
			select {
			// inscase context cancel is called exit the goroutine
			case <-ctx.Done():
				log.Println("Shutting down socketmode listener")
				return
			case event := <-socketClient.Events:
				// We have a new Events, let's type switch the event
				// Add more use cases here if you want to listen to other events.
				switch event.Type {
				// handle EventAPI events
				case socketmode.EventTypeEventsAPI:
					// The Event sent on the channel is not the same as the EventAPI events so we need to type cast it
					eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
					if !ok {
						log.Printf("Could not type cast the event to the EventsAPIEvent: %v\n", event)
						continue
					}
					// We need to send an Acknowledge to the slack server
					socketClient.Ack(*event.Request)
					// Now we have an Events API event, but this event type can in turn be many types, so we actually need another type switch
					// Now we have an Events API event, but this event type can in turn be many types, so we actually need another type switch
					err := handleEventMessage(eventsAPIEvent, client)
					if err != nil {
						// Replace with actual err handeling
						log.Fatal(err)
					}
				}

			}
		}
	}(ctx, client, socketClient)
	socketClient.Run()

	// attachment := slack.Attachment{
	// 	Pretext: "Super Bot Message",
	// 	Text:    "chitti reloaded",
	// 	// Color Styles the Text, making it possible to have like Warnings etc.
	// 	Color: "#36a64f",
	// 	// Fields are Optional extra data!
	// 	Fields: []slack.AttachmentField{
	// 		{
	// 			Title: "Date",
	// 			Value: time.Now().String(),
	// 		},
	// 	},
	// }

	// _, timestamp, err := client.PostMessage(
	// 	channelID,
	// 	// uncomment the item below to add a extra Header to the message, try it out :)
	// 	slack.MsgOptionText("New message from chitti-robot2.0 - han bhyi pandey ji kya chl rha hai", false),
	// 	slack.MsgOptionAttachments(attachment),
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("Message sent at %s", timestamp)

}

// handleEventMessage will take an event and handle it properly based on the type of event
func handleEventMessage(event slackevents.EventsAPIEvent, client *slack.Client) error {
	switch event.Type {
	// First we check if this is an CallbackEvent
	case slackevents.CallbackEvent:

		innerEvent := event.InnerEvent
		// Yet Another Type switch on the actual Data to see if its an AppMentionEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			// The application has been mentioned since this Event is a Mention event
			err := handleAppMentionEvent(ev, client)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("unsupported event type")
	}
	return nil
}

// handleAppMentionEvent is used to take care of the AppMentionEvent when the bot is mentioned
func handleAppMentionEvent(event *slackevents.AppMentionEvent, client *slack.Client) error {

	// Grab the user name based on the ID of the one who mentioned the bot
	user, err := client.GetUserInfo(event.User)
	if err != nil {
		return err
	}
	// Check if the user said Hello to the bot
	text := strings.ToLower(event.Text)

	// Create the attachment and assigned based on the message
	attachment := slack.Attachment{}
	attachment_leave := slack.Attachment{}
	attachment_final := slack.Attachment{}
	// Add Some default context like user who mentioned the bot
	attachment_leave.Fields = []slack.AttachmentField{
		{
			Title: "Leave format",
			Value: fmt.Sprintln("Reason:		\nno. of days:		\nDate From:	To Date:		\nDay:		\ncc:		\nAppliedOnPortal:		\n"),
		},
		{
			Title: "Applicant",
			Value: user.Name,
		},
	}
	if strings.Contains(text, "leave") {
		// Greet the user
		attachment_leave.Text = fmt.Sprintln("Hope you are doing well, Please send leave info in below format")
		attachment_leave.Pretext = "Greetings Opstrian" //heading
		attachment_leave.Color = "#4af030"
		attachment_final = attachment_leave
	} else {
		// Send a message to the user
		attachment.Text = fmt.Sprintf("How can I help you %s?", user.Name)
		attachment.Pretext = "How can I be of service"
		attachment.Color = "#3d3d3d"
		attachment_final = attachment
	}
	// Send the message to the channel
	// The Channel is available in the event message
	_, _, err = client.PostMessage(event.Channel, slack.MsgOptionAttachments(attachment_final))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	return nil
}
