package database

import (
	"context"
	"database/sql"
	"time"
)

//go:generate counterfeiter -o fakes/subnet_deleter.go --fake-name SubnetDeleter . SubnetDeleter
type SubnetDeleter interface {
	Delete(Transaction, string, time.Duration) (sql.Result, error)
}

type Deleter struct {
}

func (g *Deleter) Delete(tx Transaction, underlayIP string, timeout time.Duration) (sql.Result, error) {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	result, err := tx.ExecContext(
		ctx,
		tx.Rebind(`DELETE FROM subnets WHERE underlay_ip = ?`),
		underlayIP,
	)
	return result, err
}
