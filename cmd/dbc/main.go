package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

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
	connectionChecker := engine.NewSQLiteConnectionChecker()
	listConfiguredDatabases := usecase.NewListConfiguredDatabases(configStore)
	createConfiguredDatabase := usecase.NewCreateConfiguredDatabase(configStore, connectionChecker)
	updateConfiguredDatabase := usecase.NewUpdateConfiguredDatabase(configStore, connectionChecker)
	deleteConfiguredDatabase := usecase.NewDeleteConfiguredDatabase(configStore)
	getActiveConfigPath := usecase.NewGetActiveConfigPath(configStore)
	selectorState := tui.SelectorLaunchState{}

	for {
		selected, err := tui.SelectDatabaseWithState(
			context.Background(),
			listConfiguredDatabases,
			createConfiguredDatabase,
			updateConfiguredDatabase,
			deleteConfiguredDatabase,
			getActiveConfigPath,
			selectorState,
		)
		if err != nil {
			if errors.Is(err, tui.ErrDatabaseSelectionCanceled) {
				return
			}
			log.Fatalf("failed to select database: %v", err)
		}

		db, err := connectSelectedDatabase(selected)
		if err != nil {
			selectorState = tui.SelectorLaunchState{
				StatusMessage:    buildConnectionFailureStatus(selected, err.Error()),
				PreferConnString: selected.ConnString,
			}
			continue
		}
		selectorState = tui.SelectorLaunchState{}

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

func connectSelectedDatabase(selected tui.DatabaseOption) (*sql.DB, error) {
	return engine.OpenSQLiteDatabase(context.Background(), selected.ConnString)
}

func buildConnectionFailureStatus(selected tui.DatabaseOption, reason string) string {
	return fmt.Sprintf(
		"Connection failed for %q: %s. Choose another database or edit selected entry.",
		selected.Name,
		reason,
	)
}
