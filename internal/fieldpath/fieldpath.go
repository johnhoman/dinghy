package fieldpath

import (
	"github.com/pkg/errors"
	"io"
	"strconv"
)

func parseFieldPath(fieldPath string) (*FieldPath, error) {
	indexes := make([]Index, 0)
	parser := newParser(fieldPath)
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
	indexes   []Index
	fieldPath string
}

func (fp *FieldPath) SetValue(m map[string]any, value any) error {
	if len(fp.indexes) == 0 {
		return nil
	}

	var current any = m
	for k, index := range fp.indexes {
		end := len(fp.indexes)-1 == k

		switch index.it {
		case IndexTypeMapKey:
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
				if next.it == IndexTypeMapKey {
					mapping[index.index] = make(map[string]any)
				} else {
					mapping[index.index] = make([]any, 0)
				}
				v = mapping[index.index]
			}
			// if the next element is an array, make sure it's big enough. We
			// won't be able to change the size later because append
			if it, ok := v.([]any); ok && next.it != IndexTypeQuery {
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
		case IndexTypeArrayIndex, IndexTypeQuery:
			it, ok := current.([]any)
			if !ok {
				return errors.Errorf("expected type `[]any`, got `%T`", current)
			}
			if index.it == IndexTypeArrayIndex {
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
					case QueryOpCmpEqual:
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
