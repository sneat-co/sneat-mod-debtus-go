package api

import (
	"errors"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/debtstracker-translations/trans"
	"github.com/strongo/i18n"
	"net/http"
	"strings"
	"time"

	"context"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/bot"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/bot/platforms/tgbots"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/analytics"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/api/dto"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/auth"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/common"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/dtdal"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/facade"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/models"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/general"
	"github.com/sneat-co/sneat-mod-debtus-go/gae_app/invites"
	"github.com/strongo/app"
	"github.com/strongo/log"
)

func NewReceiptTransferDto(c context.Context, transfer models.Transfer) dto.ApiReceiptTransferDto {
	creator := transfer.Data.Creator()
	transferDto := dto.ApiReceiptTransferDto{
		ID:             transfer.ID,
		From:           dto.NewContactDto(*transfer.Data.From()),
		To:             dto.NewContactDto(*transfer.Data.To()),
		Amount:         transfer.Data.GetAmount(),
		IsOutstanding:  transfer.Data.IsOutstanding,
		DtCreated:      transfer.Data.DtCreated,
		CreatorComment: creator.Comment,
		Creator: dto.ApiUserDto{ // TODO: Rename field - it can be not a creator in case of bill created by 3d party (paid by not by bill creator)
			ID:   creator.UserID,
			Name: creator.ContactName,
		},
	}
	// Set acknowledge info if presented
	if !transfer.Data.AcknowledgeTime.IsZero() {
		transferDto.Acknowledge = &dto.ApiAcknowledgeDto{
			Status:   transfer.Data.AcknowledgeStatus,
			UnixTime: transfer.Data.AcknowledgeTime.Unix(),
		}
	}
	if transferDto.From.Name == "" {
		log.Warningf(c, "transferDto.From.Name is empty string")
	}

	if transferDto.To.Name == "" {
		log.Warningf(c, "transferDto.To.Name is empty string")
	}
	return transferDto
}

func handleGetReceipt(c context.Context, w http.ResponseWriter, r *http.Request) {
	receiptID := getStrID(c, w, r, "id")
	if receiptID == "" {
		return
	}

	receipt, err := dtdal.Receipt.GetReceiptByID(c, nil, receiptID)
	if hasError(c, w, err, models.ReceiptKind, receiptID, http.StatusBadRequest) {
		return
	}

	var db dal.DB
	if db, err = facade.GetDatabase(c); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	var transfer models.Transfer
	if err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) (err error) {
		transfer, err = facade.Transfers.GetTransferByID(c, tx, receipt.Data.TransferID)
		if hasError(c, w, err, models.TransferKind, receipt.Data.TransferID, http.StatusInternalServerError) {
			return
		}

		if err = facade.CheckTransferCreatorNameAndFixIfNeeded(c, tx, transfer); hasError(c, w, err, models.TransferKind, receipt.Data.TransferID, http.StatusInternalServerError) {
			return
		}
		return nil
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	sentTo := receipt.Data.SentTo
	if receipt.Data.SentVia == "telegram" {
		lang := r.URL.Query().Get("lang")
		if lang == "" {
			lang = receipt.Data.Lang
		}
		env := dtdal.HttpAppHost.GetEnvironment(c, r)
		if env == strongo.EnvUnknown {
			w.WriteHeader(http.StatusBadRequest)
			log.Warningf(c, "Unknown host")
		}
		botSettings, err := tgbots.GetBotSettingsByLang(env, bot.ProfileDebtus, lang)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Errorf(c, fmt.Errorf("failed to get bot settings by lang: %w", err).Error())
		}
		sentTo = botSettings.Code
	}

	creator := transfer.Data.Creator()

	log.Debugf(c, "transfer.Creator(): %v", creator)

	receiptDto := dto.ApiReceiptDto{
		ID:       receiptID,
		Code:     "ToDoCODE",
		SentVia:  receipt.Data.SentVia,
		SentTo:   sentTo,
		Transfer: NewReceiptTransferDto(c, transfer),
	}

	jsonToResponse(c, w, &receiptDto)
}

//func transferContactToDto(transferContact models.TransferCounterpartyInfo) dto.ContactDto {
//	return dto.NewContactDto(transferContact)
//}

func handleReceiptAccept(c context.Context, w http.ResponseWriter, r *http.Request) {
	jsonToResponse(c, w, "ok")
}

func handleReceiptDecline(c context.Context, w http.ResponseWriter, r *http.Request) {
	jsonToResponse(c, w, "ok")
}

const RECEIPT_CHANNEL_DRAFT = "draft"

