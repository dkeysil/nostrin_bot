package main

import (
	"context"

	"github.com/dkeysil/nostrinbot/internal/config"
	"github.com/dkeysil/nostrinbot/internal/handlers"
	"github.com/dkeysil/nostrinbot/pkg/telegram"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nbd-wtf/go-nostr"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	zap.ReplaceGlobals(log)

	pool := nostr.NewSimplePool(context.Background())
	for _, url := range cfg.Nostr.RelayURLs {
		_, err := pool.EnsureRelay(url)
		if err != nil {
			log.Warn("failed to add relay", zap.Error(err), zap.String("url", url))
		}
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBot.Token)
	if err != nil {
		log.Fatal("failed to create bot", zap.Error(err))
	}

	bot.Debug = true

	log.Info("authorized on account", zap.String("bot_name", bot.Self.UserName))

	h := handlers.NewSendNoteInlineQueryHandler(pool)

	router := telegram.NewRouter(bot)
	router.RegisterInlineQueryHandler(&telegram.InlineQueryHandler{
		Handler: h.HandleInlineQuery,
	})

	log.Info("starting bot")
	router.Serve()
}
