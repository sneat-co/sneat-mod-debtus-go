package api4debtus

import (
	"context"
	"errors"
	"fmt"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-core-modules/auth/token4auth"
	"github.com/sneat-co/sneat-core-modules/common4all"
	"github.com/sneat-co/sneat-core-modules/userus/dal4userus"
	"github.com/sneat-co/sneat-core-modules/userus/dbo4userus"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-mod-debtus-go/debtstracker/analytics"
	"github.com/sneat-co/sneat-mod-debtus-go/debtstracker/dtdal"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/common4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/debtusbotconst"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/facade4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/facade4debtus/dto4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/general4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/models4debtus"
	"github.com/sneat-co/sneat-mod-debtus-go/facade2bots"
	"github.com/sneat-co/sneat-mod-debtus-go/invites"
	"github.com/sneat-co/sneat-translations/trans"
	"github.com/strongo/i18n"
	"github.com/strongo/logus"
	"github.com/strongo/strongoapp"
	"net/http"
	"strings"
	"time"
)

func NewReceiptTransferDto(ctx context.Context, transfer models4debtus.TransferEntry) dto4debtus.ApiReceiptTransferDto {
	creator := transfer.Data.Creator()
	transferDto := dto4debtus.ApiReceiptTransferDto{
		ID:             transfer.ID,
		From:           dto4debtus.NewContactDto(*transfer.Data.From()),
		To:             dto4debtus.NewContactDto(*transfer.Data.To()),
		Amount:         transfer.Data.GetAmount(),
		IsOutstanding:  transfer.Data.IsOutstanding,
		DtCreated:      transfer.Data.DtCreated,
		CreatorComment: creator.Comment,
		Creator: dto4debtus.ApiUserDto{ // TODO: Rename field - it can be not a creator in case of bill created by 3d party (paid by not by bill creator)
			ID:   creator.UserID,
			Name: creator.ContactName,
		},
	}
	// Set acknowledge info if presented
	if !transfer.Data.AcknowledgeTime.IsZero() {
		transferDto.Acknowledge = &dto4debtus.ApiAcknowledgeDto{
			Status:   transfer.Data.AcknowledgeStatus,
			UnixTime: transfer.Data.AcknowledgeTime.Unix(),
		}
	}
	if transferDto.From.Name == "" {
		logus.Warningf(ctx, "transferDto.From.Name is empty string")
	}

	if transferDto.To.Name == "" {
		logus.Warningf(ctx, "transferDto.To.Name is empty string")
	}
	return transferDto
}

func HandleGetReceipt(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	receiptID := common4all.GetStrID(ctx, w, r, "id")
	if receiptID == "" {
		return
	}

	receipt, err := dtdal.Receipt.GetReceiptByID(ctx, nil, receiptID)
	if common4all.HasError(ctx, w, err, models4debtus.ReceiptKind, receiptID, http.StatusBadRequest) {
		return
	}

	var transfer models4debtus.TransferEntry
	if err = facade.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) (err error) {
		transfer, err = facade4debtus.Transfers.GetTransferByID(ctx, tx, receipt.Data.TransferID)
		if common4all.HasError(ctx, w, err, models4debtus.TransfersCollection, receipt.Data.TransferID, http.StatusInternalServerError) {
			return
		}

		if err = facade4debtus.CheckTransferCreatorNameAndFixIfNeeded(ctx, tx, transfer); common4all.HasError(ctx, w, err, models4debtus.TransfersCollection, receipt.Data.TransferID, http.StatusInternalServerError) {
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
		env := dtdal.HttpAppHost.GetEnvironment(ctx, r)
		if env == strongoapp.UnknownEnv {
			w.WriteHeader(http.StatusBadRequest)
			logus.Warningf(ctx, "Unknown host")
		}
		sentTo, err = facade2bots.GetBotID(env, "telegram", debtusbotconst.DebtusBotProfileID, lang)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logus.Errorf(ctx, fmt.Errorf("failed to get bot settings by lang: %w", err).Error())
		}
	}

	creator := transfer.Data.Creator()

	logus.Debugf(ctx, "transfer.Creator(): %v", creator)

	receiptDto := dto4debtus.ApiReceiptDto{
		ID:       receiptID,
		Code:     "ToDoCODE",
		SentVia:  receipt.Data.SentVia,
		SentTo:   sentTo,
		Transfer: NewReceiptTransferDto(ctx, transfer),
	}

	common4all.JsonToResponse(ctx, w, &receiptDto)
}

