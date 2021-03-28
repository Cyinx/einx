package mysql

import (
	"database/sql"
	"github.com/Cyinx/einx/component"
	"github.com/Cyinx/einx/event"
	"github.com/Cyinx/einx/module"
	"github.com/Cyinx/einx/slog"
	"time"
)

type EventReceiver = event.EventReceiver

type MysqlMgr struct {
	session      *sql.DB
	timeout      time.Duration
	dbcfg        *MysqlConnInfo
	component_id component.ComponentID
	m            EventReceiver
}

func NewMysqlMgr(m module.Module, dbcfg *MysqlConnInfo, timeout time.Duration) *MysqlMgr {
	return &MysqlMgr{
		session:      nil,
		timeout:      timeout,
		dbcfg:        dbcfg,
		component_id: component.GenComponentID(),
		m:            m.(event.EventReceiver),
	}
}

func (this *MysqlMgr) GetID() component.ComponentID {
	return this.component_id
}

func (this *MysqlMgr) GetType() component.ComponentType {
	return component.COMPONENT_TYPE_DB_MYSQL
}

func (this *MysqlMgr) Start() bool {
	var err error
	this.session, err = sql.Open("mysql", this.dbcfg.String())
	if err != nil {
		e := &event.ComponentEventMsg{}
		e.MsgType = event.EVENT_COMPONENT_ERROR
		e.Sender = this
		e.Attach = err
		this.m.PushEventMsg(e)
		slog.LogInfo("mysql", "mysql connect failed.")
		return false
	}
	return true
}

func (this *MysqlMgr) Close() {
	if this.session != nil {
		this.session.Close()
		this.session = nil
		slog.LogInfo("mysql", "mysql disconnect")
	}
}

func (this *MysqlMgr) Ping() error {
	if this.session != nil {
		return this.session.Ping()
	}
	return MYSQL_SESSION_NIL_ERR
}

func (this *MysqlMgr) GetSession() *sql.DB {
	return this.session
}

func (this *MysqlMgr) GetNamedRows(query interface{}) ([]map[string]interface{}, error) {
	return GetNamedRows(query)
}

func GetNamedRows(query interface{}) ([]map[string]interface{}, error) {
	row, ok := query.(*sql.Rows)
	var results []map[string]interface{}
	if ok == false {
		return results, MYSQL_GET_NAMED_RESULT_ERROR
	}
	column_types, err := row.ColumnTypes()
	if err != nil {
		slog.LogError("mysql", "columns error:%v", err)
		return nil, err
	}

	values := make([]interface{}, len(column_types))

	for c := true; c || row.NextResultSet(); c = false {

		//maybe this way is better
		//for k, c := range column_types {
		//	scans[k] = reflect.New(c.ScanType()).Interface()
		//}

		for k, c := range column_types {
			switch c.DatabaseTypeName() {
			case "INT":
				values[k] = new(int32)
			case "BIGINT":
				values[k] = new(int64)
			case "DOUBLE", "FLOAT":
				values[k] = new(float64)
			case "VARCHAR":
				values[k] = new(string)
			case "BLOB":
				values[k] = new([]byte)
			default:
				values[k] = new([]byte)
			}
		}

		for row.Next() {
			if err = row.Scan(values...); err != nil {
				slog.LogError("mysql", "Scan error:%v", err)
				return nil, err
			}

			result := make(map[string]interface{})
			for k, v := range values {
				key := column_types[k]
				switch s := v.(type) {
				case *int64:
					result[key.Name()] = *s
				case *float64:
					result[key.Name()] = *s
				case *string:
					result[key.Name()] = *s
				case *[]byte:
					result[key.Name()] = *s
				}
			}
			results = append(results, result)
		}
	}
	_ = row.Close()
	return results, nil
}
