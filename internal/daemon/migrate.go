package daemon

import (
	"context"
	"log"

	"golang.org/x/exp/maps"
	"gitlab.com/moneropay/go-monero/walletrpc"
)

func daemonMigrate() {
	migrateReceivedAmount()
}

type oldRecv struct {
	amount, height uint64
}

func migrateReceivedAmount() {
	ctx := context.Background()
	rows, err := pdb.Query(ctx,
	    "SELECT subaddress_index FROM receivers WHERE received_amount IS NULL")
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	recv := make(map[uint64]*oldRecv)
	for rows.Next() {
		var i uint64
		if err := rows.Scan(&i); err != nil {
			log.Fatal(err)
		}
		recv[i] = &oldRecv{0, 0}
	}
	if len(recv) == 0 {
		return
	}
	resp, err := GetTransfers(ctx, &walletrpc.GetTransfersRequest{
		In: true,
		SubaddrIndices: maps.Keys(recv),
	})
	if err != nil {
		log.Fatal(err)
	}
	if len(resp.In) == 0 {
		return
	}
	for _, t := range resp.In {
		if r, ok := recv[t.SubaddrIndex.Minor]; ok {
			// 10 block lock is enforced as a blockchain consensus rule
			if t.Confirmations >= 10 {
				r.amount += t.Amount
				if t.Height > r.height {
					r.height = t.Height
				}
			} else {
			}
		}
	}
	tx, err := pdb.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for i, v := range recv {
		if _, err := tx.Exec(ctx,
		    "UPDATE receivers SET received_amount=$1,last_height=$2 WHERE subaddress_index=$3",
		    v.amount, v.height, i); err != nil {
			    tx.Rollback(ctx)
			    log.Fatal(err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		log.Fatal(err)
	}
}
