// Code generated by hero.
// source: /Users/astec/go_workspace/src/bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/website/pages/inspector/contact-card.html
// DO NOT EDIT!
package inspector

import (
	"bytes"
	"fmt"
	"time"

	"github.com/shiyanhui/hero"
)

func heroContactCard(now time.Time, contact contactWithBalances, buffer *bytes.Buffer) {
	buffer.WriteString(`
<div class="card mr-2 mb-2">
    <div class=card-body>
        <h4 class=card-title>
            <!--a href="contact?id=`)
	hero.FormatInt(int64(contact.ID), buffer)
	buffer.WriteString(`">`)
	hero.EscapeHTML(contact.Data.FullName(), buffer)
	buffer.WriteString(`</a-->
            <span>`)
	hero.EscapeHTML(contact.Data.FullName(), buffer)
	buffer.WriteString(`</span>
        </h4>
        <table class="table">
            <thead>
            <tr>
                <th scope="col">Status</th>
                <th scope="col">Telegram #</th>
                <th scope="col" class=center>Linked</th>
                <th scope="col" class="d">Transfers</th>
            </tr>
            </thead>
            <tbody>
            <tr>
                <td>`)
	hero.EscapeHTML(contact.Data.Status, buffer)
	buffer.WriteString(`</td>

                `)
	if contact.Data.TelegramUserID == 0 {
		buffer.WriteString(`
                <td>no</td>
                `)
	} else {
		buffer.WriteString(`
                <td>yes</td>
                `)
	}
	if contact.Data.CounterpartyCounterpartyID == 0 {
		buffer.WriteString(`
                <td class=center>no</td>
                `)
	} else {
		buffer.WriteString(`
                <td class=center><a href=contact?id=`)
		hero.FormatInt(int64(contact.Data.CounterpartyCounterpartyID), buffer)
		buffer.WriteString(`yes</a></td>
                `)
	}
	buffer.WriteString(`
                <td class="d">`)
	hero.FormatInt(int64(contact.transfersCount), buffer)
	buffer.WriteString(`</td>
            </tr>
            </tbody>
        </table>
    </div>
    <div class="container-fluid">
        <div class="row">
            `)

	renderContactBalance(contact.ID, "Balance (no interest)", contact.balances.withoutInterest, false, buffer)
	renderContactBalance(contact.ID, "Balance with interest", contact.balances.withInterest, false, buffer)

	buffer.WriteString(`
        </div>
    </div>
    `)

	transfersInfo := contact.Data.GetTransfersInfo()
	if len(transfersInfo.OutstandingWithInterest) > 0 {

		buffer.WriteString(`
    <div class="card-body">
        <h3>Outstanding transfers: `)
		hero.FormatInt(int64(len(transfersInfo.OutstandingWithInterest)), buffer)
		buffer.WriteString(`</h3>
        <table class="table">
            <thead>
            <tr>
                <td class="d">#</td>
                <td>ID</td>
                <td>Dir</td>
                <td>Currency</td>
                <td class="d">Amount</td>
                <td class="d">%</td>
                <td class="d">Period</td>
            </tr>
            </thead>
            <tbody>
            `)
		for i, transfer := range transfersInfo.OutstandingWithInterest {
			buffer.WriteString(`
            <tr>
                <td class="d">`)
			hero.FormatInt(int64(i+1), buffer)
			buffer.WriteString(`</td>
                <td>`)
			hero.FormatInt(int64(transfer.TransferID), buffer)
			buffer.WriteString(`</td>
                <td>`)
			hero.EscapeHTML(string(transfer.Direction), buffer)
			buffer.WriteString(`</td>
                <td>`)
			hero.EscapeHTML(string(transfer.Currency), buffer)
			buffer.WriteString(`</td>
                <td class="d">`)
			hero.EscapeHTML(fmt.Sprintf("%v", transfer.Amount), buffer)
			buffer.WriteString(`</td>
                <td class="d">`)
			hero.EscapeHTML(fmt.Sprintf("%v", transfer.InterestPercent), buffer)
			buffer.WriteString(`</td>
                <td class="d">`)
			hero.FormatInt(int64(transfer.InterestPeriod), buffer)
			buffer.WriteString(`</td>
            </tr>
            `)
			for _, returned := range transfer.Returns {
				buffer.WriteString(`
            <tr>
                <td>🔙</td>
                <td>`)
				hero.FormatInt(int64(returned.TransferID), buffer)
				buffer.WriteString(`</td>
                <td colspan="2">`)
				hero.EscapeHTML(fmt.Sprintf("%v", returned.Time), buffer)
				buffer.WriteString(`</td>
                <td class="d">`)
				hero.EscapeHTML(fmt.Sprintf("%v", returned.Amount), buffer)
				buffer.WriteString(`</td>
            </tr>
            `)
			}
		}
		buffer.WriteString(`
            </tbody>
        </table>
    </div>
    `)
	}
	buffer.WriteString(`
</div>
`)

}
