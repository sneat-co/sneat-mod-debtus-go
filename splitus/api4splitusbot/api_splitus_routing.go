package api4splitusbot

import (
	"github.com/sneat-co/sneat-core-modules/auth/api4auth"
	"github.com/strongo/strongoapp"
	"net/http"
)

func InitApiForSplitus(handle strongoapp.HandleHttpWithContext) {
	handle(http.MethodPost, "/api4debtus/bill-create", api4auth.AuthOnly(handleCreateBill))
	handle(http.MethodGet, "/api4debtus/bill-get", api4auth.AuthOnly(handleGetBill))
}
