package daemon


import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/moneropay/go-monero/walletrpc"
)

func daemonMigrate() {
	migrateReceivedAmount()
}

func migrateReceivedAmount() {
	ctx := context.Background()
	rows, err := pdb.Query(ctx,
	    "SELECT subaddress_index,expected_amount,description,callback_url,created_at" +
	    " FROM receivers WHERE creation_height IS NULL")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to query payment requests to migrate")
	}
	rs := make(map[uint64]*recv)
	for rows.Next() {
		var t recv
		if err := rows.Scan(&t.index, &t.expected, &t.description, &t.callbackUrl, &t.createdAt);
		    err != nil {
			log.Fatal().Err(err).Msg("Failed to query payment requests to migrate")
		}
		t.creationHeight = lastCallbackHeight
		rs[t.index] = &t
	}
	if len(rs) == 0 {
		return
	}
	log.Info().Msg("Migration started")
	resp, err := GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		In: true,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Migration failure")
	}
	h := lastCallbackHeight
	for _, t := range resp.In {
		u := t.Height
		unlocked := false
		if r, ok := rs[t.SubaddrIndex.Minor]; ok {
			if t.Confirmations >= 10 {
				// If the transfer is unlocked compare the block which it unlocked at
				// (t.Height + t.UnlockTime) to the block that caused the last callback
				if t.UnlockTime == 0 || t.UnlockTime - t.Height < 10 {
					u += 10
					unlocked = true
				} else if t.UnlockTime - t.Height <= t.Confirmations {
					u = t.UnlockTime
					unlocked = true
				}
			}
			// Creation height will be set to the earliest locked payment's height - 1
			// In case there are no locked transfers, it'll be set to the last callback height
			if !unlocked {
				if t.Height < (r.creationHeight - 1) {
					r.creationHeight = t.Height - 1
				}
			} else {
				r.received += t.Amount
			}
			if u > lastCallbackHeight {
				if err := callback(r, &t); err != nil {
					log.Error().Err(err).Str("tx_id", t.Txid).
					    Msg("Failed callback for new payment")
					continue
				}
				log.Info().Uint64("address_index", t.SubaddrIndex.Minor).
				    Uint64("amount", t.Amount).Str("tx_id", t.Txid).
				    Uint64("callback_height", u).Bool("unlocked", unlocked).
				    Msg("Sent callback")
				if u > h {
					h = u
				}
			}
		}
	}
	for _, r := range rs {
		if _, err := pdb.Exec(ctx,
		    "UPDATE receivers SET received_amount=$1,creation_height=$2 WHERE subaddress_index=$3",
		    r.received, r.creationHeight); err != nil {
			log.Fatal().Err(err).Msg("Migration failure")
		}
	}
	if h > lastCallbackHeight {
		if err := saveLastCallbackHeight(ctx); err != nil {
			log.Fatal().Err(err).Uint64("height", lastCallbackHeight).
			    Msg("Failed to save last callback height")
		}
	}
	log.Info().Msg("Migration ended")
}
