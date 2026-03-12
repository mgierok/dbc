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

func (p *DirtyNavigationPolicy) BuildConfigPrompt(dirtyTableCount int) DirtyDecisionPrompt {
	return DirtyDecisionPrompt{
		Title:   "Config",
		Message: buildLeavePromptMessage("open config", dirtyTableCount),
		Options: []DirtyDecisionOption{
			{ID: DirtyDecisionSave, Label: "Save and open config"},
			{ID: DirtyDecisionDiscard, Label: "Discard and open config"},
			{ID: DirtyDecisionCancel, Label: "Cancel"},
		},
	}
}

func (p *DirtyNavigationPolicy) BuildQuitPrompt(dirtyTableCount int) DirtyDecisionPrompt {
	return DirtyDecisionPrompt{
		Title:   "Quit",
		Message: buildLeavePromptMessage("quit", dirtyTableCount),
		Options: []DirtyDecisionOption{
			{ID: DirtyDecisionSave, Label: "Save and quit"},
			{ID: DirtyDecisionDiscard, Label: "Discard and quit"},
			{ID: DirtyDecisionCancel, Label: "Cancel"},
		},
	}
}

func buildLeavePromptMessage(action string, dirtyTableCount int) string {
	if dirtyTableCount > 1 {
		return fmt.Sprintf("Unsaved changes detected in %d tables. Save all changes before you %s, discard them, or cancel.", dirtyTableCount, action)
	}
	return fmt.Sprintf("Unsaved changes detected. Save changes before you %s, discard them, or cancel.", action)
}