func handleSendReceipt(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo, user models.AppUser) {
	w.Header().Add("Access-Control-Allow-Origin", "*")

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid form data"))
		return
	}

	receiptID := r.FormValue("receipt")
	if receiptID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Missing receipt parameter"))
		return
	}

	channel := r.FormValue("by")

	switch channel {
	case "email":
	case "sms":
		BadRequestMessage(c, w, "Not implemented yet")
		return
	default:
		BadRequestMessage(c, w, "Unsupported channel")
		return
	}

	toAddress := strings.TrimSpace(r.FormValue("to"))

	if toAddress == "" {
		BadRequestMessage(c, w, "'To' parameter is not provided")
		return
	}

	if len(toAddress) > 1024 {
		BadRequestMessage(c, w, "'To' parameter is too large")
		return
	}

	receipt, err := dtdal.Receipt.GetReceiptByID(c, nil, receiptID)

	if err != nil {
		var status int
		if dal.IsNotFound(err) {
			status = http.StatusBadRequest
		} else {
			status = http.StatusInternalServerError
		}
		ErrorAsJson(c, w, status, err)
		return
	}

	transfer, err := facade.Transfers.GetTransferByID(c, nil, receipt.Data.TransferID)
	if err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}

	if transfer.Data.CreatorUserID != authInfo.UserID && transfer.Data.Counterparty().UserID != authInfo.UserID {
		ErrorAsJson(c, w, http.StatusUnauthorized, errors.New("this transfer does not belong to the current user"))
		return
	}

	locale := i18n.GetLocaleByCode5(user.Data.GetPreferredLocale()) // TODO: Get language from request
	translator := i18n.NewSingleMapTranslator(locale, common.TheAppContext.GetTranslator(c))

	if _, err = invites.SendReceiptByEmail(c, translator, receipt, user.Data.FullName(), transfer.Data.Counterparty().ContactName, toAddress); err != nil {
		log.Errorf(c, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		err = fmt.Errorf("failed to send receipt by email: %w", err)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	analytics.ReceiptSentFromApi(c, r, authInfo.UserID, locale.Code5, "api", "email")

	if _, _, err = updateReceiptAndTransferOnSent(c, receiptID, channel, toAddress, locale.Code5); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
}

func updateReceiptAndTransferOnSent(c context.Context, receiptID string, channel, sentTo, lang string) (receipt models.Receipt, transfer models.Transfer, err error) {
	var db dal.DB
	if db, err = facade.GetDatabase(c); err != nil {
		return
	}

	err = db.RunReadwriteTransaction(c, func(c context.Context, tx dal.ReadwriteTransaction) error {
		if receipt, err = dtdal.Receipt.GetReceiptByID(c, tx, receiptID); err != nil {
			return err
		}
		if receipt.Data.SentVia == RECEIPT_CHANNEL_DRAFT {
			if transfer, err = facade.Transfers.GetTransferByID(c, tx, receipt.Data.TransferID); err != nil {
				return err
			}
			receipt.Data.DtSent = time.Now()
			receipt.Data.SentVia = channel
			receipt.Data.SentTo = sentTo
			receipt.Data.Lang = lang
			transfer.Data.ReceiptsSentCount += 1
			transferHasThisReceiptID := false
			for _, rID := range transfer.Data.ReceiptIDs {
				if rID == receiptID {
					transferHasThisReceiptID = true
					break
				}
			}
			if !transferHasThisReceiptID {
				transfer.Data.ReceiptIDs = append(transfer.Data.ReceiptIDs, receiptID)
			}
			if err = tx.SetMulti(c, []dal.Record{receipt.Record, transfer.Record}); err != nil {
				err = fmt.Errorf("failed to save receipt & transfer: %w", err)
			}
		} else if receipt.Data.SentVia == channel {
			log.Infof(c, "Receipt already has channel '%v'", channel)
		} else {
			log.Warningf(c, "An attempt to set receipt channel to '%v' when it's alreay '%v'", channel, receipt.Data.SentVia)
		}

		return err
	})
	return
}

func handleSetReceiptChannel(c context.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid form data"))
		return
	}

	receiptID := r.FormValue("receipt")
	if receiptID == "" {
		err := fmt.Errorf("missing receipt parameter")
		log.Debugf(c, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	channel, err := getReceiptChannel(r)
	if err != nil {
		if errors.Is(err, errUnknownChannel) {
			m := "Unknown channel: " + channel
			log.Debugf(c, m)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(m))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
	}

	log.Debugf(c, "handleSetReceiptChannel(receiptID=%v, channel=%v)", receiptID, channel)
	if channel == RECEIPT_CHANNEL_DRAFT {
		m := fmt.Sprintf("Status '%v' is not supported in this method", RECEIPT_CHANNEL_DRAFT)
		log.Warningf(c, m)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(m))
	}

	if _, _, err = updateReceiptAndTransferOnSent(c, receiptID, channel, "", ""); err != nil {
		if dal.IsNotFound(err) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Receipt not found by ID"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
		return
	}
	log.Infof(c, "Done")
	_, _ = w.Write([]byte("ok"))
}

var errUnknownChannel = errors.New("Unknown channel")

func getReceiptChannel(r *http.Request) (channel string, err error) {
	channel = r.FormValue("channel")
	switch channel {
	case RECEIPT_CHANNEL_DRAFT:
	case "fbm":
	case "vk":
	case "viber":
	case "whatsapp":
	case "line":
	case "telegram":
	default:
		err = errUnknownChannel
	}
	return
}

func handleCreateReceipt(c context.Context, w http.ResponseWriter, r *http.Request, authInfo auth.AuthInfo) {
	if err := r.ParseForm(); err != nil {
		log.Debugf(c, "handleCreateReceipt() => Invalid form data: "+err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid form data"))
		return
	}
	log.Debugf(c, "handleCreateReceipt() => r.Form: %v", r.Form)
	transferID := r.FormValue("transfer")
	if transferID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Missing transfer parameter"))
		return
	}
	transfer, err := facade.Transfers.GetTransferByID(c, nil, transferID)
	if err != nil {
		if dal.IsNotFound(err) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Errorf(c, err.Error())
		}
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	var user models.AppUser
	if user, err = facade.User.GetUserByID(c, nil, authInfo.UserID); err != nil {
		ErrorAsJson(c, w, http.StatusInternalServerError, err)
		return
	}
	channel, err := getReceiptChannel(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if err == errUnknownChannel {
			_, _ = w.Write([]byte("Unknown channel: " + channel))
		}
		return
	}
	creatorUserID := transfer.Data.CreatorUserID // TODO: Get from session?

	lang := user.Data.PreferredLanguage
	if lang == "" {
		if acceptLanguage := r.Header.Get("Accept-Language"); acceptLanguage != "" {
			for _, set1 := range strings.Split(acceptLanguage, ";") {
				for _, al := range strings.Split(set1, ",") {
					switch len(al) {
					case 5:
						if _, ok := trans.SupportedLocalesByCode5[strings.ToLower(al[:2])+"-"+strings.ToUpper(al[4:])]; ok {
							lang = al
							goto langSet
						}
					case 2:
						al = strings.ToLower(al)
						for localeCode := range trans.SupportedLocalesByCode5 {
							if strings.HasPrefix(localeCode, al) {
								lang = localeCode
								goto langSet
							}
						}
					}
				}
			}
		langSet:
		}
	}
	if lang == "" {
		lang = i18n.LocaleCodeEnUS
	}
	receiptData := models.NewReceiptEntity(creatorUserID, transferID, transfer.Data.Counterparty().UserID, lang, channel, "", general.CreatedOn{
		CreatedOnPlatform: "api", // TODO: Replace with actual, pass from client
		CreatedOnID:       r.Host,
	})

	receipt, err := dtdal.Receipt.CreateReceipt(c, receiptData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf(c, err.Error())
		return
	}

	//user, err = facade.User.GetUserByID(c, transfer.CreatorUserID)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	err = errors.Wrapf(err, "Failed to get user by ID=%v", transfer.CreatorUserID)
	//	w.Write([]byte(err.Error()))
	//	log.Warningf(c, err.Error())
	//	return
	//}
	var messageToSend string

	if channel == "telegram" {
		tgBotID := transfer.Data.Creator().TgBotID
		if tgBotID == "" {
			if strings.Contains(r.URL.Host, "dev") {
				tgBotID = "DebtsTrackerDev1Bot"
			} else {
				tgBotID = "DebtsTrackerBot"
			}
		}
		messageToSend = fmt.Sprintf("https://telegram.me/%s?start=send-receipt_%s", tgBotID, receipt.ID) // TODO:
	} else {
		locale := i18n.GetLocaleByCode5(user.Data.GetPreferredLocale())
		translator := i18n.NewSingleMapTranslator(locale, common.TheAppContext.GetTranslator(c))
		//ec := strongo.NewExecutionContext(c, translator)

		log.Debugf(c, "r.Host: %v", r.Host)

		templateParams := struct {
			ReceiptURL string
		}{
			ReceiptURL: common.GetReceiptUrl(receipt.ID, r.Host),
		}
		messageToSend, err = common.TextTemplates.RenderTemplate(c, translator, trans.MESSENGER_RECEIPT_TEXT, templateParams) // TODO: Consider using just ExecutionContext instead of (c, translator)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			err = fmt.Errorf("failed to render message template: %w", err)
			log.Errorf(c, err.Error())
			_, _ = w.Write([]byte(err.Error()))
		}
	}

	jsonToResponse(c, w, struct {
		ID   string
		Link string
		Text string
	}{
		receipt.ID,
		// TODO: It seems wrong to use request host!
		//common.GetReceiptUrlForUser(receipt, receiptData.CreatorUserID, receiptData.CreatedOnPlatform, receiptData.CreatedOnID)
		fmt.Sprintf("https://%s/receipt?id=%s&t=%s", r.Host, receipt.ID, time.Now().Format("20060102-150405")),
		messageToSend,
	})
}
