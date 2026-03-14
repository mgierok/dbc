package tui

import (
	"fmt"
	"strings"

	"github.com/mgierok/dbc/internal/application/dto"
)

func (m *Model) applyOperation(op stagedOperation) error {
	switch op.kind {
	case opInsertAdded:
		return m.insertPendingRowAt(op.insert.index, op.insert.row)
	case opInsertRemoved:
		_, err := m.removePendingInsert(op.insert.index)
		return err
	case opCellEdited:
		return m.applyCellEditState(op.cell, false)
	case opDeleteToggled:
		return m.setDeleteMark(op.del.key, op.del.identity, op.del.afterMarked)
	default:
		return fmt.Errorf("unsupported staged operation")
	}
}

func (m *Model) applyInverseOperation(op stagedOperation) error {
	switch op.kind {
	case opInsertAdded:
		_, err := m.removePendingInsert(op.insert.index)
		return err
	case opInsertRemoved:
		return m.insertPendingRowAt(op.insert.index, op.insert.row)
	case opCellEdited:
		return m.applyCellEditState(op.cell, true)
	case opDeleteToggled:
		return m.setDeleteMark(op.del.key, op.del.identity, op.del.beforeMarked)
	default:
		return fmt.Errorf("unsupported staged operation")
	}
}

func (m *Model) applyCellEditState(op cellEditOperation, useBefore bool) error {
	edit := op.after
	exists := op.afterExists
	explicitAuto := op.afterExplicitAuto
	if useBefore {
		edit = op.before
		exists = op.beforeExists
		explicitAuto = op.beforeExplicitAuto
	}
	switch op.target {
	case cellEditInsert:
		if op.insertIndex < 0 || op.insertIndex >= len(m.staging.pendingInserts) {
			return fmt.Errorf("insert index out of range")
		}
		row := m.staging.pendingInserts[op.insertIndex]
		if row.values == nil {
			row.values = make(map[int]stagedEdit, len(m.read.schema.Columns))
		}
		if row.explicitAuto == nil {
			row.explicitAuto = make(map[int]bool)
		}
		if exists {
			row.values[op.columnIndex] = edit
		} else {
			delete(row.values, op.columnIndex)
		}
		if op.columnIndex >= 0 && op.columnIndex < len(m.read.schema.Columns) && m.read.schema.Columns[op.columnIndex].AutoIncrement {
			if explicitAuto {
				row.explicitAuto[op.columnIndex] = true
			} else {
				delete(row.explicitAuto, op.columnIndex)
			}
		}
		m.staging.pendingInserts[op.insertIndex] = row
		return nil
	case cellEditPersisted:
		if strings.TrimSpace(op.recordKey) == "" {
			return fmt.Errorf("record key missing")
		}
		if m.staging.pendingUpdates == nil {
			m.staging.pendingUpdates = make(map[string]recordEdits)
		}
		edits := m.staging.pendingUpdates[op.recordKey]
		if edits.changes == nil {
			edits.changes = make(map[int]stagedEdit)
		}
		edits.identity = op.identity
		if exists {
			edits.changes[op.columnIndex] = edit
			m.staging.pendingUpdates[op.recordKey] = edits
			return nil
		}
		delete(edits.changes, op.columnIndex)
		if len(edits.changes) == 0 {
			delete(m.staging.pendingUpdates, op.recordKey)
			return nil
		}
		m.staging.pendingUpdates[op.recordKey] = edits
		return nil
	default:
		return fmt.Errorf("unsupported cell edit target")
	}
}

func (m *Model) setDeleteMark(key string, identity dto.RecordIdentity, marked bool) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("record key missing")
	}
	if m.staging.pendingDeletes == nil {
		m.staging.pendingDeletes = make(map[string]recordDelete)
	}
	if marked {
		m.staging.pendingDeletes[key] = recordDelete{identity: identity}
		return nil
	}
	delete(m.staging.pendingDeletes, key)
	return nil
}

func (m *Model) recordOperation(op stagedOperation) {
	m.staging.history = append(m.staging.history, op)
	m.staging.future = nil
}
