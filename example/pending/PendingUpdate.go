package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	tdlib "github.com/frknkrc44/gotdlib/client"
)

func GetTdParameters() *tdlib.SetTdlibParametersRequest {
	return &tdlib.SetTdlibParametersRequest{
		UseTestDc:           false,
		DatabaseDirectory:   "./tdlib-db",
		FilesDirectory:      "./tdlib-files",
		UseFileDatabase:     true,
		UseChatInfoDatabase: true,
		UseMessageDatabase:  true,
		UseSecretChats:      false,
		ApiId:               132712,
		ApiHash:             "e82c07ad653399a37baca8d1e498e472",
		SystemLanguageCode:  "en",
		DeviceModel:         "HuskyNG",
		SystemVersion:       "3.0",
		ApplicationVersion:  "3.0",
	}
}

func main() {
	tdlib.SetLogLevel(0)
	tdlib.SetFilePath("./errors.txt")

	// Set pending update list
	tdlib.SetPendingUpdateType(&tdlib.UpdateNewMessage{})
	// Of coz, you can set more than one type, depending on your needs
	//tdlib.SetPendingUpdateType(&tdlib.UpdateNewMessage{}, &tdlib.UpdateMessageEdited{}, &tdlib.UpdateDeleteMessages{})

	botToken := "your_bot_token"
	authorizer := tdlib.BotAuthorizer(botToken)

	authorizer.TdlibParameters <- GetTdParameters()

	client, err := tdlib.NewClient(authorizer)
	if err != nil {
		log.Fatalf("NewClient error: %s", err)
	}

	// Handle SIGINT
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT)
	signal.Notify(ch, os.Interrupt, syscall.SIGKILL)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	signal.Notify(ch, os.Interrupt, syscall.SIGQUIT)
	signal.Notify(ch, os.Interrupt, syscall.SIGSEGV)
	go func() {
		<-ch
		client.Close()
	}()

	me, err := client.GetMe()
	if err != nil {
		log.Fatalf("GetMe error: %s", err)
	}

	log.Printf("%v connected", me.Usernames)

	listener := client.AddEventReceiver(&tdlib.UpdateNewMessage{}, 1000)

	defer listener.Close()
	for update := range listener.Updates {
		updateMsg := update.(*tdlib.UpdateNewMessage)
		chatId := updateMsg.Message.ChatId
		msgId := updateMsg.Message.Id

		var msgText string
		var msgEnt []*tdlib.TextEntity

		switch updateMsg.Message.Content.MessageContentType() {
		case "messageText":
			msgText = updateMsg.Message.Content.(*tdlib.MessageText).Text.Text
			msgEnt = updateMsg.Message.Content.(*tdlib.MessageText).Text.Entities

			cmd := tdlib.CheckCommand(msgText, msgEnt)
			switch cmd {
			case "/ping":
				text, _ := tdlib.ParseTextEntities(&tdlib.ParseTextEntitiesRequest{
					Text:      "<b>pong!</b>",
					ParseMode: &tdlib.TextParseModeHTML{},
				})
				m, err := client.SendMessage(&tdlib.SendMessageRequest{
					ChatId: chatId,
					ReplyTo: &tdlib.InputMessageReplyToMessage{
						MessageId: msgId,
					},
					InputMessageContent: &tdlib.InputMessageText{
						Text: text,
					},
				})
				if err != nil {
					continue
				}
				log.Printf("Message sent, ID: %d", m.Id)
			}
		}
	}
}
