package handlers

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/dkeysil/nostrinbot/internal/messages"
	"github.com/dkeysil/nostrinbot/pkg/nostrutils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/nbd-wtf/go-nostr/sdk"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"go.uber.org/zap"
)

var (
	cacheTime      = 600
	titleSizeLimit = 150
	textSizeLimit  = 240
)

type SendNoteInlineQueryHandler struct {
	pool *nostr.SimplePool
}

func NewSendNoteInlineQueryHandler(pool *nostr.SimplePool) *SendNoteInlineQueryHandler {
	return &SendNoteInlineQueryHandler{
		pool: pool,
	}
}

func (h *SendNoteInlineQueryHandler) HandleInlineQuery(bot *tgbotapi.BotAPI, query *tgbotapi.InlineQuery) {
	event, err := h.getEvent(query.Query)
	if event == nil {
		zap.L().Debug("problem while querying event", zap.Error(err))
		return
	}

	metadata, err := h.getAuthorMetadata(event.PubKey)
	if err != nil {
		zap.L().Error("failed to get author metadata", zap.Error(err))
		return
	}

	title, message, err := h.prepareMessage(event, metadata)
	if err != nil {
		zap.L().Error("failed to prepare message", zap.Error(err))
		return
	}

	if _, err := bot.Send(prepareInlineAnswer(query.ID, event.ID, title, message)); err != nil {
		zap.L().Error("failed to answer inline query", zap.Error(err))
	}
}

func (h *SendNoteInlineQueryHandler) getEvent(input string) (*nostr.Event, error) {
	evPointer := sdk.InputToEventPointer(strings.TrimPrefix(input, "nostr:"))
	if evPointer == nil {
		return nil, fmt.Errorf("invalid input, input: %s", input)
	}

	evs := nostrutils.QueryPoolSync(context.Background(), h.pool, nostr.Filters{{
		IDs: []string{evPointer.ID},
	}})
	if len(evs) == 0 {
		return nil, fmt.Errorf("event not found, id: %s", evPointer.ID)
	}

	return evs[0], nil
}

func (h *SendNoteInlineQueryHandler) getEngagements(
	eventID string,
) (zapsSum int64, positiveReactions int64, repliesCount int64, repostsCount int64) {
	engagementEvs := nostrutils.QueryPoolSync(context.Background(), h.pool, nostr.Filters{{
		Kinds: []int{nostr.KindTextNote, nostr.KindRepost, nostr.KindReaction, nostr.KindZap},
		Tags: nostr.TagMap{
			"e": []string{eventID},
		},
	}})

	for _, ev := range engagementEvs {
		if ev.Kind == nostr.KindZap {
			bolt11, err := decodepay.Decodepay(ev.Tags.GetFirst([]string{"bolt11"}).Value())
			if err != nil {
				continue
			}
			zapsSum += bolt11.MSatoshi
		}

		if ev.Kind == nostr.KindReaction {
			for _, positiveReaction := range []string{"👍", "🤙", "+", "♥️"} {
				if ev.Content == positiveReaction {
					positiveReactions++
				}
			}
		}
	}

	repliesCount = countKind(engagementEvs, nostr.KindTextNote)
	repostsCount = countKind(engagementEvs, nostr.KindRepost)

	return zapsSum, positiveReactions, repliesCount, repostsCount
}

func (h *SendNoteInlineQueryHandler) getAuthorMetadata(pubKey string) (*nostr.ProfileMetadata, error) {
	evs := nostrutils.QueryPoolSync(context.Background(), h.pool, nostr.Filters{{
		Kinds:   []int{nostr.KindSetMetadata},
		Authors: []string{pubKey},
	}})
	if len(evs) == 0 {
		return nil, fmt.Errorf("author not found, id: %s", pubKey)
	}

	metadata, err := nostr.ParseMetadata(*evs[0])
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func (h *SendNoteInlineQueryHandler) prepareMessage(
	event *nostr.Event,
	metadata *nostr.ProfileMetadata,
) (title string, message string, _ error) {
	npub, err := nip19.EncodeNote(event.ID)
	if err != nil {
		return "", "", err
	}

	dateString := event.CreatedAt.Time().Format("3:04 PM · Jan 2, 2006")

	zapsSum, likesCount, repliesCount, repostsCount := h.getEngagements(event.ID)

	t := messages.NoteTemplate{
		DisplayName:  metadata.DisplayName,
		Nip05:        metadata.NIP05,
		Text:         TruncateString(event.Content, textSizeLimit),
		CreatedAt:    dateString,
		Link:         "https://nostter.com/" + npub,
		RepliesCount: int(repliesCount),
		RepostsCount: int(repostsCount),
		LikesCount:   int(likesCount),
		//  milliSat to Sat
		ZapsSum: zapsSum / 1000, //nolint:gomnd
	}

	if metadata.DisplayName == "" {
		t.DisplayName = metadata.Name
	}

	var buff bytes.Buffer
	// TODO: support media
	if err := messages.NoteTextTemplate.Execute(&buff, t); err != nil {
		return "", "", nil
	}

	title = TruncateString(t.DisplayName+": "+t.Text, titleSizeLimit)

	return title, buff.String(), nil
}

func prepareInlineAnswer(inlineQueryID, eventID, title, message string) tgbotapi.InlineConfig {
	result := tgbotapi.InlineQueryResultArticle{
		Type:  "article",
		ID:    eventID,
		Title: title,
		InputMessageContent: tgbotapi.InputTextMessageContent{
			Text:                  message,
			ParseMode:             "HTML",
			DisableWebPagePreview: true,
		},
	}

	return tgbotapi.InlineConfig{
		InlineQueryID: inlineQueryID,
		Results:       []interface{}{result},
		CacheTime:     cacheTime,
	}
}

func TruncateString(str string, length int) string {
	if length <= 0 {
		return ""
	}

	if utf8.RuneCountInString(str) < length {
		return str
	}

	return string([]rune(str)[:length]) + "..."
}

func countKind(evs []*nostr.Event, kind int) (count int64) {
	for _, ev := range evs {
		if ev.Kind == kind {
			count++
		}
	}
	return count
}
