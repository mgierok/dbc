package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"

	"github.com/mgierok/dbc/internal/application/usecase"
	"github.com/mgierok/dbc/internal/infrastructure/config"
	"github.com/mgierok/dbc/internal/infrastructure/engine"
	"github.com/mgierok/dbc/internal/interfaces/tui"
)

func main() {
	cfgPath, err := config.DefaultPath()
	if err != nil {
		log.Fatalf("failed to resolve config path: %v", err)
	}

	cfg, err := config.LoadFile(cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := sql.Open("sqlite", cfg.Database.Path)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	engine := engine.NewSQLiteEngine(db)
	listTables := usecase.NewListTables(engine)
	getSchema := usecase.NewGetSchema(engine)
	listRecords := usecase.NewListRecords(engine)
	listOperators := usecase.NewListOperators(engine)

	if err := tui.Run(context.Background(), listTables, getSchema, listRecords, listOperators); err != nil {
		fmt.Printf("application error: %v\n", err)
	}
}
