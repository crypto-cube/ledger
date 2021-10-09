package storage

import (
	"github.com/numary/ledger/core"
	"github.com/numary/ledger/ledger/query"
	"github.com/numary/ledger/storage/postgres"
)

type Store interface {
	SaveTransactions([]core.Transaction) error
	CountTransactions() (int64, error)
	FindTransactions(query.Query) (query.Cursor, error)
	AggregateBalances(string) (map[string]int64, error)
	AggregateVolumes(string) (map[string]map[string]int64, error)
	CountAccounts() (int64, error)
	FindAccounts(query.Query) (query.Cursor, error)
	SaveMeta(string, string, core.Metadata) error
	GetMeta(string, string) (core.Metadata, error)
	Initialize() error
	Close()
}

func GetStore(name string) (Store, error) {
	return postgres.NewStore(name)
}
