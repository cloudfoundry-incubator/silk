package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"code.cloudfoundry.org/silk/controller"

	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

const postgresTimeNow = "EXTRACT(EPOCH FROM now())::numeric::integer"
const mysqlTimeNow = "UNIX_TIMESTAMP()"

var RecordNotAffectedError = errors.New("record not affected")
var MultipleRecordsAffectedError = errors.New("multiple records affected")

//go:generate counterfeiter -o fakes/db.go --fake-name Db . Db
type Db interface {
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	DriverName() string
}

//go:generate counterfeiter -o fakes/transaction.go --fake-name Transaction . Transaction
type Transaction interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Commit() error
	Rollback() error
	Rebind(string) string
}

//go:generate counterfeiter -o fakes/migrateAdapter.go --fake-name MigrateAdapter . migrateAdapter
type migrateAdapter interface {
	Exec(db Db, dialect string, m migrate.MigrationSource, dir migrate.MigrationDirection) (int, error)
}

type DatabaseHandler struct {
	migrator   migrateAdapter
	migrations *migrate.MemoryMigrationSource
	db         Db
	deleter    SubnetDeleter
	timeout    time.Duration
}

func NewDatabaseHandler(migrator migrateAdapter, db Db, sd SubnetDeleter, timeout time.Duration) *DatabaseHandler {
	return &DatabaseHandler{
		migrator: migrator,
		migrations: &migrate.MemoryMigrationSource{
			Migrations: []*migrate.Migration{
				&migrate.Migration{
					Id:   "1",
					Up:   []string{createSubnetTable(db.DriverName())},
					Down: []string{"DROP TABLE subnets"},
				},
			},
		},
		db:      db,
		deleter: sd,
		timeout: timeout,
	}
}

func (d *DatabaseHandler) All() ([]controller.Lease, error) {
	leases := []controller.Lease{}
	ctx, _ := context.WithTimeout(context.Background(), d.timeout)
	rows, err := d.db.QueryContext(ctx, "SELECT underlay_ip, overlay_subnet, overlay_hwaddr FROM subnets")
	if err != nil {
		return nil, fmt.Errorf("selecting all subnets: %s", err)
	}
	for rows.Next() {
		var underlayIP, overlaySubnet, overlayHWAddr string
		err := rows.Scan(&underlayIP, &overlaySubnet, &overlayHWAddr)
		if err != nil {
			return nil, fmt.Errorf("parsing result for all subnets: %s", err)
		}
		leases = append(leases, controller.Lease{
			UnderlayIP:          underlayIP,
			OverlaySubnet:       overlaySubnet,
			OverlayHardwareAddr: overlayHWAddr,
		})
	}

	return leases, nil
}

func (d *DatabaseHandler) AllActive(duration int) ([]controller.Lease, error) {
	timestamp, err := timestampForDriver(d.db.DriverName())
	if err != nil {
		return nil, err
	}
	leases := []controller.Lease{}
	ctx, _ := context.WithTimeout(context.Background(), d.timeout)
	rows, err := d.db.QueryContext(ctx, fmt.Sprintf("SELECT underlay_ip, overlay_subnet, overlay_hwaddr FROM subnets WHERE last_renewed_at + %d > %s", duration, timestamp))
	if err != nil {
		return nil, fmt.Errorf("selecting all active subnets: %s", err)
	}
	for rows.Next() {
		var underlayIP, overlaySubnet, overlayHWAddr string
		err := rows.Scan(&underlayIP, &overlaySubnet, &overlayHWAddr)
		if err != nil {
			return nil, fmt.Errorf("parsing result for all active subnets: %s", err)
		}
		leases = append(leases, controller.Lease{
			UnderlayIP:          underlayIP,
			OverlaySubnet:       overlaySubnet,
			OverlayHardwareAddr: overlayHWAddr,
		})
	}

	return leases, nil
}

func (d *DatabaseHandler) OldestExpired(expirationTime int) (*controller.Lease, error) {
	timestamp, err := timestampForDriver(d.db.DriverName())
	if err != nil {
		return nil, err
	}

	var underlayIP, overlaySubnet, overlayHWAddr string
	ctx, _ := context.WithTimeout(context.Background(), d.timeout)
	result := d.db.QueryRowContext(ctx, fmt.Sprintf("SELECT underlay_ip, overlay_subnet, overlay_hwaddr FROM subnets WHERE last_renewed_at + %d <= %s ORDER BY last_renewed_at ASC LIMIT 1", expirationTime, timestamp))
	err = result.Scan(&underlayIP, &overlaySubnet, &overlayHWAddr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan result: %s", err)
	}
	return &controller.Lease{
		UnderlayIP:          underlayIP,
		OverlaySubnet:       overlaySubnet,
		OverlayHardwareAddr: overlayHWAddr,
	}, nil
}

func (d *DatabaseHandler) Migrate() (int, error) {
	migrations := d.migrations
	numMigrations, err := d.migrator.Exec(d.db, d.db.DriverName(), *migrations, migrate.Up)
	if err != nil {
		return 0, fmt.Errorf("migrating: %s", err)
	}
	return numMigrations, nil
}

