package main

import (
	"github.com/primadi/lokstra/lokstra_init"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/middleware/request_logger"
	"github.com/primadi/lokstra/services/dbpool_pg"
	"github.com/primadi/lokstra/services/email_smtp"
	"github.com/primadi/lokstra/services/eventbus"
)

func main() {
	// Register lokstra services
	dbpool_pg.Register()
	recovery.Register()
	request_logger.Register()
	email_smtp.Register()
	eventbus.Register()

	if err := lokstra_init.BootstrapAndRun(
		lokstra_init.WithAnnotations(true, "../../pkg"),
		lokstra_init.WithDbMigrations(true, "migrations"),
	); err != nil {
		panic(err)
	}
}
