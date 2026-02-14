package main

import (
	"context"
	"database/sql"
	"errors"
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

	options := make([]tui.DatabaseOption, len(cfg.Databases))
	for i, database := range cfg.Databases {
		options[i] = tui.DatabaseOption{
			Name:       database.Name,
			ConnString: database.Path,
		}
	}

	selected, err := tui.SelectDatabase(options)
	if err != nil {
		if errors.Is(err, tui.ErrDatabaseSelectionCanceled) {
			return
		}
		log.Fatalf("failed to select database: %v", err)
	}

	db, err := sql.Open("sqlite", selected.ConnString)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("failed to close database after ping failure: %v", closeErr)
		}
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("failed to close database: %v", closeErr)
		}
	}()

	engine := engine.NewSQLiteEngine(db)
	listTables := usecase.NewListTables(engine)
	getSchema := usecase.NewGetSchema(engine)
	listRecords := usecase.NewListRecords(engine)
	listOperators := usecase.NewListOperators(engine)
	saveChanges := usecase.NewSaveTableChanges(engine)

	if err := tui.Run(context.Background(), listTables, getSchema, listRecords, listOperators, saveChanges); err != nil {
		fmt.Printf("application error: %v\n", err)
	}
}