func (d *DatabaseHandler) AddEntry(lease controller.Lease) error {
	timestamp, err := timestampForDriver(d.db.DriverName())
	if err != nil {
		return err
	}

	query := fmt.Sprintf("INSERT INTO subnets (underlay_ip, overlay_subnet, overlay_hwaddr, last_renewed_at) VALUES ('%s', '%s', '%s', %s)", lease.UnderlayIP, lease.OverlaySubnet, lease.OverlayHardwareAddr, timestamp)
	ctx, _ := context.WithTimeout(context.Background(), d.timeout)
	_, err = d.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("adding entry: %s", err)
	}
	return nil
}

func commit(tx Transaction) error {
	err := tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction: %s", err) // TODO untested
	}
	return nil
}

func rollback(tx Transaction, err error) error {
	txErr := tx.Rollback()
	if txErr != nil {
		return fmt.Errorf("db rollback: %s (sql error: %s)", txErr, err)
	}
	return err
}

func (d *DatabaseHandler) DeleteEntry(underlayIP string) error {
	ctx, _ := context.WithTimeout(context.Background(), d.timeout)

	// For reasons we don't understand, postgres db.ExecContext does
	// not time out on network partition. But exec w/ transaction
	// correctly times out. So we want to use a transaction here.
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %s", err)
	}

	result, err := d.deleter.Delete(tx, underlayIP, d.timeout)
	if err != nil {
		return rollback(tx, fmt.Errorf("deleting entry: %s", err))
	}
	nRows, err := result.RowsAffected()
	if err != nil {
		return rollback(tx, fmt.Errorf("parse result: %s", err))
	}
	if nRows == 0 {
		return rollback(tx, RecordNotAffectedError)
	}
	if nRows > 1 {
		return rollback(tx, MultipleRecordsAffectedError)
	}
	return commit(tx)
}

func (d *DatabaseHandler) LeaseForUnderlayIP(underlayIP string) (*controller.Lease, error) {
	var overlaySubnet, overlayHWAddr string
	ctx, _ := context.WithTimeout(context.Background(), d.timeout)
	result := d.db.QueryRowContext(ctx, fmt.Sprintf("SELECT overlay_subnet, overlay_hwaddr FROM subnets WHERE underlay_ip = '%s'", underlayIP))
	err := result.Scan(&overlaySubnet, &overlayHWAddr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err // test me
	}
	return &controller.Lease{
		UnderlayIP:          underlayIP,
		OverlaySubnet:       overlaySubnet,
		OverlayHardwareAddr: overlayHWAddr,
	}, nil
}

func (d *DatabaseHandler) RenewLeaseForUnderlayIP(underlayIP string) error {
	timestamp, err := timestampForDriver(d.db.DriverName())
	if err != nil {
		return err
	}

	query := fmt.Sprintf("UPDATE subnets SET last_renewed_at = %s WHERE underlay_ip = '%s'", timestamp, underlayIP)
	ctx, _ := context.WithTimeout(context.Background(), d.timeout)
	_, err = d.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("renewing lease: %s", err)
	}
	return nil
}

func (d *DatabaseHandler) LastRenewedAtForUnderlayIP(underlayIP string) (int64, error) {
	var lastRenewedAt int64
	ctx, _ := context.WithTimeout(context.Background(), d.timeout)
	result := d.db.QueryRowContext(ctx, fmt.Sprintf("SELECT last_renewed_at FROM subnets WHERE underlay_ip = '%s'", underlayIP))
	err := result.Scan(&lastRenewedAt)
	if err != nil {
		return 0, err
	}
	return lastRenewedAt, nil
}

func (d *DatabaseHandler) SubnetForUnderlayIP(underlayIP string) (string, error) {
	var subnet string
	ctx, _ := context.WithTimeout(context.Background(), d.timeout)
	result := d.db.QueryRowContext(ctx, fmt.Sprintf("SELECT subnet FROM subnets WHERE underlay_ip = '%s'", underlayIP))
	err := result.Scan(&subnet)
	if err != nil {
		return "", err
	}
	return subnet, nil
}

func createSubnetTable(dbType string) string {
	baseCreateTable := "CREATE TABLE IF NOT EXISTS subnets (" +
		"%s" +
		", underlay_ip varchar(15) NOT NULL" +
		", overlay_subnet varchar(18) NOT NULL" +
		", overlay_hwaddr varchar(17) NOT NULL" +
		", last_renewed_at bigint NOT NULL" +
		", UNIQUE (underlay_ip)" +
		", UNIQUE (overlay_subnet)" +
		", UNIQUE (overlay_hwaddr)" +
		");"
	mysqlId := "id int NOT NULL AUTO_INCREMENT, PRIMARY KEY (id)"
	psqlId := "id SERIAL PRIMARY KEY"

	switch dbType {
	case "postgres":
		return fmt.Sprintf(baseCreateTable, psqlId)
	case "mysql":
		return fmt.Sprintf(baseCreateTable, mysqlId)
	}

	return ""
}

func timestampForDriver(driverName string) (string, error) {
	switch driverName {
	case "mysql":
		return mysqlTimeNow, nil
	case "postgres":
		return postgresTimeNow, nil
	default:
		return "", fmt.Errorf("database type %s is not supported", driverName)
	}
}
