package sqlgen

// StmtRange is a half-open byte interval [Start, End) in the original script.
type StmtRange struct {
	Start int
	End   int
}

// SplitStatements splits a SQL script on statement terminators outside literals and comments.
func SplitStatements(sql string) []StmtRange {
	if sql == "" {
		return nil
	}
	delimiter := ";"
	var out []StmtRange
	i := 0
	b := []byte(sql)
	n := len(b)

	skipWS := func(p int) int {
		for p < n && (b[p] == ' ' || b[p] == '\t' || b[p] == '\r' || b[p] == '\n') {
			p++
		}
		return p
	}

	advanceStmtStart := func(p int) int {
		for p < n {
			p = skipWS(p)
			if p+1 < n && b[p] == '-' && b[p+1] == '-' {
				p += 2
				for p < n && b[p] != '\n' {
					p++
				}
				continue
			}
			break
		}
		return p
	}

	stmtStart := advanceStmtStart(0)

	lineStart := func(p int) bool {
		if p == 0 {
			return true
		}
		for k := p - 1; k >= 0; k-- {
			if b[k] == '\n' {
				return true
			}
			if b[k] != ' ' && b[k] != '\t' {
				return false
			}
		}
		return true
	}

	for i < n {
		if i+9 <= n && lineStart(i) && matchASCII(b, i, "DELIMITER") {
			j := i + 9
			j = skipByteWS(b, j, n)
			delimStart := j
			for j < n && b[j] != '\r' && b[j] != '\n' {
				j++
			}
			delimiter = string(b[delimStart:j])
			i = j
			if i < n {
				i++
			}
			stmtStart = advanceStmtStart(i)
			continue
		}

		if matchDelimiter(b, i, n, delimiter) {
			end := i
			if end > stmtStart {
				out = append(out, StmtRange{Start: stmtStart, End: end})
			}
			i += len(delimiter)
			stmtStart = advanceStmtStart(i)
			i = stmtStart
			continue
		}

		switch b[i] {
		case '-':
			if i+1 < n && b[i+1] == '-' {
				i += 2
				for i < n && b[i] != '\n' {
					i++
				}
				continue
			}
		case '/':
			if i+1 < n && b[i+1] == '*' {
				i += 2
				for i+1 < n && (b[i] != '*' || b[i+1] != '/') {
					i++
				}
				if i+1 < n {
					i += 2
				}
				continue
			}
		case '\'':
			i = scanSingleQuoted(b, i, n)
			continue
		case '"':
			i = scanDoubleQuoted(b, i, n)
			continue
		case '`':
			i = scanBacktick(b, i, n)
			continue
		case '$':
			if next, ok := scanDollarQuote(b, i, n); ok {
				i = next
				continue
			}
		}
		i++
	}
	if stmtStart < n {
		end := n
		for end > stmtStart && (b[end-1] == ' ' || b[end-1] == '\t' || b[end-1] == '\r' || b[end-1] == '\n') {
			end--
		}
		if end > stmtStart {
			out = append(out, StmtRange{Start: stmtStart, End: end})
		}
	}
	return out
}

func matchASCII(b []byte, i int, lit string) bool {
	if i+len(lit) > len(b) {
		return false
	}
	for j := 0; j < len(lit); j++ {
		if b[i+j] != lit[j] {
			return false
		}
	}
	if i+len(lit) < len(b) {
		c := b[i+len(lit)]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			return false
		}
	}
	return true
}

func skipByteWS(b []byte, i, n int) int {
	for i < n && (b[i] == ' ' || b[i] == '\t') {
		i++
	}
	return i
}

func matchDelimiter(b []byte, i, n int, delim string) bool {
	if i+len(delim) > n {
		return false
	}
	return string(b[i:i+len(delim)]) == delim
}

func scanSingleQuoted(b []byte, i, n int) int {
	i++
	for i < n {
		if b[i] == '\'' {
			if i+1 < n && b[i+1] == '\'' {
				i += 2
				continue
			}
			return i + 1
		}
		i++
	}
	return n
}

func scanDoubleQuoted(b []byte, i, n int) int {
	i++
	for i < n {
		if b[i] == '"' {
			if i+1 < n && b[i+1] == '"' {
				i += 2
				continue
			}
			return i + 1
		}
		i++
	}
	return n
}

func scanBacktick(b []byte, i, n int) int {
	i++
	for i < n {
		if b[i] == '`' {
			if i+1 < n && b[i+1] == '`' {
				i += 2
				continue
			}
			return i + 1
		}
		i++
	}
	return n
}

func scanDollarQuote(b []byte, i, n int) (next int, ok bool) {
	if b[i] != '$' {
		return i, false
	}
	j := i + 1
	for j < n && b[j] != '$' {
		j++
	}
	if j >= n {
		return i, false
	}
	tag := string(b[i : j+1])
	k := j + 1
	for k+len(tag) <= n {
		if string(b[k:k+len(tag)]) == tag {
			return k + len(tag), true
		}
		k++
	}
	return n, true
}

// StatementAtOffset returns the statement range containing offset (byte index).
func StatementAtOffset(sql string, offset int) (StmtRange, bool) {
	ranges := SplitStatements(sql)
	if len(ranges) == 0 {
		return StmtRange{}, false
	}
	if offset < 0 {
		offset = 0
	}
	if offset > len(sql) {
		offset = len(sql)
	}
	for _, r := range ranges {
		if offset >= r.Start && offset < r.End {
			return r, true
		}
	}
	for i, r := range ranges {
		if offset < r.Start {
			return r, true
		}
		if i+1 < len(ranges) && offset >= r.End && offset < ranges[i+1].Start {
			return ranges[i+1], true
		}
	}
	return ranges[len(ranges)-1], true
}
