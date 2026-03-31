package dto

type DatabaseSelectorOptionSource string

const (
	DatabaseSelectorOptionSourceConfig DatabaseSelectorOptionSource = "config"
	DatabaseSelectorOptionSourceCLI    DatabaseSelectorOptionSource = "cli"
)

type DatabaseSelectorAdditionalOption struct {
	Name       string
	ConnString string
	Source     DatabaseSelectorOptionSource
}

type DatabaseSelectorLoadInput struct {
	AdditionalOptions []DatabaseSelectorAdditionalOption
}

type DatabaseSelectorOption struct {
	Name        string
	ConnString  string
	Source      DatabaseSelectorOptionSource
	ConfigIndex int
	CanEdit     bool
	CanDelete   bool
}

type DatabaseSelectorState struct {
	ActiveConfigPath   string
	Options            []DatabaseSelectorOption
	RequiresFirstEntry bool
}
