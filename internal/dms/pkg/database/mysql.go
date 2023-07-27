package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/go-sql-driver/mysql"
)

type mysqlManager struct {
	host     string
	port     string
	user     string
	password string
}

func NewMysqlManager(host, port, user, password string) ConnectorImpl {
	return &mysqlManager{
		host:     host,
		port:     port,
		user:     user,
		password: password,
	}
}

func (mm *mysqlManager) IsConnectable(ctx context.Context) (bool, error) {
	config := mysql.NewConfig()
	config.User = mm.user
	config.Passwd = mm.password
	config.Addr = net.JoinHostPort(mm.host, mm.port)
	config.ParseTime = true
	config.Loc = time.Local
	config.Timeout = 5 * time.Second
	config.Params = map[string]string{"charset": "utf8"}
	driver, err := mysql.NewConnector(config)
	if err != nil {
		return false, fmt.Errorf("IsConnectable connector err: %v", err)
	}

	pool := sql.OpenDB(driver)
	pool.SetMaxOpenConns(1)
	pool.SetMaxIdleConns(1)

	defer pool.Close()

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	if err = pool.PingContext(timeoutCtx); err != nil {
		return false, fmt.Errorf("IsConnectable ping err: %v", err)
	}

	return true, nil
}
