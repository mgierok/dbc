package engine

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/mgierok/dbc/internal/domain/model"
	"github.com/mgierok/dbc/internal/domain/service"
)

const maxMaterializedRecordCellBytes = 256 * 1024

type tableColumnInfo struct {
	cid             int
	name            string
	typ             string
	notNull         bool
	unique          bool
	defaultValue    *string
	primaryKeyOrder int
	autoIncrement   bool
	foreignKeys     []model.ForeignKeyRef
}

func (c tableColumnInfo) toModelColumn() model.Column {
	return model.Column{
		Name:          c.name,
		Type:          c.typ,
		Nullable:      !c.notNull,
		PrimaryKey:    c.primaryKeyOrder > 0,
		Unique:        c.unique,
		DefaultValue:  c.defaultValue,
		AutoIncrement: c.autoIncrement,
		ForeignKeys:   append([]model.ForeignKeyRef(nil), c.foreignKeys...),
	}
}

func (e *SQLiteEngine) tableColumnInfos(ctx context.Context, tableName string) (columns []tableColumnInfo, err error) {
	tableSQL, err := e.tableDefinitionSQL(ctx, tableName)
	if err != nil {
		return nil, err
	}
	uniqueColumns, err := e.uniqueSingleColumnIndexes(ctx, tableName)
	if err != nil {
		return nil, err
	}
	foreignKeysByColumn, err := e.foreignKeysByColumn(ctx, tableName)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("PRAGMA table_info(%s)", quoteIdentifier(tableName))
	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	for rows.Next() {
		var (
			cid     int
			name    string
			typ     string
			notNull int
			dflt    sql.NullString
			pk      int
		)
		if err := rows.Scan(&cid, &name, &typ, &notNull, &dflt, &pk); err != nil {
			return nil, err
		}
		var defaultValue *string
		if dflt.Valid {
			defaultValue = &dflt.String
		}
		columns = append(columns, tableColumnInfo{
			cid:             cid,
			name:            name,
			typ:             typ,
			notNull:         notNull != 0,
			unique:          pk == 0 && uniqueColumns[name],
			defaultValue:    defaultValue,
			primaryKeyOrder: pk,
			autoIncrement:   pk > 0 && columnHasAutoIncrement(tableSQL, name),
			foreignKeys:     append([]model.ForeignKeyRef(nil), foreignKeysByColumn[name]...),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

func (e *SQLiteEngine) uniqueSingleColumnIndexes(ctx context.Context, tableName string) (uniqueColumns map[string]bool, err error) {
	query := fmt.Sprintf("PRAGMA index_list(%s)", quoteIdentifier(tableName))
	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	uniqueColumns = make(map[string]bool)
	for rows.Next() {
		var (
			seq     int
			name    string
			unique  int
			origin  string
			partial int
		)
		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			return nil, err
		}
		if unique == 0 || partial != 0 || strings.EqualFold(origin, "pk") {
			continue
		}
		columns, err := e.indexColumns(ctx, name)
		if err != nil {
			return nil, err
		}
		if len(columns) == 1 {
			uniqueColumns[columns[0]] = true
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return uniqueColumns, nil
}

func (e *SQLiteEngine) indexColumns(ctx context.Context, indexName string) (columns []string, err error) {
	query := fmt.Sprintf("PRAGMA index_info(%s)", quoteIdentifier(indexName))
	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	columns = make([]string, 0, 1)
	for rows.Next() {
		var (
			seqno int
			cid   int
			name  sql.NullString
		)
		if err := rows.Scan(&seqno, &cid, &name); err != nil {
			return nil, err
		}
		if cid < 0 || !name.Valid {
			continue
		}
		columns = append(columns, name.String)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return columns, nil
}

func (e *SQLiteEngine) foreignKeysByColumn(ctx context.Context, tableName string) (foreignKeys map[string][]model.ForeignKeyRef, err error) {
	query := fmt.Sprintf("PRAGMA foreign_key_list(%s)", quoteIdentifier(tableName))
	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	foreignKeys = make(map[string][]model.ForeignKeyRef)
	for rows.Next() {
		var (
			id       int
			seq      int
			refTable string
			from     string
			to       sql.NullString
			onUpdate string
			onDelete string
			match    string
		)
		if err := rows.Scan(&id, &seq, &refTable, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			return nil, err
		}
		ref := model.ForeignKeyRef{Table: refTable}
		if to.Valid {
			ref.Column = to.String
		}
		foreignKeys[from] = append(foreignKeys[from], ref)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return foreignKeys, nil
}

func appendIdentityProjection(selectParts []string, pkColumns []tableColumnInfo) []string {
	for index, column := range pkColumns {
		selectParts = append(selectParts, identityProjectionForColumn(column, index))
	}
	if len(pkColumns) > 0 {
		selectParts = append(selectParts, identityAvailabilityProjection(pkColumns))
	}
	return selectParts
}

func appendEditableProjection(selectParts []string, columns []tableColumnInfo) []string {
	for index, column := range columns {
		selectParts = append(selectParts, editableFromBrowseProjectionForColumn(column, index))
	}
	return selectParts
}

func displayProjectionForColumn(column tableColumnInfo, index int) string {
	columnRef := quoteIdentifier(column.name)
	byteLength := fmt.Sprintf("length(CAST(%s AS BLOB))", columnRef)
	if isBlobType(column.typ) {
		return fmt.Sprintf(
			"CASE WHEN %s IS NULL THEN NULL WHEN %s > %d THEN printf('<blob truncated %%d bytes>', %s) ELSE printf('<blob %%d bytes>', %s) END AS %s",
			columnRef,
			byteLength,
			maxMaterializedRecordCellBytes,
			byteLength,
			byteLength,
			quoteIdentifier(displayColumnAlias(index)),
		)
	}
	return fmt.Sprintf(
		"CASE WHEN %s IS NULL THEN NULL WHEN %s > %d THEN printf('<truncated %%d bytes>', %s) ELSE CAST(%s AS TEXT) END AS %s",
		columnRef,
		byteLength,
		maxMaterializedRecordCellBytes,
		byteLength,
		columnRef,
		quoteIdentifier(displayColumnAlias(index)),
	)
}

func editableFromBrowseProjectionForColumn(column tableColumnInfo, index int) string {
	columnRef := quoteIdentifier(column.name)
	byteLength := fmt.Sprintf("length(CAST(%s AS BLOB))", columnRef)
	if isBlobType(column.typ) {
		return fmt.Sprintf(
			"CASE WHEN %s IS NULL THEN 1 ELSE 0 END AS %s",
			columnRef,
			quoteIdentifier(editableColumnAlias(index)),
		)
	}
	return fmt.Sprintf(
		"CASE WHEN %s IS NULL THEN 1 WHEN %s > %d THEN 0 ELSE 1 END AS %s",
		columnRef,
		byteLength,
		maxMaterializedRecordCellBytes,
		quoteIdentifier(editableColumnAlias(index)),
	)
}

func identityProjectionForColumn(column tableColumnInfo, index int) string {
	columnRef := quoteIdentifier(column.name)
	return fmt.Sprintf(
		"CASE WHEN %s THEN %s END AS %s",
		materializationSafeCondition(columnRef),
		columnRef,
		quoteIdentifier(identityColumnAlias(index)),
	)
}

func identityAvailabilityProjection(pkColumns []tableColumnInfo) string {
	conditions := make([]string, 0, len(pkColumns))
	for _, column := range pkColumns {
		conditions = append(conditions, materializationSafeCondition(quoteIdentifier(column.name)))
	}
	return fmt.Sprintf(
		"CASE WHEN %s THEN 1 ELSE 0 END AS %s",
		strings.Join(conditions, " AND "),
		quoteIdentifier(identityAvailabilityAlias),
	)
}

func materializationSafeCondition(columnRef string) string {
	return fmt.Sprintf("%s IS NULL OR length(CAST(%s AS BLOB)) <= %d", columnRef, columnRef, maxMaterializedRecordCellBytes)
}

func primaryKeyColumnsInOrder(columns []tableColumnInfo) []tableColumnInfo {
	pkColumns := make([]tableColumnInfo, 0, len(columns))
	for _, column := range columns {
		if column.primaryKeyOrder > 0 {
			pkColumns = append(pkColumns, column)
		}
	}
	sort.SliceStable(pkColumns, func(i, j int) bool {
		if pkColumns[i].primaryKeyOrder == pkColumns[j].primaryKeyOrder {
			return pkColumns[i].cid < pkColumns[j].cid
		}
		return pkColumns[i].primaryKeyOrder < pkColumns[j].primaryKeyOrder
	})
	return pkColumns
}

func materializeDisplayValue(raw any) model.Value {
	if raw == nil {
		return model.Value{IsNull: true}
	}
	switch typed := raw.(type) {
	case []byte:
		return model.Value{Text: string(typed)}
	default:
		return model.Value{Text: fmt.Sprint(typed)}
	}
}

func materializeIdentityValue(columnType string, raw any) (model.Value, error) {
	text, isNull, err := identityInputFromRaw(columnType, raw)
	if err != nil {
		return model.Value{}, err
	}
	return service.ParseValue(columnType, text, isNull, true)
}

func identityInputFromRaw(columnType string, raw any) (string, bool, error) {
	if raw == nil {
		return "", true, nil
	}
	if isBlobType(columnType) {
		bytes, err := rawBytes(raw)
		if err != nil {
			return "", false, err
		}
		return "0x" + hex.EncodeToString(bytes), false, nil
	}
	switch typed := raw.(type) {
	case []byte:
		return string(typed), false, nil
	default:
		return fmt.Sprint(typed), false, nil
	}
}

func rawBytes(raw any) ([]byte, error) {
	switch typed := raw.(type) {
	case []byte:
		return append([]byte(nil), typed...), nil
	case string:
		return []byte(typed), nil
	default:
		return nil, fmt.Errorf("unsupported blob identity type %T", raw)
	}
}

func recordRowKey(identity model.RecordIdentity) string {
	if len(identity.Keys) == 0 {
		return ""
	}
	parts := make([]string, 0, len(identity.Keys))
	for _, key := range identity.Keys {
		value := key.Value.Text
		if key.Value.IsNull {
			value = "NULL"
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key.Column, value))
	}
	return strings.Join(parts, "|")
}

func identityAvailable(raw any) bool {
	return projectedFlagEnabled(raw)
}

func projectedFlagEnabled(raw any) bool {
	switch typed := raw.(type) {
	case int64:
		return typed != 0
	case int:
		return typed != 0
	case bool:
		return typed
	case []byte:
		return string(typed) == "1"
	case string:
		return typed == "1"
	default:
		return false
	}
}

const identityAvailabilityAlias = "__dbc_identity_available"

func identityColumnAlias(index int) string {
	return fmt.Sprintf("__dbc_identity_%d", index)
}

func displayColumnAlias(index int) string {
	return fmt.Sprintf("__dbc_display_%d", index)
}

func editableColumnAlias(index int) string {
	return fmt.Sprintf("__dbc_editable_%d", index)
}

func isBlobType(columnType string) bool {
	return strings.Contains(strings.ToUpper(strings.TrimSpace(columnType)), "BLOB")
}
