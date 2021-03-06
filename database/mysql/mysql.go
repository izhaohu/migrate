package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/libgo/mysqlx"

	"github.com/izhaohu/migrate/database"
	"github.com/izhaohu/migrate/source"
)

func init() {
	database.Register("mysql", &Mysql{})
}

type Mysql struct {
	db     *sqlx.DB
	locked bool
}

func conf(uri string) mysqlx.Conf {
	return mysqlx.Conf{
		DSN:             uri,
		MaxOpenConns:    16,
		MaxIdleConns:    8,
		ConnMaxLifetime: time.Minute * 15,
	}
}

func (m *Mysql) Open(uri string) (database.Driver, error) {
	db := mysqlx.Register("db", conf(uri))

	// check if migration table exists
	query := `SHOW TABLES LIKE "_migration_table"`
	result := ""
	if err := db.QueryRow(query).Scan(&result); err != nil {
		if err != sql.ErrNoRows {
			db.Close()
			return nil, fmt.Errorf("db operation failed: %s", err.Error())
		} else {
			query = `CREATE TABLE _migration_table (module varchar(100) not null, version int not null, dirty boolean not null)`
			if _, err := db.Exec(query); err != nil {
				db.Close()
				return nil, fmt.Errorf("db operation failed: %s", err.Error())
			}
		}
	}

	return &Mysql{db: db, locked: false}, nil
}

func (m *Mysql) Close() error {
	m.db.Close()
	return nil
}

func (m *Mysql) Lock(md source.Module) error {
	if m.locked {
		return database.ErrLocked
	}

	query := fmt.Sprintf(`SELECT GET_LOCK("migration_lock_%s", 10)`, string(md))
	var success bool
	if err := m.db.QueryRow(query).Scan(&success); err != nil {
		return fmt.Errorf("lock failed: %s", err.Error())
	}

	if !success {
		return database.ErrLocked
	}

	m.locked = true
	return nil
}

func (m *Mysql) Unlock(md source.Module) error {
	if !m.locked {
		return nil
	}

	query := fmt.Sprintf(`SELECT RELEASE_LOCK("migration_lock_%s")`, string(md))
	if _, err := m.db.Exec(query); err != nil {
		return fmt.Errorf("unlock failed: %s", err.Error())
	}

	m.locked = false
	return nil
}

func (m *Mysql) Exec(sql string) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(sql)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (m *Mysql) Version(md source.Module) (int, bool, error) {
	d := false
	v := 0

	err := m.db.QueryRow(`SELECT version, dirty FROM _migration_table WHERE module=?`, string(md)).Scan(&v, &d)

	if err != nil && err != sql.ErrNoRows {
		return 0, false, err
	}

	if err != nil && err == sql.ErrNoRows {
		m.db.Exec(`INSERT INTO _migration_table(module, version, dirty) VALUE (?, ?, ?)`, string(md), v, d)
	}

	return v, d, nil
}

func (m *Mysql) SetVer(md source.Module, v int, d bool) error {
	_, err := m.db.Exec(`UPDATE _migration_table SET version=?, dirty=? WHERE module=?`, v, d, string(md))
	return err
}
