package api

import (
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/platforms/telegram"
	"bitbucket.com/asterus/debtstracker-server/gae_app/bot/profiles/debtus/cmd/dtb_transfer"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/auth"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/dal"
	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"bitbucket.com/asterus/debtstracker-server/gae_app/gaestandard"
	"fmt"
	"github.com/pkg/errors"
	"github.com/strongo/log"
	"github.com/strongo/bots-api-telegram"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/bots-framework/platforms/telegram"
	"golang.org/x/net/context"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

func handleTgHelperCurrencySelected(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	selectedCurrency := r.FormValue("currency")
	if selectedCurrency == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing required parameter 'currency'"))
		return
	}
	if len(selectedCurrency) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong lengths of parameter 'currency'"))
		return
	}
	if strings.ToUpper(selectedCurrency) != selectedCurrency {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong currency code"))
		return
	}

	tgChatKeyID := r.Form.Get("tg-chat")
	if tgChatKeyID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing required parameter chat ID."))
		return
	}

	if !strings.Contains(tgChatKeyID, ":") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong foramt of Telegram chat ID parameter"))
		return
	}

	tgChatID, err := strconv.ParseInt(strings.Split(tgChatKeyID, ":")[1], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Telegram chat ID should be integer"))
		w.Write([]byte(err.Error()))
	}
	log.Debugf(c, "AppUserIntID: %v, tgChatKeyID: %v", authInfo.UserID, tgChatKeyID)

	errs := make(chan error, 2) // We use errors channel as sync pipe

	var user models.AppUser

	var userTask sync.WaitGroup

	userTask.Add(1)
	go func() {
		if err := dal.User.SetLastCurrency(c, authInfo.UserID, models.Currency(selectedCurrency)); err != nil {
			err = errors.WithMessage(err, "Failed to save user last currency")
		}
		userTask.Done()
		errs <- nil
	}()

	go func(currency string) {
		errs <- dal.TgChat.DoSomething(c, &userTask, currency, tgChatID, authInfo, user,
			func(tgChat telegram_bot.TelegramChatEntityBase) error {
				// TODO: This is some serious architecture sheet. Too sleepy to make it right, just make it working.
				return sendToTelegram(c, user, tgChatID, tgChat, &userTask, r)
			},
		)
	}(selectedCurrency)

	for i := range []int{1, 2} {
		if err := <-errs; err != nil {
			log.Errorf(c, "%v: %v", i, err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}

	log.Debugf(c, "Selected currency processed: %v", selectedCurrency)
}

// TODO: This is some serious architecture sheet. Too sleepy to make it right, just make it working.
func sendToTelegram(c context.Context, user models.AppUser, tgChatID int64, tgChat telegram_bot.TelegramChatEntityBase, userTask *sync.WaitGroup, r *http.Request) (err error) {
	telegramBots := telegram.Bots(gaestandard.GetEnvironment(c), nil)
	botSettings, ok := telegramBots.ByCode[tgChat.BotID]
	if !ok {
		return fmt.Errorf("Bot settings not found by tgChat.BotID=%v, out of %v items", tgChat.BotID, len(telegramBots.ByCode))
	}

	log.Debugf(c, "botSettings(%v : %v)", botSettings.Code, botSettings.Token)

	tgBotApi := tgbotapi.NewBotAPIWithClient(botSettings.Token, dal.HttpClient(c))
	tgBotApi.EnableDebug(c)

	userTask.Wait()

	whc := NewApiWebhookContext(
		r,
		user.AppUserEntity,
		user.ID,
		tgChatID,
		&tgChat,
	)

	var messageFromBot bots.MessageFromBot
	switch {
	case strings.Contains(tgChat.AwaitingReplyTo, "lending"):
		messageFromBot, err = dtb_transfer.AskLendingAmountCommand.Action(whc)
	case strings.Contains(tgChat.AwaitingReplyTo, "borrowing"):
		messageFromBot, err = dtb_transfer.AskBorrowingAmountCommand.Action(whc)
	default:
		return fmt.Errorf("tgChat.AwaitingReplyTo has unexpected value: %v", tgChat.AwaitingReplyTo)
	}
	if err != nil {
		return errors.Wrap(err, "Failed to create message from bot")
	}

	messageConfig := tgbotapi.NewMessage(tgChatID, messageFromBot.Text)
	messageConfig.ReplyMarkup = messageFromBot.Keyboard
	messageConfig.ParseMode = "HTML"

	if _, err = tgBotApi.Send(messageConfig); err != nil {
		return errors.Wrapf(err, "Failed to send message to Telegram chat=%v", tgChatID)
	}
	return nil
}
