package facade4debtus

import (
	"context"
	"errors"
	"fmt"
	"github.com/sneat-co/sneat-mod-debtus-go/debtus/const4debtus"
	"github.com/strongo/delaying"
	"sync"
)

var delayerUpdateUserHasDueTransfers,
	delayerUpdateSpaceHasDueTransfers delaying.Delayer

func InitDelays4debtus(mustRegisterFunc func(key string, i any) delaying.Delayer) {
	delayerUpdateUserHasDueTransfers = mustRegisterFunc("delayedUpdateUserHasDueTransfers", delayedUpdateUserHasDueTransfers)
	delayerUpdateSpaceHasDueTransfers = mustRegisterFunc("delayedUpdateSpaceHasDueTransfers", delayedUpdateSpaceHasDueTransfers)
}

func DelayUpdateHasDueTransfers(ctx context.Context, userID, spaceID string) error {
	if userID == "" {
		return errors.New("userID is a required parameter")
	}
	if spaceID == "" {
		return errors.New("userID is a required parameter")
	}
	var wg sync.WaitGroup
	wg.Add(2)
	errs := make([]error, 0, 2)
	go func() {
		defer wg.Done()
		err := delayerUpdateUserHasDueTransfers.EnqueueWork(ctx, delaying.With(const4debtus.QueueDebtus, "delayedUpdateUserHasDueTransfers", 0), userID, spaceID)
		if err != nil {
			errs = append(errs, err)
		}
	}()
	go func() {
		defer wg.Done()
		err := delayerUpdateSpaceHasDueTransfers.EnqueueWork(ctx, delaying.With(const4debtus.QueueDebtus, "delayedUpdateSpaceHasDueTransfers", 0), userID, spaceID)
		if err != nil {
			errs = append(errs, err)
		}
	}()
	if len(errs) > 0 {
		return fmt.Errorf("failed to DelayUpdateHasDueTransfers: %w", errors.Join(errs...))
	}
	return nil
}
