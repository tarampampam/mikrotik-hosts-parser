package mikrotik

import "io"

type DNSStaticEntries []DNSStaticEntry

type RenderingOptions struct {
	Prefix, Postfix string
}

// Render mikrotik static dns entry and write it into some writer. Returned values is count of wrote bytes and error,
// if something goes wrong.
func (se DNSStaticEntries) Render(to io.Writer, opts ...RenderingOptions) (int, error) {
	var (
		buf     = make([]byte, 0, 128) // reusable
		total   int
		options RenderingOptions
	)

	if len(opts) > 0 {
		options = opts[0]
	}

	for i := 0; i < len(se); i++ {
		// append line breaker only for non-first entries
		if total > 0 {
			buf = append(buf, "\n"...)
		}

		if formattingErr := se[i].format(&buf, options.Prefix, options.Postfix); formattingErr == nil {
			// write buffer
			wrote, err := to.Write(buf)
			if err != nil {
				return total, err
			}

			total += wrote
		}

		// make buffer clean (capacity will keep maximum length)
		buf = buf[:0]
	}

	return total, nil
}
