package micha

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"
)

const (
	API_URL = "https://api.telegram.org/bot%s/%s"
)

type ApiResponse struct {
	Ok          bool            `json:"ok"`
	ErrorCode   int             `json:"error_code"`
	Description string          `json:"description"`
	Result      json.RawMessage `json:"result"`
}

type Bot struct {
	token   string
	Me      User
	Timeout time.Duration
	Updates chan Update
}

// Create new bot instance
func NewBot(token string) (*Bot, error) {
	bot := Bot{
		token:   token,
		Timeout: 25 * time.Second,
		Updates: make(chan Update),
	}

	if me, err := bot.GetMe(); err != nil {
		return nil, err
	} else {
		bot.Me = *me
		return &bot, nil
	}
}

// Build url for API method
func (bot *Bot) buildUrl(method string) string {
	return fmt.Sprintf(API_URL, bot.token, method)
}

// Decode response body to ApiResponse object
func (bot *Bot) decodeResponse(data []byte, target interface{}) error {
	apiResponse := new(ApiResponse)
	if err := json.Unmarshal(data, apiResponse); err != nil {
		return fmt.Errorf("Decode response error (%s)", err.Error())
	}

	if !apiResponse.Ok {
		return fmt.Errorf("%d: %s", apiResponse.ErrorCode, apiResponse.Description)
	}
	if err := json.Unmarshal(apiResponse.Result, target); err != nil {
		return fmt.Errorf("Decode result error (%s)", err.Error())
	} else {
		return nil
	}
}

// Make GET request to Telegram API
func (bot *Bot) get(method string, params url.Values, target interface{}) error {
	response, err := get(bot.buildUrl(method) + "?" + params.Encode())
	if err != nil {
		return err
	} else {
		return bot.decodeResponse(response, target)
	}
}

// Make POST request to Telegram API
func (bot *Bot) post(method string, data, target interface{}) error {
	response, err := post(bot.buildUrl(method), data)
	if err != nil {
		return err
	} else {
		return bot.decodeResponse(response, target)
	}
}

// Use this method to receive incoming updates using long polling.
// An Array of Update objects is returned.
func (bot *Bot) getUpdates(offset uint64) ([]Update, error) {
	params := url.Values{
		"offset":  {fmt.Sprintf("%d", offset)},
		"timeout": {fmt.Sprintf("%d", bot.Timeout/time.Second)},
	}

	updates := []Update{}
	err := bot.get("getUpdates", params, &updates)

	return updates, err
}

// A simple method for testing your bot's auth token.
// Returns basic information about the bot in form of a User object.
func (bot *Bot) GetMe() (*User, error) {
	me := new(User)
	err := bot.get("getMe", url.Values{}, me)

	return me, err
}

// Use this method to send text messages. On success, the sent Message is returned.
func (bot *Bot) SendMessage(chatId int64, text string, options *SendMessageOptions) (*Message, error) {
	params := SendMessageParams{
		ChatId: chatId,
		Text:   text,
	}
	if options != nil {
		params.SendMessageOptions = *options
	}

	message := new(Message)
	err := bot.post("sendMessage", params, message)

	return message, err
}

// Use this method to edit text messages sent by the bot or via the bot (for inline bots).
// On success, the edited Message is returned.
func (bot *Bot) EditMessageText(chatId, messageId int64, inlineMessageId, text string, options *EditMessageTextOptions) (*Message, error) {
	params := EditMessageTextParams{
		ChatId:          chatId,
		MessageId:       messageId,
		InlineMessageId: inlineMessageId,
		Text:            text,
	}
	if options != nil {
		params.EditMessageTextOptions = *options
	}

	message := new(Message)
	err := bot.post("editMessageText", params, message)

	return message, err
}

// Use this method to edit captions of messages sent by the bot or via the bot (for inline bots).
// On success, the edited Message is returned.
func (bot *Bot) EditMessageCaption(chatId, messageId int64, inlineMessageId string, options *EditMessageCationOptions) (*Message, error) {
	params := EditMessageCationParams{
		ChatId:          chatId,
		MessageId:       messageId,
		InlineMessageId: inlineMessageId,
	}
	if options != nil {
		params.EditMessageCationOptions = *options
	}

	message := new(Message)
	err := bot.post("editMessageCaption", params, message)

	return message, err
}

// Use this method to edit only the reply markup of messages sent by the bot or via the bot (for inline bots).
// On success, the edited Message is returned.
func (bot *Bot) EditMessageReplyMarkup(chatId, messageId int64, inlineMessageId string, replyMarkup *ReplyMarkup) (*Message, error) {
	params := EditMessageReplyMarkupParams{
		ChatId:          chatId,
		MessageId:       messageId,
		InlineMessageId: inlineMessageId,
		ReplyMarkup:     replyMarkup,
	}

	message := new(Message)
	err := bot.post("editMessageReplyMarkup", params, message)

	return message, err
}

// Start getting updates
func (bot *Bot) Start() {
	offset := uint64(0)

	for {
		updates, err := bot.getUpdates(offset + 1)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		for _, update := range updates {
			bot.Updates <- update

			offset = update.UpdateId
		}
	}
}