//func transferContactToDto(transferContact models.TransferCounterpartyInfo) dto4debtus.ContactDto {
//	return dto4debtus.NewContactDto(transferContact)
//}

func HandleReceiptAccept(ctx context.Context, w http.ResponseWriter, _ *http.Request) {
	common4all.JsonToResponse(ctx, w, "ok")
}

func HandleReceiptDecline(ctx context.Context, w http.ResponseWriter, _ *http.Request) {
	common4all.JsonToResponse(ctx, w, "ok")
}

const RECEIPT_CHANNEL_DRAFT = "draft"

func HandleSendReceipt(ctx context.Context, w http.ResponseWriter, r *http.Request, authInfo token4auth.AuthInfo, user dbo4userus.UserEntry) {
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
		common4all.BadRequestMessage(ctx, w, "Not implemented yet")
		return
	default:
		common4all.BadRequestMessage(ctx, w, "Unsupported channel")
		return
	}

	toAddress := strings.TrimSpace(r.FormValue("to"))

	if toAddress == "" {
		common4all.BadRequestMessage(ctx, w, "'To' parameter is not provided")
		return
	}

	if len(toAddress) > 1024 {
		common4all.BadRequestMessage(ctx, w, "'To' parameter is too large")
		return
	}

	receipt, err := dtdal.Receipt.GetReceiptByID(ctx, nil, receiptID)

	if err != nil {
		var status int
		if dal.IsNotFound(err) {
			status = http.StatusBadRequest
		} else {
			status = http.StatusInternalServerError
		}
		common4all.ErrorAsJson(ctx, w, status, err)
		return
	}

	transfer, err := facade4debtus.Transfers.GetTransferByID(ctx, nil, receipt.Data.TransferID)
	if err != nil {
		common4all.ErrorAsJson(ctx, w, http.StatusInternalServerError, err)
		return
	}

	if transfer.Data.CreatorUserID != authInfo.UserID && transfer.Data.Counterparty().UserID != authInfo.UserID {
		common4all.ErrorAsJson(ctx, w, http.StatusUnauthorized, errors.New("this transfer does not belong to the current user"))
		return
	}

	locale := i18n.GetLocaleByCode5(user.Data.GetPreferredLocale()) // TODO: Get language from request
	translator := i18n.NewSingleMapTranslator(locale, nil /*anybot.TheAppContext.GetTranslator(ctx)*/)

	if _, err = invites.SendReceiptByEmail(ctx, translator, receipt, user.Data.GetFullName(), transfer.Data.Counterparty().ContactName, toAddress); err != nil {
		logus.Errorf(ctx, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		err = fmt.Errorf("failed to send receipt by email: %w", err)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	analytics.ReceiptSentFromApi(ctx, r, authInfo.UserID, locale.Code5, "api4debtus", "email")

	if _, _, err = updateReceiptAndTransferOnSent(ctx, receiptID, channel, toAddress, locale.Code5); err != nil {
		common4all.ErrorAsJson(ctx, w, http.StatusInternalServerError, err)
		return
	}
}

func updateReceiptAndTransferOnSent(ctx context.Context, receiptID string, channel, sentTo, lang string) (receipt models4debtus.ReceiptEntry, transfer models4debtus.TransferEntry, err error) {

	err = facade.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if receipt, err = dtdal.Receipt.GetReceiptByID(ctx, tx, receiptID); err != nil {
			return err
		}
		if receipt.Data.SentVia == RECEIPT_CHANNEL_DRAFT {
			if transfer, err = facade4debtus.Transfers.GetTransferByID(ctx, tx, receipt.Data.TransferID); err != nil {
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
			if err = tx.SetMulti(ctx, []dal.Record{receipt.Record, transfer.Record}); err != nil {
				err = fmt.Errorf("failed to save receipt & transfer: %w", err)
			}
		} else if receipt.Data.SentVia == channel {
			logus.Infof(ctx, "ReceiptEntry already has channel '%s'", channel)
		} else {
			logus.Warningf(ctx, "An attempt to set receipt channel to '%s' when it's alreay '%s'", channel, receipt.Data.SentVia)
		}

		return err
	})
	return
}

