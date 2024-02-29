//go:build enterprise

package data_export

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	pkgParams "github.com/actiontech/dms/pkg/params"

	// mysql
	"github.com/go-sql-driver/mysql"
	// pg
	_ "github.com/lib/pq"
	// sqlserver
	_ "github.com/denisenkom/go-mssqldb"
	// oracle
	go_ora "github.com/sijms/go-ora/v2"
)

// 驱动参考来源 https://github.com/xo/usql
func NewDBConnection(ctx context.Context, dbType, host, port, user, password, schema string, addParams pkgParams.Params) (conn *sql.DB, err error) {
	switch strings.ToUpper(dbType) {
	case "MYSQL":
		conn, err = NewMysqlConn(host, port, user, password, schema)
	case "POSTGRESQL":
		conn, err = NewPGConn(host, port, user, password, schema)
	case "ORACLE":
		conn, err = NewOracleConn(host, port, user, password, schema, addParams)
	case "SQL SERVER":
		conn, err = NewSqlServerConn(host, port, user, password, schema)
	default:
		return nil, fmt.Errorf("db type %s is no supportd", dbType)
	}

	timeoutCtx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()

	if err = conn.PingContext(timeoutCtx); err != nil {
		return nil, fmt.Errorf("is connectable ping err: %v", err)
	}
	return
}

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

	return conn, nil
}

func NewPGConn(host, port, user, password, schema string) (*sql.DB, error) {
	// Initialize connection object.
	var connectionString string = fmt.Sprintf("host=%s  port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=3", host, port, user, password, schema)
	return sql.Open("postgres", connectionString)
}

func NewOracleConn(host, port, user, password, schema string, addParams pkgParams.Params) (*sql.DB, error) {
	urlOptions := map[string]string{
		"SID": addParams.GetParam("sid").String(),
	}
	// 端口转换
	_port, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}
	connStr := go_ora.BuildUrl(host, _port, addParams.GetParam("service_name").String(), user, password, urlOptions)
	return sql.Open("oracle", connStr)
}

func NewSqlServerConn(host, port, user, password, schema string) (*sql.DB, error) {
	var connectionString string = fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s", host, user, password, port, schema)
	return sql.Open("mssql", connectionString)
}
