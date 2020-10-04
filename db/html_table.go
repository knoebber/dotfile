package db

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
)

const (
	htmlDownArrow  = "&#8595;"
	htmlUpArrow    = "&#8593;"
	htmlRightArrow = "&rarr;"
	htmlLeftArrow  = "&larr;"
	htmlSpace      = "&nbsp;"
	htmlCommaSpace = "," + htmlSpace
)

func aTag(q string, inner string) string {
	if q == "" {
		return ""
	}
	return fmt.Sprintf(`<a href="?%s">%s</a>`, q, inner)
}

// HTMLTable contains methods for generating HTML output from rows and columns.
type HTMLTable struct {
	Columns  []string
	Rows     []interface{}
	Controls *PageControls
}

func (h *HTMLTable) Query() string {
	return h.Controls.query
}

func (h *HTMLTable) TotalRows() int {
	return h.Controls.totalRows
}

// Header returns table headers tags with links for setting order.
func (h *HTMLTable) Header() template.HTML {
	var b strings.Builder

	for i, col := range h.Columns {
		query, curr := h.Controls.orderCol(i + 1)
		b.WriteString("<th>")
		if curr == "asc" {
			b.WriteString(aTag(query, htmlUpArrow+col))
		} else if curr == "desc" {
			b.WriteString(aTag(query, htmlDownArrow+col))
		} else {
			b.WriteString(aTag(query, col))
		}
		b.WriteString("</th>")
	}
	return template.HTML(b.String())
}

// Pages returns a series of <a> tags that link to the first, last, and surrounding pages.
func (h *HTMLTable) Pages() template.HTML {
	const pagesAround = 3
	var b strings.Builder

	curr := h.Controls.page
	total := h.Controls.totalPages()
	if total == 0 {
		return ""
	}

	b.WriteString(fmt.Sprintf("%d of %d results", len(h.Rows), h.Controls.totalRows))
	b.WriteString(htmlSpace)

	if curr > 1 {
		b.WriteString(aTag(h.Controls.previousPage(), htmlLeftArrow+" previous"))
		b.WriteString(htmlCommaSpace)
	}

	b.WriteString(aTag(h.Controls.firstPage(), "1"))
	b.WriteString(htmlCommaSpace)

	for p := curr - 1; p > 1 && p > curr-pagesAround; p-- {
		b.WriteString(aTag(h.Controls.toPage(p), strconv.Itoa(p)))
		b.WriteString(htmlCommaSpace)
	}
	if curr > 1 && curr < total {
		b.WriteString(aTag(h.Controls.toPage(curr), strconv.Itoa(curr)))
		b.WriteString(htmlCommaSpace)
	}
	for p := curr + 1; p < total && p < curr+pagesAround; p++ {
		b.WriteString(aTag(h.Controls.toPage(p), strconv.Itoa(p)))
		b.WriteString(htmlCommaSpace)
	}
	b.WriteString(aTag(h.Controls.lastPage(), strconv.Itoa(total)))
	b.WriteString(htmlSpace)
	b.WriteString(aTag(h.Controls.nextPage(), "next "+htmlRightArrow))
	return template.HTML(b.String())
}
