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

func migrateReceivedAmount() {
	ctx := context.Background()
	rows, err := pdb.Query(ctx,
	    "SELECT subaddress_index FROM receivers WHERE received_amount IS NULL")
	defer rows.Close()
	if err != nil {
		log.Fatal(err)
	}
	recv := make(map[uint64]uint64)
	for rows.Next() {
		var i uint64
		if err := rows.Scan(&i); err != nil {
			log.Fatal(err)
		}
		recv[i] = 0
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
	for _, t := range resp.In {
		recv[t.SubaddrIndex.Minor] += t.Amount
	}
	tx, err := pdb.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for i, r := range recv {
		if _, err := tx.Exec(ctx, 
		    "UPDATE receivers SET received_amount=$1 WHERE subaddress_index=$2",
		    r, i); err != nil {
			    tx.Rollback(ctx)
			    log.Fatal(err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		log.Fatal(err)
	}
}
