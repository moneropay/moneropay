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
	transfer *walletrpc.Transfer
}

func countUnlockedTransfers(ctx context.Context, r *recvAcct) (uint64, error) {
	resp, err := GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		In: true,
		SubaddrIndices: []uint64{r.index},
		FilterByHeight: true,
		MinHeight: r.height,
	})
	if err != nil {
		return 0, err
	}
	var received uint64
	leftToPay := r.expected - r.received
	for _, i := range resp.In {
		if i.Confirmations >= 10 {
			received += i.Amount
			if i.Height > r.height {
				r.height = i.Height
			}
			r.transfer = &i
			if received >= leftToPay {
				break
			}
		}
	}
	r.received += received
	return received, nil
}

func processUnlockedPayment(ctx context.Context, r recvAcct) {
	// Receivers with expected amount 0 never get removed
	if r.expected != 0 && r.received >= r.expected {
		if err := sendCompleteCallback(ctx, r); err != nil {
			log.Error().Err(err).Msg("Failed to send callback for unlocked payment")
		} else {
			log.Info().Uint64("address_index", r.index).Uint64("amount", r.transfer.Amount).
			    Msg("Sent callback for unlocked payment")
		}
		if _, err := pdb.Exec(ctx, "DELETE FROM subaddresses WHERE address_index=$1",
		    r.index); err != nil {
			log.Error().Err(err).Msg("Failed to delete finished payment request")
		}
	} else {
		if _, err := pdb.Exec(ctx,
		    "UPDATE receivers SET received_amount=$1,last_height=$2 WHERE subaddress_index=$3",
		    r.received, r.height, r.index); err != nil {
			log.Error().Err(err).Msg("Failed to update payment request")
		}
	}
}

func sendCompleteCallback(ctx context.Context, r recvAcct) error {
	var c callbackRequest
	c.Amount.Expected = r.expected
	c.Amount.Covered.Total = r.received
	c.Amount.Covered.Unlocked = r.received
	c.Complete = true
	c.CreatedAt = time.Now()
	c.Transaction = ReceiveTransaction{
		Amount: r.transfer.Amount,
		Confirmations: r.transfer.Confirmations,
		DoubleSpendSeen: r.transfer.DoubleSpendSeen,
		Fee: r.transfer.Fee,
		Height: r.transfer.Height,
		Timestamp: time.Unix(int64(r.transfer.Timestamp), 0),
		TxHash: r.transfer.Txid,
		UnlockTime: r.transfer.UnlockTime,
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
		received, err := countUnlockedTransfers(ctx, &r)
		if err != nil {
			log.Error().Err(err).Msg("Failed to count unlocked transfers")
			continue
		}
		if received == 0 {
			continue
		}
		processUnlockedPayment(ctx, r)
	}
}

func transferAccountingRunner() {
	for {
		accountTransfers()
		time.Sleep(30 * time.Second)
	}
}
