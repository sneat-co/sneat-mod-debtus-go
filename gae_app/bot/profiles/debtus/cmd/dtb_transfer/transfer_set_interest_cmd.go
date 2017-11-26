package dtb_transfer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"bitbucket.com/asterus/debtstracker-server/gae_app/debtstracker/models"
	"github.com/DebtsTracker/translations/trans"
	"github.com/pkg/errors"
	"github.com/strongo/bots-framework/core"
	"github.com/strongo/decimal"
)

var reInterest = regexp.MustCompile(`^\s*(?P<percent>\d+(?:[\.,]\d+)?)%?(?:/(?P<period>\d+|w(?:eek)?|y(?:ear)?|m(?:onth)?))?(?:/(?P<minimum>\d+))?(?:/(?P<grace>\d+))?(?::\s*(?P<comment>.+?))?\s*$`)

func interestAction(whc bots.WebhookContext, nextAction bots.CommandAction) (m bots.MessageFromBot, err error) {
	mt := whc.Input().(bots.WebhookTextMessage).Text()

	if matches := reInterest.FindStringSubmatch(mt); len(matches) > 0 {
		chatEntity := whc.ChatEntity()

		var data models.TransferInterest

		for i, name := range reInterest.SubexpNames() {
			v := matches[i]
			switch name {
			case "percent":
				v = strings.Replace(v, ",", ".", 1)
				if data.InterestPercent, err = decimal.ParseDecimal64p2(v); err != nil {
					return
				}
			case "period":
				if v == "" {
					m.Text = whc.Translate(trans.MESSAGE_TEXT_INTEREST_PLEASE_SPECIFY_PERIOD)
					return
				}
				switch v[0] {
				case "w"[0]:
					data.InterestPeriod = models.InterestRatePeriodWeekly
				case "m"[0]:
					data.InterestPeriod = models.InterestRatePeriodMonthly
				case "y"[0]:
					data.InterestPeriod = models.InterestRatePeriodYearly
				default:
					if data.InterestPeriod, err = strconv.Atoi(v); err != nil {
						return
					}
				}
			case "minimum":
				if v != "" {
					if data.InterestMinimumPeriod, err = strconv.Atoi(v); err != nil {
						return
					}
				}
			case "grace":
				if v != "" {
					if data.InterestGracePeriod, err = strconv.Atoi(v); err != nil {
						return
					}
				}
			case "comment":
				chatEntity.AddWizardParam(TRANSFER_WIZARD_PARAM_COMMENT, v)
			}
		}
		chatEntity.AddWizardParam(TRANSFER_WIZARD_PARAM_INTEREST, fmt.Sprintf("%v/%v/%v/%v/%v",
			models.InterestPercentSimple, data.InterestPercent, data.InterestPeriod, data.InterestMinimumPeriod, data.InterestGracePeriod),
		)

		return nextAction(whc)
	}

	return
}

const TRANSFER_WIZARD_PARAM_INTEREST = "interest"

func getInterestData(s string) (data models.TransferInterest, err error) {
	v := strings.Split(s, "/")
	switch v[0] {
	case models.InterestPercentSimple:
	case models.InterestPercentCompound:
	default:
		err = errors.New("unknown interest type: " + v[0])
		return
	}
	data.InterestType = models.InterestPercentType(v[0])
	if data.InterestPercent, err = decimal.ParseDecimal64p2(v[1]); err != nil {
		return
	}
	if data.InterestPeriod, err = strconv.Atoi(v[2]); err != nil {
		return
	}
	if data.InterestMinimumPeriod, err = strconv.Atoi(v[3]); err != nil {
		return
	}
	if data.InterestGracePeriod, err = strconv.Atoi(v[4]); err != nil {
		return
	}
	return
}
