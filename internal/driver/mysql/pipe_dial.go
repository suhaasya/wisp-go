package mysql

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/suhaasya/wisp-go/internal/driver"
)

var mysqlPipeReg atomic.Uint64

func registerMySQLPipeDial(pipeName string) (string, error) {
	name := fmt.Sprintf("wisp-pipe-%d", mysqlPipeReg.Add(1))
	pipe := pipeName
	mysqldriver.RegisterDialContext(name, func(ctx context.Context, addr string) (net.Conn, error) {
		return driver.DialMySQLPipe(ctx, pipe)
	})
	return name, nil
}
