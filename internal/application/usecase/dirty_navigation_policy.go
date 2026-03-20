package usecase

import "fmt"

const (
	DirtyDecisionSave    = "save"
	DirtyDecisionDiscard = "discard"
	DirtyDecisionCancel  = "cancel"
)

type DirtyDecisionOption struct {
	ID    string
	Label string
}

type DirtyDecisionPrompt struct {
	Title   string
	Message string
	Options []DirtyDecisionOption
}

type DirtyNavigationPolicy struct{}

func NewDirtyNavigationPolicy() *DirtyNavigationPolicy {
	return &DirtyNavigationPolicy{}
}

func (p *DirtyNavigationPolicy) BuildTableSwitchPrompt(changeCount int) DirtyDecisionPrompt {
	if changeCount < 0 {
		changeCount = 0
	}
	return DirtyDecisionPrompt{
		Title: "Switch Table",
		Message: fmt.Sprintf(
			"Switching tables will cause loss of unsaved data (%d rows). Are you sure you want to discard unsaved data?",
			changeCount,
		),
		Options: []DirtyDecisionOption{
			{ID: DirtyDecisionDiscard, Label: "Discard changes and switch table"},
			{ID: DirtyDecisionCancel, Label: "Continue editing"},
		},
	}
}

func (p *DirtyNavigationPolicy) BuildConfigPrompt() DirtyDecisionPrompt {
	return DirtyDecisionPrompt{
		Title:   "Config",
		Message: "Unsaved changes detected. Choose save, discard, or cancel.",
		Options: []DirtyDecisionOption{
			{ID: DirtyDecisionSave, Label: "Save and open config"},
			{ID: DirtyDecisionDiscard, Label: "Discard and open config"},
			{ID: DirtyDecisionCancel, Label: "Cancel"},
		},
	}
}

func (p *DirtyNavigationPolicy) BuildDatabaseReloadPrompt(changeCount int) DirtyDecisionPrompt {
	if changeCount < 0 {
		changeCount = 0
	}
	return DirtyDecisionPrompt{
		Title: "Reload Database",
		Message: fmt.Sprintf(
			"Reloading the current database will cause loss of unsaved data (%d rows) unless you save first. Choose save, discard, or cancel.",
			changeCount,
		),
		Options: []DirtyDecisionOption{
			{ID: DirtyDecisionSave, Label: "Save and reload database"},
			{ID: DirtyDecisionDiscard, Label: "Discard changes and reload database"},
			{ID: DirtyDecisionCancel, Label: "Cancel"},
		},
	}
}

func (p *DirtyNavigationPolicy) BuildDatabaseOpenPrompt(changeCount int) DirtyDecisionPrompt {
	if changeCount < 0 {
		changeCount = 0
	}
	return DirtyDecisionPrompt{
		Title: "Open Database",
		Message: fmt.Sprintf(
			"Opening another database will cause loss of unsaved data (%d rows) unless you save first. Choose save, discard, or cancel.",
			changeCount,
		),
		Options: []DirtyDecisionOption{
			{ID: DirtyDecisionSave, Label: "Save and open database"},
			{ID: DirtyDecisionDiscard, Label: "Discard changes and open database"},
			{ID: DirtyDecisionCancel, Label: "Cancel"},
		},
	}
}

func (p *DirtyNavigationPolicy) BuildQuitPrompt(changeCount int) DirtyDecisionPrompt {
	if changeCount < 0 {
		changeCount = 0
	}
	return DirtyDecisionPrompt{
		Title: "Quit",
		Message: fmt.Sprintf(
			"Quitting will cause loss of unsaved data (%d rows). Are you sure you want to discard unsaved data and quit?",
			changeCount,
		),
		Options: []DirtyDecisionOption{
			{ID: DirtyDecisionDiscard, Label: "Discard changes and quit"},
			{ID: DirtyDecisionCancel, Label: "Continue editing"},
		},
	}
}
