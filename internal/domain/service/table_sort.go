package service

import (
	"sort"

	"github.com/mgierok/dbc/internal/domain/model"
)

func SortedTablesByName(tables []model.Table) []model.Table {
	sorted := append([]model.Table(nil), tables...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}
