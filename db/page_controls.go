package db

import (
	"fmt"
	"github.com/knoebber/dotfile/usererror"
	"math"
	"net/url"
	"strconv"
	"strings"
)

const (
	// DefaultLimit is the default limit for paginated queries.
	DefaultLimit = 25
	// MaxLimit is the max limit for paginated queries.
	MaxLimit = 500
)

// PageControls maps url queries to query parameters.
type PageControls struct {
	Values    url.Values
	totalRows int
	query     string
	orderBy   int    // The column number to order by, 1 indexed.
	order     string // desc, asc or an empty string.
	limit     int
	page      int
}

// Set sets private attributes from Values.
// This must be called before other methods.
//
// query: 'q'
// page: 'p'
// order: 'o'
// order by: 'ob'
// limit: 'l'
func (p *PageControls) Set() error {
	if p.Values == nil {
		p.Values = make(url.Values)
	}

	p.query = p.Values.Get("q")

	orderBy := p.Values.Get("ob")
	if orderBy == "" {
		p.orderBy = 1
	} else if p.orderBy, _ = strconv.Atoi(orderBy); p.orderBy < 1 {
		return usererror.Invalid(`Invalid query: "ob" (order by) must be positive integer.`)
	}

	p.order = strings.ToLower(p.Values.Get("o"))
	if !(p.order == "" || p.order == "desc" || p.order == "asc") {
		return usererror.Invalid(`Invalid query: "o" (order) must be one of: "", "asc", "desc".`)
	}

	limit := p.Values.Get("l")
	if limit == "" {
		p.limit = DefaultLimit
	} else {
		p.limit, _ = strconv.Atoi(limit)
		if p.limit < 1 || p.limit > MaxLimit {
			return usererror.Invalid(fmt.Sprintf(`Invalid query: "l" (limit) must be positive integer less than %d.`, MaxLimit))
		}
	}

	page := p.Values.Get("p")
	if page == "" {
		p.page = 1
	} else {
		p.page, _ = strconv.Atoi(page)
		if p.page < 1 {
			return usererror.Invalid(`Invalid query: "p" (page) must be positive integer.`)
		}
	}
	return nil
}

func (p *PageControls) sqlSuffix() string {
	return fmt.Sprintf(" ORDER BY %d %s LIMIT %d OFFSET %d", p.orderBy, p.order, p.limit, (p.page-1)*p.limit)
}

func (p *PageControls) totalPages() int {
	if p.totalRows == 0 || p.limit == 0 {
		return 0
	}
	return int(math.Ceil(float64(p.totalRows) / float64(p.limit)))
}

func (p *PageControls) orderCol(n int) (encoded string, curr string) {
	// Reset value map after mutating.
	defer p.Values.Set("o", p.Values.Get("o"))
	defer p.Values.Set("ob", p.Values.Get("ob"))

	if n != p.orderBy { // Ordering a new column, set next order to desc.
		p.Values.Set("ob", strconv.Itoa(n))
		p.Values.Set("o", "desc")
	} else if p.order == "" { // Ordering the same column, cycle through desc, asc, "".
		p.Values.Set("o", "desc")
	} else if p.order == "desc" {
		p.Values.Set("o", "asc")
		curr = "desc"
	} else if p.order == "asc" {
		p.Values.Set("o", "")
		p.Values.Set("ob", "")
		curr = "asc"
	}

	return p.Values.Encode(), curr
}

func (p *PageControls) toPage(n int) string {
	if n > p.totalPages() {
		return ""
	}
	defer p.Values.Set("p", p.Values.Get("p"))

	p.Values.Set("p", strconv.Itoa(n))
	return p.Values.Encode()
}

func (p *PageControls) nextPage() string {
	if p.page >= p.totalPages() {
		return ""
	}
	defer p.Values.Set("p", p.Values.Get("p"))

	p.Values.Set("p", strconv.Itoa(p.page+1))
	return p.Values.Encode()
}

func (p *PageControls) previousPage() string {
	if p.page == 1 {
		return ""
	}
	defer p.Values.Set("p", p.Values.Get("p"))

	p.Values.Set("p", strconv.Itoa(p.page-1))
	return p.Values.Encode()
}

func (p *PageControls) firstPage() string {
	defer p.Values.Set("p", p.Values.Get("p"))

	p.Values.Set("p", "1")
	return p.Values.Encode()
}

func (p *PageControls) lastPage() string {
	defer p.Values.Set("p", p.Values.Get("p"))

	p.Values.Set("p", strconv.Itoa(p.totalPages()))
	return p.Values.Encode()
}
