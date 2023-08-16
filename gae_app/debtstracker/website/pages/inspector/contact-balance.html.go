// Code generated by hero.
// source: /Users/astec/go_workspace/src/github.com/sneat-co/sneat-mod-debtus-go/gae_app/debtstracker/website/pages/inspector/contact-balance.html
// DO NOT EDIT!
package inspector

import (
	"bytes"
	"fmt"

	"github.com/shiyanhui/hero"
)

func renderContactBalance(contactID string, title string, balances balancesByCurrency, showTransfers bool, buf *bytes.Buffer) {
	buf.WriteString(`
<div class="col-sm">
    <h3>`)
	hero.EscapeHTML(title, buf)
	buf.WriteString(`</h3>
    <table class="table table-bordered">
        <thead>
        <tr>
            <th>Currency</th>
            <th class=d>User</th>
            <th class=d>Contacts</th>
            `)
	if showTransfers {
		buf.WriteString(`
            <th class=d>Transfers</th>
            `)
	}
	buf.WriteString(`
        </tr>
        </thead>
        <tbody>
        `)
	for currency, balance := range balances.byCurrency {
		if balance.user == balance.contacts {
			buf.WriteString(`
        <tr>`)
		} else {
			buf.WriteString(`
        <tr class="table-danger">`)
		}
		buf.WriteString(`
            <td><a href="transfers?contact=`)
		hero.EscapeHTML(contactID, buf)
		buf.WriteString(`&currency=`)
		hero.EscapeHTML(fmt.Sprintf("%v", currency), buf)
		buf.WriteString(`">`)
		hero.EscapeHTML(fmt.Sprintf("%v", currency), buf)
		buf.WriteString(`</a></td>
            <td class=d>`)
		buf.WriteString(fmt.Sprintf("%v", balance.user))
		buf.WriteString(`</td>
            <td class=d>`)
		buf.WriteString(fmt.Sprintf("%v", balance.contacts))
		buf.WriteString(`</td>
            `)
		if showTransfers {
			buf.WriteString(`
            <td class=d>`)
			buf.WriteString(fmt.Sprintf("%v", balance.transfers))
			buf.WriteString(`</td>
            `)
		}
		buf.WriteString(`
        </tr>
        `)
	}
	buf.WriteString(`
        </tbody>
    </table>
</div>`)

}
