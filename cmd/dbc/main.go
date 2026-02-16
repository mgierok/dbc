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

	configStore := config.NewStore(cfgPath)
	listConfiguredDatabases := usecase.NewListConfiguredDatabases(configStore)
	createConfiguredDatabase := usecase.NewCreateConfiguredDatabase(configStore)
	updateConfiguredDatabase := usecase.NewUpdateConfiguredDatabase(configStore)
	deleteConfiguredDatabase := usecase.NewDeleteConfiguredDatabase(configStore)
	getActiveConfigPath := usecase.NewGetActiveConfigPath(configStore)

	for {
		selected, err := tui.SelectDatabase(
			context.Background(),
			listConfiguredDatabases,
			createConfiguredDatabase,
			updateConfiguredDatabase,
			deleteConfiguredDatabase,
			getActiveConfigPath,
		)
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

		engine := engine.NewSQLiteEngine(db)
		listTables := usecase.NewListTables(engine)
		getSchema := usecase.NewGetSchema(engine)
		listRecords := usecase.NewListRecords(engine)
		listOperators := usecase.NewListOperators(engine)
		saveChanges := usecase.NewSaveTableChanges(engine)

		runErr := tui.Run(context.Background(), listTables, getSchema, listRecords, listOperators, saveChanges)
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("failed to close database: %v", closeErr)
		}
		if errors.Is(runErr, tui.ErrOpenConfigSelector) {
			continue
		}
		if runErr != nil {
			fmt.Printf("application error: %v\n", runErr)
		}
		return
	}
}
