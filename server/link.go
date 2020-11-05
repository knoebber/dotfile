package server

// Link populates a navbar link.
type Link struct {
	URL    string
	Title  string
	Active bool
}

// Class returns a class name for active styling.
func (l *Link) Class() string {
	if l.Active {
		return "active"
	}
	return ""
}

func newLink(url, title, currentTitle string) Link {
	return Link{
		URL:    url,
		Title:  title,
		Active: title == currentTitle,
	}
}
