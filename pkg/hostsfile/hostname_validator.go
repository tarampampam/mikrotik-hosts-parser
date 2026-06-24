package hostsfile

// validateHostname checks domain and Punycode structural validity with 0 allocations.
func validateHostname(domain []byte) bool { //nolint:gocyclo
	domainLen := len(domain)
	if domainLen == 0 || domainLen > 253 {
		return false
	}

	labelStart := 0
	for i := 0; i <= domainLen; i++ {
		// Check for label boundaries (dot or end of string)
		if i == domainLen || domain[i] == '.' {
			labelLen := i - labelStart
			if labelLen == 0 {
				return false
			}
			if labelLen > 63 {
				return false
			}

			// Validate the structure of the isolated label slice
			if !validateLabelStructure(domain[labelStart:i]) {
				return false
			}
			labelStart = i + 1

			continue
		}

		// Fast ASCII valid-character bitmap check (a-z, 0-9, '-', '_')
		ch := domain[i]
		isValid := (ch >= 'a' && ch <= 'z') ||
			(ch >= '0' && ch <= '9') ||
			(ch >= 'A' && ch <= 'Z') ||
			ch == '-' || ch == '_'

		if !isValid {
			return false
		}
	}

	return true
}

func validateLabelStructure(label []byte) bool { //nolint:gocyclo
	labelLen := len(label)

	// CVE-2026-39821: label can't start or end with hyphens
	if label[0] == '-' || label[labelLen-1] == '-' {
		return false
	}

	// is label is Punycode (starts with "xn--")
	if labelLen >= 4 && (label[0] == 'x' || label[0] == 'X') && //nolint:nestif
		(label[1] == 'n' || label[1] == 'N') &&
		label[2] == '-' && label[3] == '-' {

		if labelLen == 4 {
			return false
		}

		hasHyphen := false
		for i := 4; i < labelLen; i++ {
			ch := label[i]
			isPunyRune := (ch >= 'a' && ch <= 'z') ||
				(ch >= '0' && ch <= '9') ||
				(ch >= 'A' && ch <= 'Z') ||
				ch == '-'

			if !isPunyRune {
				return false
			}

			// "--" in Punycode is allowed only in the prefix "xn--", so we need to check for consecutive hyphens
			if ch == '-' {
				if hasHyphen {
					return false
				}
				hasHyphen = true
			} else {
				hasHyphen = false
			}
		}
	}

	return true
}
