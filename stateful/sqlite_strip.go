package stateful

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"

	sqlite "github.com/glebarez/go-sqlite"
)

// sqliteStripDriver 自定义 SQLite 驱动，在 Prepare 阶段剥离 FOR UPDATE 子句
// SQLite 不支持 FOR UPDATE（行锁），此驱动自动移除该子句以兼容测试环境
type sqliteStripDriver struct {
	inner driver.Driver
}

func init() {
	sql.Register("sqlite-strip", &sqliteStripDriver{inner: &sqlite.Driver{}})
}

func (d *sqliteStripDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.inner.Open(name)
	if err != nil {
		return nil, err
	}
	return &stripConn{Conn: conn}, nil
}

// stripConn 包装 driver.Conn，在 Prepare 时剥离 FOR UPDATE
type stripConn struct {
	driver.Conn
}

func (c *stripConn) Prepare(query string) (driver.Stmt, error) {
	return c.Conn.Prepare(stripForUpdate(query))
}

func (c *stripConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if cc, ok := c.Conn.(driver.ConnPrepareContext); ok {
		return cc.PrepareContext(ctx, stripForUpdate(query))
	}
	return c.Conn.Prepare(stripForUpdate(query))
}

// stripForUpdate 移除 SQL 中的 FOR UPDATE 子句（SQLite 不支持）
func stripForUpdate(query string) string {
	upper := strings.ToUpper(query)
	idx := strings.Index(upper, "FOR UPDATE")
	if idx < 0 {
		return query
	}
	return strings.TrimSpace(query[:idx])
}