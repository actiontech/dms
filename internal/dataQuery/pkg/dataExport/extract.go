package data_export

import (
	"context"
	"database/sql"
)

type BaseExtract struct {
	db      *sql.DB
	rows    *sql.Rows
	columns []string
	query   string
}

func NewExtract(db *sql.DB, query string) Extractor {
	return &BaseExtract{
		db:    db,
		query: query,
	}
}

func (ex *BaseExtract) Start() error {
	rows, err := ex.db.QueryContext(context.TODO(), ex.query)
	if err != nil {
		return err
	}
	ex.rows = rows
	ex.columns, err = rows.Columns()
	if err != nil {
		return err
	}
	return nil
}

func (ex *BaseExtract) Columns() []string {
	return ex.columns
}

type Raw struct {
	ColumnName string
	Value      string // todo: 处理其他各种类型
}

func (ex *BaseExtract) Raws() ([]*Raw, bool, error) {
	value := make([]*Raw, len(ex.columns))
	if ex.rows.Next() {
		buf := make([]interface{}, len(ex.columns))
		data := make([]sql.NullString, len(ex.columns))
		for i := range buf {
			buf[i] = &data[i]
		}
		if err := ex.rows.Scan(buf...); err != nil {
			return nil, false, err
		}
		for i := 0; i < len(ex.columns); i++ {
			value[i] = &Raw{
				ColumnName: ex.columns[i],
				Value:      data[i].String,
			}
		}
	} else {
		return nil, false, nil
	}
	if err := ex.rows.Err(); err != nil {
		return nil, false, err
	}
	return value, true, nil
}

func (ex *BaseExtract) Close() error {
	if ex.rows != nil {
		return ex.rows.Close()
	}
	return nil
}
