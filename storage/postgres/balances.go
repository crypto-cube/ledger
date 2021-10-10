package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v4"
	"github.com/numary/ledger/core"
	"github.com/numary/ledger/ledger/query"
)

func (s *PGStore) Balances(q query.Query) (query.Cursor, error) {
	c := query.Cursor{}

	var getBalances func(acc string) ([]core.Balance, error)
	getBalances = func(acc string) ([]core.Balance, error) {

		myChildren := strings.Count(acc, ":") + 1
		res := []core.Balance{}
		queryBalances := sqlbuilder.Select("account", "balance", "asset").From(s.table("balances"))
		var sqlBalances string
		var args []interface{}
		if acc == "" {
			queryBalances.Where(queryBalances.NotLike("account", "%:%"))
			sqlBalances, args = queryBalances.BuildWithFlavor(sqlbuilder.PostgreSQL)
		} else {
			queryBalances.Where(queryBalances.Like("account", acc+":%"))
			sqlBalances, args = queryBalances.BuildWithFlavor(sqlbuilder.PostgreSQL)
		}

		rows, err := s.Conn().Query(context.Background(),
			sqlBalances,
			args...)
		if err != nil {
			return res, err
		}

		for rows.Next() {
			var account string
			var balance int64
			var asset string

			rows.Scan(&account,
				&balance,
				&asset)

			if acc != "" && strings.Count(account, ":") != myChildren {
				continue
			}

			crv := core.Balance{
				Balance: balance,
				Account: account,
				Asset:   asset,
			}

			crv.Children, err = getBalances(account)

			if err != nil {
				return res, err
			}
			res = append(res, crv)
		}

		return res, nil
	}

	results, err := getBalances("")
	if err != nil {
		return c, err
	}

	c.HasMore = false
	c.PageSize = len(results)
	c.Remaining = 0
	c.Data = results

	return c, nil
}

func (s *PGStore) TouchBalance(ctx pgx.Tx, from string,
	to string, asset string,
	amount int64) error {
	if from != core.WORLD {
		acc := ""
		accs := strings.Split(from, ":")
		for _, k := range accs {
			if acc == "" {
				acc = k
			} else {
				acc = acc + ":" + k
			}
			queryBalance := sqlbuilder.Select("balance").
				From(s.table("balances"))
			queryBalance.Where(queryBalance.Equal("account", acc)).
				And(queryBalance.Equal("asset", asset))
			sqlBalance, args := queryBalance.BuildWithFlavor(sqlbuilder.PostgreSQL)
			var balance int64
			if err := s.Conn().QueryRow(context.TODO(), sqlBalance, args...).Scan(&balance); err != nil {
				return errors.New("Missing balance entry for group!")
			}

			balance = balance - amount
			if balance < 0 {
				return errors.New("Inconssistent balance for group / account")
			}

			updateBalance := sqlbuilder.Update(s.table("balances"))
			updateBalance.Set(updateBalance.Assign("balance", balance)).
				Where(updateBalance.Equal("account", acc)).
				And(updateBalance.Equal("asset", asset))
			updateSql, args := updateBalance.BuildWithFlavor(sqlbuilder.PostgreSQL)
			_, err := ctx.Exec(context.Background(),
				updateSql,
				args...,
			)

			if err != nil {
				return err
			}

		}
	}

	acc := ""
	accs := strings.Split(to, ":")
	for _, k := range accs {
		if acc == "" {
			acc = k
		} else {
			acc = acc + ":" + k
		}

		queryBalance := sqlbuilder.Select("balance").
			From(s.table("balances"))
		queryBalance.Where(queryBalance.Equal("account", acc)).
			And(queryBalance.Equal("asset", asset))
		sqlBalance, args := queryBalance.BuildWithFlavor(sqlbuilder.PostgreSQL)
		var balance int64 = 0
		if err := s.Conn().QueryRow(context.TODO(), sqlBalance, args...).Scan(&balance); err != nil {
			queryInsertBalance := sqlbuilder.InsertInto(s.table("balances"))
			queryInsertBalance.Cols("account", "balance", "asset")
			queryInsertBalance.Values(acc, amount, asset)
			insertBalanceSql, args := queryInsertBalance.BuildWithFlavor(sqlbuilder.PostgreSQL)
			_, err := ctx.Exec(
				context.Background(),
				insertBalanceSql,
				args...,
			)
			if err != nil {
				return errors.New("Undefined behaiviour inserting balances")
			}
		} else {
			balance = balance + amount

			updateBalance := sqlbuilder.Update(s.table("balances"))
			updateBalance.Set(updateBalance.Assign("balance", balance)).
				Where(updateBalance.Equal("account", acc)).
				And(updateBalance.Equal("asset", asset))
			updateSql, args := updateBalance.BuildWithFlavor(sqlbuilder.PostgreSQL)
			_, err := ctx.Exec(context.Background(),
				updateSql,
				args...,
			)

			if err != nil {
				return errors.New("Unable to update balance")
			}
		}

	}

	return nil
}
