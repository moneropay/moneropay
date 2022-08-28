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
	rows, err := pdbQuery(ctx,
	    "SELECT subaddress_index,expected_amount,description,callback_url,created_at" +
	    " FROM receivers WHERE creation_height IS NULL")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to query payment requests to migrate")
	}
	h, err := wallet.GetHeight(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get wallet height")
	}
	rs := make(map[uint64]*recv)
	for rows.Next() {
		var t recv
		if err := rows.Scan(&t.index, &t.expected, &t.description, &t.callbackUrl, &t.createdAt);
		    err != nil {
			log.Fatal().Err(err).Msg("Failed to query payment requests to migrate")
		}
		t.creationHeight = h.Height
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
	maxHeight := lastCallbackHeight
	for _, t := range resp.In {
		if r, ok := rs[t.SubaddrIndex.Minor]; ok {
			locked, eventHeight := getTransferLockStatus(t)
			// Creation height will be set to the earliest locked payment's height - 1
			// In case there are no locked transfers, it'll be set to the wallet height
			if locked {
				if t.Height < r.creationHeight {
					r.creationHeight = t.Height - 1
				}
			} else {
				r.received += t.Amount
			}
			if eventHeight > lastCallbackHeight {
				if err := callback(ctx, r, &t, locked); err != nil {
					log.Error().Err(err).Uint64("address_index", t.SubaddrIndex.Minor).
					    Uint64("amount", t.Amount).Str("tx_id", t.Txid).
					    Uint64("event_height", eventHeight).Bool("locked", locked).
					    Msg("Failed callback")
					continue
				}
				log.Info().Uint64("address_index", t.SubaddrIndex.Minor).
				    Uint64("amount", t.Amount).Str("tx_id", t.Txid).
				    Uint64("event_height", eventHeight).Bool("locked", locked).
				    Msg("Sent callback")
				if eventHeight > maxHeight {
					maxHeight = eventHeight
				}
			}
		}
	}
	for _, r := range rs {
		if err := pdbExec(ctx,
		    "UPDATE receivers SET received_amount=$1,creation_height=$2 WHERE subaddress_index=$3",
		    r.received, r.creationHeight, r.index); err != nil {
			log.Fatal().Err(err).Msg("Migration failure")
		}
	}
	if maxHeight > lastCallbackHeight {
		lastCallbackHeight = maxHeight
		if err := saveLastCallbackHeight(ctx); err != nil {
			log.Fatal().Err(err).Uint64("height", lastCallbackHeight).
			    Msg("Failed to save last callback height")
		}
	}
	log.Info().Msg("Migration ended")
}
