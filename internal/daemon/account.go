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

func countUnlockedTransfers(ctx context.Context, index, height uint64) (uint64, uint64,
    *walletrpc.Transfer, error) {
	resp, err := GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		In: true,
		SubaddrIndices: []uint64{index},
		FilterByHeight: true,
		MinHeight: height,
	})
	if err != nil {
		return 0, 0, nil, err
	}
	var r, h uint64
	var t *walletrpc.Transfer
	for _, i := range resp.In {
		if i.Confirmations >= 10 {
			r += i.Amount
			if i.Height > h {
				h = i.Height
			}
			t = &i
		}
	}
	return r, h, t, nil
}

func sendCompleteCallback(ctx context.Context, r recvAcct, t *walletrpc.Transfer) error {
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
	/* For now I can think of two ways to count transfers:
	 * 1. For each index get transfers given their last unlocked transfer's height.
	 * 2. Keep track of the smallest last unlocked transfer height, provide an array
	 * of indices and the height to a single get_transfers. Most efficient with a cron
	 * removing stale receiver rows, might cause very large queries to be sent to RPC
	 * server.
	*/
	ctx := context.Background()
	rows, err := pdb.Query(ctx,
	    "SELECT subaddress_index,expected_amount,received_amount,last_height FROM receivers")
	if err != nil {
		log.Error().Err(err).Msg("Failed to query payment requests")
		return
	}
	for rows.Next() {
		var r recvAcct
		// TODO: what if errnorows?
		if err := rows.Scan(&r.index, &r.expected, &r.received, &r.height); err != nil {
			log.Error().Err(err).Msg("Failed to query payment requests")
			return
		}
		// TODO: this should be an array of transfers unlocked after last height!
		received, height, transfer, err := countUnlockedTransfers(ctx, r.index, r.height)
		if err != nil {
			log.Error().Err(err).Msg("Failed to count unlocked transfers")
			continue
		}
		if received == 0 {
			continue
		}
		r.received += received
		r.height = height
		// Receivers with expected amount 0 never get removed
		if r.expected != 0 && r.received >= r.expected {
			if err := sendCompleteCallback(ctx, r, transfer); err != nil {
				log.Error().Err(err).Msg("Failed to send callback for unlocked payment")
			} else {
				log.Info().Uint64("address_index", r.index).Uint64("amount", transfer.Amount).
				    Msg("Sent callback for unlocked payment")
			}
			if _, err := pdb.Exec(ctx, "DELETE FROM receivers WHERE subaddress_index=$1",
			    r.index); err != nil {
				log.Error().Err(err).Msg("Failed to delete finished payment request")
			}
		} else {
			if _, err := pdb.Exec(ctx,
			    "UPDATE receivers SET received_amount=$1,last_height=$2 WHERE subaddress_index=$3",
			    r.received, r.height, r.index); err != nil {
				log.Error().Err(err).Msg("Failed to update payment request")
				continue
			}
		}
	}
}

func transferAccountingRunner() {
	for {
		accountTransfers()
		time.Sleep(30 * time.Second)
	}
}
