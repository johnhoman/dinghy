package kustomize

import (
	qt "github.com/frankban/quicktest"
	"testing"
)

func TestFieldPath_SetValue(t *testing.T) {
	cases := map[string]struct {
		fieldPath string
		value     any
		in        map[string]any
		out       map[string]any
	}{
		"SetsAValueInAMap": {
			fieldPath: "data",
			value:     "foo",
			in:        make(map[string]any),
			out: map[string]any{
				"data": "foo",
			},
		},
		"CreatesAValueInAMap": {
			fieldPath: "data",
			value: map[string]any{
				"data": map[string]any{
					"foo": "bar",
				},
			},
			in: make(map[string]any),
			out: map[string]any{
				"data": map[string]any{
					"data": map[string]any{
						"foo": "bar",
					},
				},
			},
		},
		"SetsAValueInANestedList": {
			fieldPath: "data[0]",
			value: map[string]any{
				"data": map[string]any{
					"foo": "bar",
				},
			},
			in: make(map[string]any),
			out: map[string]any{
				"data": []any{
					map[string]any{
						"data": map[string]any{
							"foo": "bar",
						},
					},
				},
			},
		},
		"SetsAValueUsingAQuery": {
			fieldPath: "data[name=main].foo",
			value: map[string]any{
				"data": map[string]any{
					"foo": "bar",
				},
			},
			in: map[string]any{
				"data": []any{
					map[string]any{
						"name": "foo",
					},
					map[string]any{
						"name": "main",
					},
				},
			},
			out: map[string]any{
				"data": []any{
					map[string]any{
						"name": "foo",
					},
					map[string]any{
						"name": "main",
						"foo": map[string]any{
							"data": map[string]any{
								"foo": "bar",
							},
						},
					},
				},
			},
		},
	}

	for name, testcase := range cases {
		t.Run(name, func(t *testing.T) {
			fp, err := NewFieldPath(testcase.fieldPath)
			qt.Assert(t, err, qt.IsNil)
			qt.Assert(t, fp.SetValue(testcase.in, testcase.value), qt.IsNil)
			qt.Assert(t, testcase.in, qt.DeepEquals, testcase.out)
		})
	}
}

func TestFieldPathParser(t *testing.T) {
	cases := map[string]struct {
		indexes []fieldPathIndex
	}{
		`100`: {
			indexes: []fieldPathIndex{
				{indexType: indexTypeArrayIndex, index: "100"},
			},
		},
		`data`: {
			indexes: []fieldPathIndex{
				{indexType: indexTypeMapKey, index: "data"},
			},
		},
		`data.data`: {
			indexes: []fieldPathIndex{
				{indexType: indexTypeMapKey, index: "data"},
				{indexType: indexTypeMapKey, index: "data"},
			},
		},
		`data[data].data`: {
			indexes: []fieldPathIndex{
				{indexType: indexTypeMapKey, index: "data"},
				{indexType: indexTypeMapKey, index: "data"},
				{indexType: indexTypeMapKey, index: "data"},
			},
		},
		`data['data.com/example']`: {
			indexes: []fieldPathIndex{
				{indexType: indexTypeMapKey, index: "data"},
				{indexType: indexTypeMapKey, index: "data.com/example"},
			},
		},
		`data[0].data`: {
			indexes: []fieldPathIndex{
				{indexType: indexTypeMapKey, index: "data"},
				{indexType: indexTypeArrayIndex, index: "0"},
				{indexType: indexTypeMapKey, index: "data"},
			},
		},
		`data["0"].data`: {
			indexes: []fieldPathIndex{
				{indexType: indexTypeMapKey, index: "data"},
				{indexType: indexTypeMapKey, index: "0"},
				{indexType: indexTypeMapKey, index: "data"},
			},
		},
		`data[name='main'].data`: {
			indexes: []fieldPathIndex{
				{indexType: indexTypeMapKey, index: "data"},
				{indexType: indexTypeQuery, index: "name", query: fieldPathIndexQuery{
					op:       queryOpEq,
					argument: "main",
				}},
				{indexType: indexTypeMapKey, index: "data"},
			},
		},
		`data[name=main].data`: {
			indexes: []fieldPathIndex{
				{indexType: indexTypeMapKey, index: "data"},
				{indexType: indexTypeQuery, index: "name", query: fieldPathIndexQuery{
					op:       queryOpEq,
					argument: "main",
				}},
				{indexType: indexTypeMapKey, index: "data"},
			},
		},
		`data[name="main"].data`: {
			indexes: []fieldPathIndex{
				{indexType: indexTypeMapKey, index: "data"},
				{indexType: indexTypeQuery, index: "name", query: fieldPathIndexQuery{
					op:       queryOpEq,
					argument: "main",
				}},
				{indexType: indexTypeMapKey, index: "data"},
			},
		},
	}
	for fieldPath, testcase := range cases {
		t.Run(fieldPath, func(t *testing.T) {
			parser := newFieldPathParser(fieldPath)
			for _, index := range testcase.indexes {
				next, err := parser.nextIndex()
				qt.Assert(t, err, qt.IsNil)
				qt.Assert(t, next.indexType, qt.DeepEquals, index.indexType)
				qt.Assert(t, next.index, qt.DeepEquals, index.index)
				qt.Assert(t, next.query.op, qt.DeepEquals, index.query.op)
				qt.Assert(t, next.query.argument, qt.DeepEquals, index.query.argument)
			}
		})
	}
}
