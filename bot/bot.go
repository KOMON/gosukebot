package bot

import (
	"fmt"
	"log"
	"os"

	"github.com/komon/gosukebot/handler"
	"github.com/nlopes/slack"
)

//Run is the main operation of the bot, we set up a new slack api conn
// and start receiving and sending messages
func Run() int {
	api := slack.New(slackToken)

	logger, err := loggerSetup()
	if err != nil {
		return 1
	}
	handler.Init()
	slack.SetLogger(logger)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for {
		event := <-rtm.IncomingEvents
		if event.Type == "message" {
			msg := event.Data.(*slack.MessageEvent)
			resp, err := handler.Handle(msg.Text)
			if err != nil {
				logger.Printf("message handle error: %v", err)
				rtm.SendMessage(rtm.NewOutgoingMessage(err.Error(), msg.Channel))
				if err.Error() == "shutdown" {
					break
				}
			}
			rtm.SendMessage(rtm.NewOutgoingMessage(resp, msg.Channel))
		}
	}

	return 0
}

func loggerSetup() (*log.Logger, error) {
	f, err := os.OpenFile("jojolog", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		fmt.Printf("error opening logfile: %v", err)
		return nil, err
	}

	defer f.Close()
	return log.New(f, "jojobot: ", log.Lshortfile|log.LstdFlags), nil
}
