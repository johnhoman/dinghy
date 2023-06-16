package fieldpath

import (
	"io"

	"github.com/pkg/errors"
)

// Parse the string representation into a FieldPath
func Parse(fp string) (*FieldPath, error) {
	return parseFieldPath(fp)
}

func newParser(fieldPath string) *parser {
	p := &parser{fieldPath: fieldPath}
	p.inc()
	return p
}

// parser parses a fieldPath string
// into indexes used for iterating on arbitrary json like
// data structures
type parser struct {
	fieldPath string
	char      byte
	pos       int
	next      int
}

func (fp *parser) inc() {
	if fp.next >= len(fp.fieldPath) {
		fp.char = 0
	} else {
		fp.char = fp.fieldPath[fp.next]
	}
	fp.pos = fp.next
	fp.next += 1
}

func (fp *parser) nextIndex() (index Index, err error) {

	switch fp.char {
	case '\'', '"':
		open := fp.char
		fp.inc()
		// quotes are used to parse non valid identifiers, e.g.
		// 'eks.amazonaws.com/role-arn', so consume all characters
		// until reaching the closing brace
		pos := fp.pos
		for fp.char != 0 && fp.char != open {
			fp.inc()
		}
		// the character here should be the open character
		if fp.char == 0 {
			err = errors.Wrapf(io.EOF, "quote %q at pos %d is never closed", open, pos)
			return
		}
		index = Index{it: IndexTypeMapKey, index: fp.fieldPath[pos:fp.pos]}
		fp.inc()
		return
	case '[':
		// brackets should behave like a map index
		pos := fp.pos
		fp.inc()
		index, err = fp.nextIndex()
		if err != nil {
			return
		}
		switch fp.char {
		case '=':
			fp.inc()
			var next Index
			next, err = fp.nextIndex()
			if err != nil {
				return
			}
			index = Index{
				it: IndexTypeQuery,
				query: Query{
					op:       QueryOpCmpEqual,
					argument: next.index,
				},
			}
			fallthrough
		case ']':
			fp.inc()
			return
		default:
			err = errors.Wrapf(io.EOF, "opening bracket was never closed: %q (pos %d)", "[", pos)
			return
		}
	case '.':
		fp.inc()
		index, err = fp.nextIndex()
		return
	case 0:
		err = io.EOF
		return
	default:
		switch {
		case isNumber(fp.char):
			pos := fp.pos
			for isNumber(fp.char) {
				fp.inc()
			}
			index = Index{it: IndexTypeArrayIndex, index: fp.fieldPath[pos:fp.pos]}
			return
		case isLetter(fp.char):
			pos := fp.pos
			for isAlpha(fp.char) {
				fp.inc()
			}
			index = Index{it: IndexTypeMapKey, index: fp.fieldPath[pos:fp.pos]}
			return
		default:
			err = errors.Errorf("unexpected character: %q (pos %d)", fp.char, fp.pos)
			return
		}
	}
}

func isNumber(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isAlpha(ch byte) bool {
	return isLetter(ch) || isNumber(ch)
}
