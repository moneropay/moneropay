/*
 * MoneroPay is a Monero payment processor.
 * Copyright (C) 2026 Laurynas Četyrkinas <laurynas@digilol.net>
 * Copyright (C) 2022 İrem Kuyucu <siren@kernal.eu>
 *
 * MoneroPay is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * MoneroPay is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with MoneroPay.  If not, see <https://www.gnu.org/licenses/>.
 */

package daemon

import (
	"context"

	"github.com/rs/zerolog/log"
	"gitlab.com/moneropay/go-monero/walletrpc"
)

// runMigrations runs data migrations and shows deprecation warnings.
func (d *Daemon) runMigrations() {
	d.migrateReceivedAmount()
	d.showDeprecationWarnings()
}

// migrateReceivedAmount migrates old payment requests that don't have creation_height set.
func (d *Daemon) migrateReceivedAmount() {
	ctx := context.Background()
	rows, err := d.db.QueryContext(ctx,
		"SELECT subaddress_index,expected_amount,description,callback_url,created_at"+
			" FROM receivers WHERE creation_height IS NULL")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to query payment requests to migrate")
	}
	defer rows.Close()

	h, err := d.wallet.GetHeight(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get wallet height")
	}

	rs := make(map[uint64]*recv)
	for rows.Next() {
		var t recv
		if err := rows.Scan(&t.index, &t.expected, &t.description, &t.callbackUrl, &t.createdAt); err != nil {
			log.Fatal().Err(err).Msg("Failed to query payment requests to migrate")
		}
		t.creationHeight = h.Height
		rs[t.index] = &t
	}

	if len(rs) == 0 {
		return
	}

	log.Info().Msg("Migration started")
	resp, err := d.GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		In: true,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Migration failure")
	}

	maxHeight := d.lastCallbackHeight
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
			if eventHeight > d.lastCallbackHeight {
				if err := d.sendPaymentCallback(ctx, r, &t, locked); err != nil {
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
		if _, err := d.db.ExecContext(ctx,
			"UPDATE receivers SET received_amount=$1,creation_height=$2 WHERE subaddress_index=$3",
			r.received, r.creationHeight, r.index); err != nil {
			log.Fatal().Err(err).Msg("Migration failure")
		}
	}

	if maxHeight > d.lastCallbackHeight {
		d.lastCallbackHeight = maxHeight
		if err := d.saveLastCallbackHeight(ctx); err != nil {
			log.Fatal().Err(err).Uint64("height", d.lastCallbackHeight).
				Msg("Failed to save last callback height")
		}
		log.Info().Uint64("height", d.lastCallbackHeight).Msg("Saved last callback height")
	}
	log.Info().Msg("Migration ended")
}

// showDeprecationWarnings logs deprecation warnings for upcoming changes.
func (d *Daemon) showDeprecationWarnings() {
	log.Warn().Msg("Deprecated: TxHash (tx_hash) field will be removed the next major release (3.0.0). Please use TxHashList (tx_hash_list) instead. See here for more details: https://gitlab.com/moneropay/moneropay/-/merge_requests/13")
}
