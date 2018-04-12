package mysql

import (
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
)

var (
	MYSQL_SESSION_NIL_ERR = errors.New("MysqlMgr session nil.")
	MYSQL_NOTFOUND_ERR    = errors.New("not found!")
	MYSQL_DBFINDALL_ERR   = errors.New("MysqlMgr found error")
)

type MysqlConnInfo struct {
	mysql_conn_info *mysql.Config
}

func NewMysqlConnInfo(host string, port int, name, user, pass string) *MysqlConnInfo {
	cfg := mysql.NewConfig()
	cfg.User = user
	cfg.Passwd = pass
	cfg.Addr = fmt.Sprintf("%s:%d", host, port)
	cfg.Net = "tcp"
	cfg.DBName = name
	cfg.MultiStatements = true
	return &MysqlConnInfo{mysql_conn_info: cfg}
}

func (this *MysqlConnInfo) String() string {
	return this.mysql_conn_info.FormatDSN()
}
