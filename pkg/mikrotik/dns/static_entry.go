package dns

import (
	"io"
	"reflect"
	"strings"
)

type (
	// Structure that can be "rendered" in RouterOS script format.
	StaticEntry struct {
		Address  string `comment:"IP address" property:"address" examples:"0.0.0.0"`
		Comment  string `comment:"Short description of the item" property:"comment" examples:"Any text"`
		Disabled bool   `comment:"Defines whether item is ignored or used" property:"disabled" examples:"yes,no"`
		Name     string `comment:"Host name" property:"name" examples:"www.example.com"`
		Regexp   string `property:"regexp" examples:".*\\.example\\.com"`
		TTL      string `comment:"Time To Live" property:"ttl" examples:"1d"` // @todo: Need more examples
	}

	// "Render-able" structure
	StaticEntries []StaticEntry
)

type (
	// Single entry "rendering" options.
	RenderEntryOptions struct {
		Prefix  string
		Postfix string
	}

	// Summary "rendering" options.
	RenderOptions struct {
		RenderEntryOptions
		RenderEmpty bool
	}
)

// Render mikrotik static dns entry and write it into some writer. Returned values is count of wrote bytes and error,
// if something goes wrong
func (e StaticEntries) Render(to io.Writer, options *RenderOptions) (int, error) { //nolint:gocyclo
	var (
		wroteTotal = 0
		ref        = reflect.TypeOf(StaticEntry{})
		address    = e.getStructPropertyValue(ref, "Address")
		comment    = e.getStructPropertyValue(ref, "Comment")
		disabled   = e.getStructPropertyValue(ref, "Disabled")
		name       = e.getStructPropertyValue(ref, "Name")
		regexp     = e.getStructPropertyValue(ref, "Regexp")
		ttl        = e.getStructPropertyValue(ref, "TTL")
	)

	var buf []byte

	for _, entry := range e {
		// skip entries without filled address property
		if entry.Address == "" {
			continue
		}

		// add line breaker if bugger is used in previous iteration
		if cap(buf) > 0 {
			buf = append(buf, "\n"...)
		}

		// write entry Prefix
		if options.RenderEntryOptions.Prefix != "" {
			buf = append(buf, options.RenderEntryOptions.Prefix+" "...)
		}

		// write "address"
		buf = append(buf, address+"="+entry.Address...)

		// write "comment"
		if entry.Comment != "" || options.RenderEmpty {
			buf = append(buf, " "+comment+`="`+e.escapeString(entry.Comment)+`"`...)
		}

		// write "disabled"
		if entry.Disabled {
			buf = append(buf, " "+disabled+"=yes"...)
		} else {
			buf = append(buf, " "+disabled+"=no"...)
		}

		// write "name"
		if entry.Name != "" || options.RenderEmpty {
			buf = append(buf, " "+name+`="`+e.escapeString(entry.Name)+`"`...)
		}

		// write "regexp"
		if entry.Regexp != "" || options.RenderEmpty {
			buf = append(buf, " "+regexp+`="`+entry.Regexp+`"`...)
		}

		// write "ttl"
		if entry.TTL != "" || options.RenderEmpty {
			buf = append(buf, " "+ttl+`="`+e.escapeString(entry.TTL)+`"`...)
		}

		// write entry Postfix
		if options.RenderEntryOptions.Postfix != "" {
			buf = append(buf, " "+options.RenderEntryOptions.Postfix...)
		}

		// write buffer
		wrote, err := to.Write(buf)
		if err != nil {
			return wroteTotal, err
		}
		wroteTotal += wrote

		// make buffer clean (capacity will keep maximum length)
		buf = buf[:0]
	}

	return wroteTotal, nil
}

// Escape string value chars for using in rendering.
func (StaticEntries) escapeString(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, `\`, ``), `"`, `\"`)
}

// Small helper for getting structure tag value.
func (StaticEntries) getStructPropertyValue(r reflect.Type, field string) string {
	const propertyTag string = "property"

	if field, ok := r.FieldByName(field); ok {
		val, _ := field.Tag.Lookup(propertyTag)

		return val
	}

	return ""
}
