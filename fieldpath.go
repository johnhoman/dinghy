package kustomize

import (
	"github.com/pkg/errors"
	"io"
	"strconv"
)

type (
	indexType string
	queryOp   string
)

const (
	indexTypeQuery      indexType = "Query"
	indexTypeMapKey     indexType = "MapKey"
	indexTypeArrayIndex indexType = "ArrayIndex"

	queryOpEq = "="
)

func NewFieldPath(fieldPath string) (*FieldPath, error) {
	indexes := make([]fieldPathIndex, 0)
	parser := newFieldPathParser(fieldPath)
	for {
		index, err := parser.nextIndex()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		indexes = append(indexes, index)
	}
	return &FieldPath{
		indexes:   indexes,
		fieldPath: fieldPath,
	}, nil
}

// FieldPath represents a path traversal sequence
// of a yaml document. A FieldPath will support indexing,
// both arrays and maps. It will also support indexing
// arrays on field names and values of the map, such as
// name="main". Field names that require using characters
// that aren't alpha-numeric, such as on of /-. will need
// to be enclosed in single or double quotes.
//
// Valid fieldPaths
// ----------------
// foo.bar
// foo[example.com].baz
// foo[example.com].baz
// foo[0].com.baz
// foo.bar[0].baz
// foo['example.com/foo'].com.baz
// foo[name=main].com.baz
//
// brackets are equivalent to
//
// There are only two types of indexing
//  1. Comparison, which is only valid for an array type. Comparison
//     is used for selecting a single map in an array of maps
//  2. Fixed, which is an integer for an array or a string for a map.
type FieldPath struct {
	indexes   []fieldPathIndex
	fieldPath string
}

func (fp *FieldPath) SetValue(m map[string]any, value any) error {
	if len(fp.indexes) == 0 {
		return nil
	}

	var current any = m
	for k, index := range fp.indexes {
		end := len(fp.indexes)-1 == k

		switch index.indexType {
		case indexTypeMapKey:
			mapping, ok := current.(map[string]any)
			if !ok {
				return errors.Errorf("expected type `map[string]any{}`, got `%T`", current)
			}
			if end {
				mapping[index.index] = value
				return nil
			}
			// lookahead to the next element
			next := fp.indexes[k+1]
			v, ok := mapping[index.index]
			if !ok {
				// I might need to make this a slice depending
				// on what the next type is
				if next.indexType == indexTypeMapKey {
					mapping[index.index] = make(map[string]any)
				} else {
					mapping[index.index] = make([]any, 0)
				}
				v = mapping[index.index]
			}
			// if the next element is an array, make sure it's big enough. We
			// won't be able to change the size later because append
			if it, ok := v.([]any); ok && next.indexType != indexTypeQuery {
				rank, err := strconv.Atoi(next.index)
				if err != nil {
					return err
				}
				for rank >= len(it) {
					it = append(it, nil)
				}
				mapping[index.index] = it
				v = mapping[index.index]
			}
			current = v
		case indexTypeArrayIndex, indexTypeQuery:
			it, ok := current.([]any)
			if !ok {
				return errors.Errorf("expected type `[]any`, got `%T`", current)
			}
			if index.indexType == indexTypeArrayIndex {
				rank, err := strconv.Atoi(index.index)
				if err != nil {
					return errors.Wrapf(err, "failed to convert index to int: %q", index.index)
				}
				if rank >= len(it) {
					panic("Slice is too small")
				}
				for rank >= len(it) {
					it = append(it, nil)
				}
				if end {
					it[rank] = value
					return nil
				}
				current = it[rank]
			} else {
				matchFound := false
				for _, e := range it {
					m, ok := e.(map[string]any)
					if !ok {
						return errors.Errorf("expected type map[string]any, got %T", current)
					}
					switch index.query.op {
					case queryOpEq:
						if v, ok := m[index.index]; ok && v == index.query.argument {
							current = m
							matchFound = true
							break
						}
					default:
						return errors.Errorf("unsupported query operation: %q", index.query.op)
					}
				}
				if !matchFound {
					query := index.index + string(index.query.op) + index.query.argument
					return errors.Errorf("no match found for query: %q", query)
				}
			}
		default:
			panic("BUG: there are no other types")
		}
	}
	return nil
}

type fieldPathIndexQuery struct {
	op       queryOp
	argument string
}

type fieldPathIndex struct {
	indexType indexType
	// index is either a map key (string) or a slice index (int). When the indexType
	// is Query, then index will be the query key in a map.
	index string

	// query is only defined when the indexType is type Query
	query fieldPathIndexQuery
}

func newFieldPathParser(fieldPath string) *fieldPathParser {
	parser := &fieldPathParser{fieldPath: fieldPath}
	parser.inc()
	return parser
}

// fieldPathParser parses a fieldPath string
// into indexes used for iterating on arbitrary json like
// data structured
type fieldPathParser struct {
	fieldPath string
	char      byte
	pos       int
	next      int
}

func (fp *fieldPathParser) inc() {
	if fp.next >= len(fp.fieldPath) {
		fp.char = 0
	} else {
		fp.char = fp.fieldPath[fp.next]
	}
	fp.pos = fp.next
	fp.next += 1
}

func (fp *fieldPathParser) peek() byte {
	if fp.pos >= len(fp.fieldPath) {
		return 0
	}
	return fp.fieldPath[fp.next]
}

func (fp *fieldPathParser) nextIndex() (index fieldPathIndex, err error) {

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
		index = fieldPathIndex{indexType: indexTypeMapKey, index: fp.fieldPath[pos:fp.pos]}
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
			var next fieldPathIndex
			next, err = fp.nextIndex()
			if err != nil {
				return
			}
			index.indexType = indexTypeQuery
			index.query.op = queryOpEq
			index.query.argument = next.index
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
			index = fieldPathIndex{indexType: indexTypeArrayIndex, index: fp.fieldPath[pos:fp.pos]}
			return
		case isLetter(fp.char):
			pos := fp.pos
			for isAlpha(fp.char) {
				fp.inc()
			}
			index = fieldPathIndex{indexType: indexTypeMapKey, index: fp.fieldPath[pos:fp.pos]}
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
