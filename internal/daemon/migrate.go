package daemon

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/moneropay/go-monero/walletrpc"
	"golang.org/x/exp/maps"
)

func daemonMigrate() {
	migrateReceivedAmount()
}

func migrateReceivedAmount() {
	ctx := context.Background()
	rows, err := pdb.Query(ctx,
	    "SELECT subaddress_index,expected_amount FROM receivers WHERE received_amount IS NULL")
	defer rows.Close()
	if err != nil {
		log.Fatal().Err(err).Msg("Migration failure")
	}
	recv := make(map[uint64]*recvAcct)
	for rows.Next() {
		var i, e uint64
		if err := rows.Scan(&i, &e); err != nil {
			log.Fatal().Err(err).Msg("Migration failure")
		}
		recv[i] = &recvAcct{index: i, expected: e}
	}
	if len(recv) == 0 {
		return
	}
	log.Info().Msg("Migration started")
	resp, err := GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		In: true,
		SubaddrIndices: maps.Keys(recv),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Migration failure")
	}
	for _, t := range resp.In {
		if r, ok := recv[t.SubaddrIndex.Minor]; ok {
			// 10 block lock is enforced as a blockchain consensus rule
			if t.Confirmations >= 10 {
				r.received += t.Amount
				// Don't depend on monero-wallet-rpc's ordering of transfers
				if t.Height > r.height {
					r.height = t.Height
				}
			}
		}
	}
	for _, r := range recv {
		if err := updatePaymentOnUnlock(ctx, *r); err != nil {
			log.Fatal().Err(err).Msg("Migration failure")
		}
	}
	log.Info().Msg("Migration ended")
}
