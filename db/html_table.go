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

// Query returns the current query, 'q'.
func (h *HTMLTable) Query() string {
	return h.Controls.query
}

// TotalRows returns the total amount of rows available.
func (h *HTMLTable) TotalRows() int {
	return h.Controls.totalRows
}

// Header returns table headers tags with links for setting order.
func (h *HTMLTable) Header() template.HTML {
	var b strings.Builder

	for i, col := range h.Columns {
		query, curr := h.Controls.orderCol(i + 1)
		_, _ = b.WriteString("<th>")
		if curr == "asc" {
			_, _ = b.WriteString(aTag(query, htmlUpArrow+col))
		} else if curr == "desc" {
			_, _ = b.WriteString(aTag(query, htmlDownArrow+col))
		} else {
			_, _ = b.WriteString(aTag(query, col))
		}
		_, _ = b.WriteString("</th>")
	}
	return template.HTML(b.String())
}

// Pages returns a series of <a> tags that link to the first, last, and surrounding pages.
func (h *HTMLTable) Pages() template.HTML {
	const pagesAround = 3
	var b strings.Builder

	curr := h.Controls.page
	total := h.Controls.totalPages()
	if total <= 1 {
		return ""
	}

	_, _ = b.WriteString(fmt.Sprintf("<i>%d of %d results</i>", len(h.Rows), h.Controls.totalRows))
	_, _ = b.WriteString(htmlSpace)

	if curr > 1 {
		// Write the first page and the previous page.
		_, _ = b.WriteString(aTag(h.Controls.previousPage(), htmlLeftArrow+" previous"))
		_, _ = b.WriteString(htmlCommaSpace)
		_, _ = b.WriteString(aTag(h.Controls.firstPage(), "1"))
		_, _ = b.WriteString(htmlCommaSpace)
	}

	for p := curr - pagesAround; p < curr; p++ {
		if p <= 1 {
			continue
		}
		// Write the pages before the current.
		_, _ = b.WriteString(aTag(h.Controls.toPage(p), strconv.Itoa(p)))
		_, _ = b.WriteString(htmlCommaSpace)
	}

	// Write the current page.
	_, _ = b.WriteString("<strong>")
	_, _ = b.WriteString(strconv.Itoa(curr))
	_, _ = b.WriteString("</strong>")

	for p := curr + 1; p < total && p < curr+pagesAround; p++ {
		// Write the pages after current.
		_, _ = b.WriteString(htmlCommaSpace)
		_, _ = b.WriteString(aTag(h.Controls.toPage(p), strconv.Itoa(p)))
	}

	if curr < total {
		// Write the last page and the next page.
		_, _ = b.WriteString(htmlCommaSpace)
		_, _ = b.WriteString(aTag(h.Controls.lastPage(), strconv.Itoa(total)))
		_, _ = b.WriteString(htmlSpace)
		_, _ = b.WriteString(aTag(h.Controls.nextPage(), "next "+htmlRightArrow))
	}
	return template.HTML(b.String())
}
