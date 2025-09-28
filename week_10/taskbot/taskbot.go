package main

import (
	"context"
	"io"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	WebhookURL string
	BotToken   string
)

type BotApi struct {
	Db     *Storage
	Router *Router
}

func NewBot() *BotApi {
	return &BotApi{
		Db:     NewStorage(),
		Router: NewRouter(),
	}
}

type UserID string
type UserTag string

func startTaskBot(ctx context.Context, httpListenAddr string) error {

	bot, err := tgbotapi.NewBotAPI(BotToken)

	if err != nil {
		return err
	}

	whCfg := tgbotapi.NewWebhook(WebhookURL)
	if _, err := bot.SetWebhook(whCfg); err != nil {
		log.Panic(err)
	}

	updates := bot.ListenForWebhook("/")

	go func() {
		log.Fatal(http.ListenAndServe(httpListenAddr, nil))
	}()

	log.Printf("Бот запущен как @%s", bot.Self.UserName)

	newBotApi := NewBot()

	for update := range updates {
		if update.Message != nil {
			ctxWithValue := context.WithValue(ctx, UserID("userId"), update.Message.From.ID)
			ctxWithValue = context.WithValue(ctxWithValue, UserTag("userTag"), update.Message.From.UserName)
			msgs, err := newBotApi.Router.SelectRoute(ctxWithValue, newBotApi.Db, update.Message.Text)
			if err != nil {
				log.Println(err.Error())
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
				bot.Send(msg)
				continue
			}

			for _, msg := range msgs {
				msgT := tgbotapi.NewMessage(msg.ID, msg.Msg)
				if _, err := bot.Send(msgT); err != nil && err != io.EOF {
					log.Println("Ошибка отправки:", err)
				}
			}
		}
	}

	return nil
}

func main() {
	err := startTaskBot(context.Background(), ":8081")
	if err != nil {
		log.Fatalln(err)
	}
}

// это заглушка чтобы импорт сохранился
func __dummy() {
	tgbotapi.APIEndpoint = "_dummy"
}
