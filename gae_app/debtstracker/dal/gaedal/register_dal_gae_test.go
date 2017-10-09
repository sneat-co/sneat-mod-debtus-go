package gaedal

import (
	"bitbucket.com/debtstracker/gae_app/debtstracker/dal"
	"testing"
)

func TestRegisterDal(t *testing.T) {
	// Pre-clean
	dal.Admin = nil
	dal.Contact = nil
	dal.DB = nil
	dal.Group = nil
	dal.Twilio = nil
	dal.HttpClient = nil
	dal.Invite = nil
	dal.LoginCode = nil
	dal.LoginPin = nil
	dal.Bill = nil
	dal.Receipt = nil
	dal.Reminder = nil
	dal.TgUser = nil
	dal.Transfer = nil
	dal.User = nil
	dal.UserBrowser = nil
	dal.UserGaClient = nil
	dal.UserGooglePlus = nil
	dal.UserFacebook = nil
	dal.UserOneSignal = nil

	// Execute
	RegisterDal()
	// Assert
	if dal.Admin == nil {
		t.Error("dal.Admin == nil")
	}
	if dal.Bill == nil {
		t.Error("dal.Bill == nil")
	}
	if dal.Contact == nil {
		t.Error("dal.Contact == nil")
	}
	if dal.DB == nil {
		t.Error("dal.DB == nil")
	}
	if dal.Receipt == nil {
		t.Error("dal.Receipt == nil")
	}
	if dal.Reminder == nil {
		t.Error("dal.Reminder == nil")
	}
	if dal.UserBrowser == nil {
		t.Error("dal.UserBrowser == nil")
	}
	if dal.Bill == nil {
		t.Error("dal.Bill == nil")
	}
	if dal.HttpClient == nil {
		t.Error("dal.HttpClient == nil")
	}
	if dal.Invite == nil {
		t.Error("dal.Invite == nil")
	}
	if dal.Group == nil {
		t.Error("dal.Invite == nil")
	}
	if dal.TgUser == nil {
		t.Error("dal.TgUser == nil")
	}
	if dal.Transfer == nil {
		t.Error("dal.Transfer == nil")
	}
	if dal.Twilio == nil {
		t.Error("dal.Twilio == nil")
	}
	if dal.User == nil {
		t.Error("dal.User == nil")
	}
	if dal.UserBrowser == nil {
		t.Error("dal.UserBrowser == nil")
	}
	if dal.UserGaClient == nil {
		t.Error("dal.UserGaClient == nil")
	}
	if dal.UserGooglePlus == nil {
		t.Error("dal.UserGooglePlus == nil")
	}
	if dal.UserFacebook == nil {
		t.Error("dal.UserFacebook == nil")
	}
	if dal.UserOneSignal == nil {
		t.Error("dal.UserOneSignal == nil")
	}
}
