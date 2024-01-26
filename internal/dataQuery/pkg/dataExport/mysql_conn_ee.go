//go:build enterprise

package data_export

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/go-sql-driver/mysql"
)

func NewMysqlConn(host, port, user, password, schema string) (*sql.DB, error) {
	config := mysql.NewConfig()
	config.User = user
	config.Passwd = password
	config.DBName = schema
	config.Addr = net.JoinHostPort(host, port)
	config.ParseTime = true
	config.Loc = time.Local
	config.Timeout = 5 * time.Second
	config.Params = map[string]string{"charset": "utf8"}
	driver, err := mysql.NewConnector(config)
	if err != nil {
		return nil, fmt.Errorf("IsConnectable connector err: %v", err)
	}

	conn := sql.OpenDB(driver)
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)

	timeoutCtx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()

	if err = conn.PingContext(timeoutCtx); err != nil {
		return nil, fmt.Errorf("IsConnectable ping err: %v", err)
	}

	return conn, nil
}
