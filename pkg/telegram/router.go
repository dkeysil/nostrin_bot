package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Router struct {
	bot *tgbotapi.BotAPI

	MessageHandlers     []Handler[*tgbotapi.Message]
	InlineQueryHandlers []Handler[*tgbotapi.InlineQuery]
}

func NewRouter(bot *tgbotapi.BotAPI) *Router {
	return &Router{
		bot:                 bot,
		MessageHandlers:     make([]Handler[*tgbotapi.Message], 0),
		InlineQueryHandlers: make([]Handler[*tgbotapi.InlineQuery], 0),
	}
}

func (r *Router) Serve() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := r.bot.GetUpdatesChan(u)

	for update := range updates {
		switch {
		case update.Message != nil:
			for _, handler := range r.MessageHandlers {
				if handler.Satisfy(update.Message) {
					go handler.Handle(r.bot, update.Message)
					continue
				}
			}
		case update.InlineQuery != nil:
			for _, handler := range r.InlineQueryHandlers {
				if handler.Satisfy(update.InlineQuery) {
					go handler.Handle(r.bot, update.InlineQuery)
					continue
				}
			}
		}
	}
}

func (r *Router) RegisterMessageHandler(handler Handler[*tgbotapi.Message]) {
	r.MessageHandlers = append(r.MessageHandlers, handler)
}

func (r *Router) RegisterInlineQueryHandler(handler Handler[*tgbotapi.InlineQuery]) {
	r.InlineQueryHandlers = append(r.InlineQueryHandlers, handler)
}

type MessageHandler struct {
	Filter struct {
		Text string
	}
	Handler func(bot *tgbotapi.BotAPI, update *tgbotapi.Message)
}

func (mh *MessageHandler) Satisfy(update *tgbotapi.Message) bool {
	if mh.Filter.Text != "" && mh.Filter.Text != update.Text {
		return false
	}

	return true
}

func (mh *MessageHandler) Handle(bot *tgbotapi.BotAPI, update *tgbotapi.Message) {
	mh.Handler(bot, update)
}

type InlineQueryHandler struct {
	Filter struct {
		Query    string
		ChatType string
	}
	Handler func(bot *tgbotapi.BotAPI, update *tgbotapi.InlineQuery)
}

func (iqh *InlineQueryHandler) Satisfy(update *tgbotapi.InlineQuery) bool {
	if iqh.Filter.Query != "" && iqh.Filter.Query != update.Query {
		return false
	}

	if iqh.Filter.ChatType != "" && iqh.Filter.ChatType != update.ChatType {
		return false
	}

	return true
}

func (iqh *InlineQueryHandler) Handle(bot *tgbotapi.BotAPI, update *tgbotapi.InlineQuery) {
	iqh.Handler(bot, update)
}

type Handler[U Update] interface {
	Satisfy(update U) bool
	Handle(bot *tgbotapi.BotAPI, update U)
}

type Update interface {
	*tgbotapi.Message | *tgbotapi.InlineQuery
}
