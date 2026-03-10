package primitives

type StandardizedPopupWidthMode int

const (
	PopupWidthClamp   StandardizedPopupWidthMode = StandardizedPopupWidthMode(popupWidthClamp)
	PopupWidthContent StandardizedPopupWidthMode = StandardizedPopupWidthMode(popupWidthContent)
)

type StandardizedPopupRow struct {
	Text       string
	Selectable bool
	Selected   bool
}

type StandardizedPopupFooter struct {
	Left  string
	Right string
}

type StandardizedPopupSpec struct {
	Title               string
	Summary             string
	Rows                []StandardizedPopupRow
	Footer              StandardizedPopupFooter
	ScrollOffset        int
	VisibleRows         int
	ShowScrollIndicator bool
	WidthMode           StandardizedPopupWidthMode
	DefaultWidth        int
	MinWidth            int
	MaxWidth            int
	Styles              RenderStyles
}

func RenderStandardizedPopup(totalWidth, totalHeight int, spec StandardizedPopupSpec) []string {
	rows := make([]standardizedPopupRow, len(spec.Rows))
	for i, row := range spec.Rows {
		rows[i] = standardizedPopupRow{
			text:       row.Text,
			selectable: row.Selectable,
			selected:   row.Selected,
		}
	}

	return renderStandardizedPopup(totalWidth, totalHeight, standardizedPopupSpec{
		title:               spec.Title,
		summary:             spec.Summary,
		rows:                rows,
		footer:              standardizedPopupFooter{left: spec.Footer.Left, right: spec.Footer.Right},
		scrollOffset:        spec.ScrollOffset,
		visibleRows:         spec.VisibleRows,
		showScrollIndicator: spec.ShowScrollIndicator,
		widthMode:           standardizedPopupWidthMode(spec.WidthMode),
		defaultWidth:        spec.DefaultWidth,
		minWidth:            spec.MinWidth,
		maxWidth:            spec.MaxWidth,
		styles:              spec.Styles.inner,
	})
}

func PopupTextRows(rows []string) []StandardizedPopupRow {
	result := make([]StandardizedPopupRow, len(rows))
	for i, row := range rows {
		result[i] = StandardizedPopupRow{Text: row}
	}
	return result
}

func PopupSelectableRows(rows []string, selected int) []StandardizedPopupRow {
	result := make([]StandardizedPopupRow, len(rows))
	for i, row := range rows {
		result[i] = StandardizedPopupRow{
			Text:       row,
			Selectable: true,
			Selected:   i == selected,
		}
	}
	return result
}
