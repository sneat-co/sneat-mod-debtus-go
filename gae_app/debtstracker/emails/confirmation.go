package emails

import (
	"fmt"

	"bitbucket.org/asterus/debtstracker-server/gae_app/debtstracker/models"
	"context"
)

func CreateConfirmationEmailAndQueueForSending(c context.Context, user models.AppUser, userEmail models.UserEmail) error {
	emailEntity := &models.EmailEntity{
		From:    "Alex @ DebtsTracker.io <alex@debtstracker.io>",
		To:      userEmail.ID,
		Subject: "Please confirm your account at DebtsTracker.io",
		BodyText: fmt.Sprintf(`%v, we are thrilled to have you on board!

To keep your account secure please confirm your email by clicking this link:

  >> https://debtstracker.io/confirm?email=%v&pin=%v

If you have any questions or issue please drop me an email to alex@debtstracker.io
--
Alex
Creator of https://DebtsTracker.io

We are social:
  FB page - https://www.facebook.com/debtstracker
  Twitter - https://twitter.com/debtstracker
`, user.FullName(), userEmail.ID, userEmail.ConfirmationPin()),
	}
	_, err := CreateEmailRecordAndQueueForSending(c, emailEntity)
	return err
}
