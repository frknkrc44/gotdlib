package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	tdlib "github.com/c0re100/gotdlib/client"
)

func GetSenderId(sender tdlib.MessageSender) int64 {
	if sender.MessageSenderType() == "messageSenderUser" {
		return sender.(*tdlib.MessageSenderUser).UserId
	} else {
		return sender.(*tdlib.MessageSenderChat).ChatId
	}
}

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

	authorizer := tdlib.ClientAuthorizer()
	go tdlib.CliInteractor(authorizer)

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
		senderId := GetSenderId(updateMsg.Message.SenderId)
		msgId := updateMsg.Message.Id

		if senderId == me.Id {
			var msgText string
			var msgEnt []*tdlib.TextEntity

			switch updateMsg.Message.Content.MessageContentType() {
			case "messageText":
				msgText = updateMsg.Message.Content.(*tdlib.MessageText).Text.Text
				msgEnt = updateMsg.Message.Content.(*tdlib.MessageText).Text.Entities
			}

			cmd := tdlib.CheckCommand(msgText, msgEnt)
			switch cmd {
			case "/test":
				text, _ := tdlib.ParseTextEntities(&tdlib.ParseTextEntitiesRequest{
					Text:      "<b>Hi test user</b>",
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
			case "/repeat":
				m, err := client.SendMessage(&tdlib.SendMessageRequest{
					ChatId: chatId,
					ReplyTo: &tdlib.InputMessageReplyToMessage{
						MessageId: msgId,
					},
					InputMessageContent: &tdlib.InputMessageText{
						Text: &tdlib.FormattedText{Text: tdlib.CommandArgument(msgText)},
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
