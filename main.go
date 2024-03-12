package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

const name = "btclog"

const version = "0.0.1"

var revision = "HEAD"

type BtcLog struct {
	bun.BaseModel `bun:"table:btclog,alias:f"`
	Timestamp     int64     `bun:"timestamp,pk,notnull" json:"timestamp"`
	Last          float64   `bun:"last,notnull" json:"last"`
	Bid           float64   `bun:"bid,notnull" json:"bid"`
	Ask           float64   `bun:"ask,notnull" json:"ask"`
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

func main() {
	var dsn string
	var ver bool

	flag.StringVar(&dsn, "dsn", os.Getenv("DATABASE_URL"), "Database source")
	flag.BoolVar(&ver, "v", false, "show version")
	flag.Parse()

	if ver {
		fmt.Println(version)
		os.Exit(0)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	bundb := bun.NewDB(db, pgdialect.New())
	defer bundb.Close()

	_, err = bundb.NewCreateTable().Model((*BtcLog)(nil)).IfNotExists().Exec(context.Background())
	if err != nil {
		log.Println(err)
		return
	}

	resp, err := http.Get("https://coincheck.jp/api/ticker")
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Println(err)
		return
	}
	var btclog BtcLog
	err = json.NewDecoder(resp.Body).Decode(&btclog)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = bundb.NewInsert().Model(&btclog).Exec(context.Background())
	if err != nil {
		log.Println(err)
		return
	}
}
