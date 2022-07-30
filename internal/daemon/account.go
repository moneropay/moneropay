package daemon

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/moneropay/go-monero/walletrpc"
)

type recvAcct struct {
	index uint64
	expected uint64
	received uint64
	height uint64
}

func updatePaymentOnUnlock(ctx context.Context, r recvAcct) error {
	_, err := pdb.Exec(ctx,
	    "UPDATE receivers SET received_amount=$1,last_height=$2 WHERE subaddress_index=$3",
	    r.received, r.height, r.index)
	return err
}

func countUnlockedTransfers(ctx context.Context, r recvAcct) {
	resp, err := GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		In: true,
		SubaddrIndices: []uint64{r.index},
		FilterByHeight: true,
		MinHeight: r.height,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to get transfers")
		return
	}
	updated := false
	for _, i := range resp.In {
		if i.Confirmations >= 10 {
			r.received += i.Amount
			if i.Height > r.height {
				r.height = i.Height
			}
			updated = true
			if err := sendUnlockedCallback(ctx, r, i); err != nil {
				log.Error().Err(err).Uint64("address_index", r.index).
				    Uint64("amount", i.Amount).Str("tx_id", i.Txid).
				    Msg("Failed to send callback for unlocked payment")
			} else {
				log.Info().Uint64("address_index", r.index).Uint64("amount", i.Amount).
				    Str("tx_id", i.Txid).Msg("Sent callback for unlocked payment")
			}
		}
	}
	if updated {
		if err := updatePaymentOnUnlock(ctx, r); err != nil {
			log.Error().Err(err).Msg("Failed to update payment request")
		}
	}
}

func sendUnlockedCallback(ctx context.Context, r recvAcct, t walletrpc.Transfer) error {
	var c callbackRequest
	c.Amount.Expected = r.expected
	c.Amount.Covered.Total = r.received
	c.Amount.Covered.Unlocked = r.received
	c.Complete = true
	c.CreatedAt = time.Now()
	c.Transaction = ReceiveTransaction{
		Amount: t.Amount,
		Confirmations: t.Confirmations,
		DoubleSpendSeen: t.DoubleSpendSeen,
		Fee: t.Fee,
		Height: t.Height,
		Timestamp: time.Unix(int64(t.Timestamp), 0),
		TxHash: t.Txid,
		UnlockTime: t.UnlockTime,
	}
	var u string
	if err := pdb.QueryRow(ctx,
	    "SELECT callback_url FROM receivers WHERE subaddress_index=$1", r.index).
	    Scan(&u); err != nil {
		return err
	}
	return sendCallback(u, c)
}

func accountTransfers() {
	ctx := context.Background()
	rows, err := pdb.Query(ctx,
	    "SELECT subaddress_index,expected_amount,received_amount,last_height FROM receivers")
	if err != nil {
		log.Error().Err(err).Msg("Failed to query payment requests")
		return
	}
	for rows.Next() {
		var r recvAcct
		if err := rows.Scan(&r.index, &r.expected, &r.received, &r.height); err != nil {
			log.Error().Err(err).Msg("Failed to query payment requests")
			return
		}
		countUnlockedTransfers(ctx, r)
	}
}

func transferAccountingRunner() {
	for {
		accountTransfers()
		time.Sleep(30 * time.Second)
	}
}
