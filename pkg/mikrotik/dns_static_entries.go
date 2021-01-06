package mikrotik

import "io"

type DNSStaticEntries []DNSStaticEntry

// Render mikrotik static dns entry and write it into some writer. Returned values is count of wrote bytes and error,
// if something goes wrong.
func (e DNSStaticEntries) Render(to io.Writer, prefix, postfix string) (int, error) {
	var (
		buf   = make([]byte, 0, 128) // reusable
		total int
	)

	for i := 0; i < len(e); i++ {
		// skip giveEntries without filled address property
		if e[i].Address == "" {
			continue
		}

		//// skip giveEntries without hostname and regex // FIXME apply this?
		//if e[i].Name == "" && e[i].Regexp == "" {
		//	continue
		//}

		// append line breaker only for non-first entries
		if total > 0 {
			buf = append(buf, "\n"...)
		}

		e[i].format(&buf, prefix, postfix)

		// write buffer
		if wrote, err := to.Write(buf); err != nil {
			return total, err
		} else {
			total += wrote
		}

		// make buffer clean (capacity will keep maximum length)
		buf = buf[:0]
	}

	return total, nil
}