func HandleSetReceiptChannel(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid form data"))
		return
	}

	receiptID := r.FormValue("receipt")
	if receiptID == "" {
		err := fmt.Errorf("missing receipt parameter")
		logus.Debugf(ctx, err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	channel, err := getReceiptChannel(r)
	if err != nil {
		if errors.Is(err, errUnknownChannel) {
			m := "Unknown channel: " + channel
			logus.Debugf(ctx, m)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(m))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
	}

	logus.Debugf(ctx, "HandleSetReceiptChannel(receiptID=%s, channel=%s)", receiptID, channel)
	if channel == RECEIPT_CHANNEL_DRAFT {
		m := fmt.Sprintf("Status '%s' is not supported in this method", RECEIPT_CHANNEL_DRAFT)
		logus.Warningf(ctx, m)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(m))
	}

	if _, _, err = updateReceiptAndTransferOnSent(ctx, receiptID, channel, "", ""); err != nil {
		if dal.IsNotFound(err) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("ReceiptEntry not found by ContactID"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
		return
	}
	logus.Infof(ctx, "Done")
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

func HandleCreateReceipt(ctx context.Context, w http.ResponseWriter, r *http.Request, authInfo token4auth.AuthInfo) {
	if err := r.ParseForm(); err != nil {
		logus.Debugf(ctx, "HandleCreateReceipt() => Invalid form data: "+err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid form data"))
		return
	}
	//logus.Debugf(ctx, "HandleCreateReceipt() => r.Form: %+v", r.Form)
	transferID := r.FormValue("transfer")
	if transferID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Missing transfer parameter"))
		return
	}
	transfer, err := facade4debtus.Transfers.GetTransferByID(ctx, nil, transferID)
	if err != nil {
		if dal.IsNotFound(err) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			logus.Errorf(ctx, err.Error())
		}
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	var user dbo4userus.UserEntry
	if user, err = dal4userus.GetUserByID(ctx, nil, authInfo.UserID); err != nil {
		common4all.ErrorAsJson(ctx, w, http.StatusInternalServerError, err)
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

	lang := user.Data.PreferredLocale
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
	receiptData := models4debtus.NewReceiptEntity(creatorUserID, transferID, transfer.Data.Counterparty().UserID, lang, channel, "", general4debtus.CreatedOn{
		CreatedOnPlatform: "api4debtus", // TODO: Replace with actual, pass from client
		CreatedOnID:       r.Host,
	})

	receipt, err := dtdal.Receipt.CreateReceipt(ctx, receiptData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logus.Errorf(ctx, err.Error())
		return
	}

	//user, err = dal4userus.GetUserByID(ctx, transfer.CreatorUserID)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	err = errors.Wrapf(err, "failed to get user by CreatorUserID=%v", transfer.CreatorUserID)
	//	w.Write([]byte(err.Error()))
	//	logus.Warningf(ctx, err.Error())
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
		translator := i18n.NewSingleMapTranslator(locale, nil /*anybot.TheAppContext.GetTranslator(ctx)*/)
		//ec := strongoapp.NewExecutionContext(ctx, translator)

		logus.Debugf(ctx, "r.Host: %s", r.Host)

		templateParams := struct {
			ReceiptURL string
		}{
			ReceiptURL: common4debtus.GetReceiptUrl(receipt.ID, r.Host),
		}
		messageToSend, err = common4all.TextTemplates.RenderTemplate(ctx, translator, trans.MESSENGER_RECEIPT_TEXT, templateParams) // TODO: Consider using just ExecutionContext instead of (ctx, translator)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			err = fmt.Errorf("failed to render message template: %w", err)
			logus.Errorf(ctx, err.Error())
			_, _ = w.Write([]byte(err.Error()))
		}
	}

	common4all.JsonToResponse(ctx, w, struct {
		ID   string
		Link string
		Text string
	}{
		receipt.ID,
		// TODO: It seems wrong to use request host!
		//anybot.GetReceiptUrlForUser(receipt, receiptData.CreatorUserID, receiptData.CreatedOnPlatform, receiptData.CreatedOnID)
		fmt.Sprintf("https://%s/receipt?id=%s&t=%s", r.Host, receipt.ID, time.Now().Format("20060102-150405")),
		messageToSend,
	})
}
