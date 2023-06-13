package dinghy

type (
	IndexType string
	QueryOp   string
)

const (
	IndexTypeQuery      IndexType = "Query"
	IndexTypeMapKey     IndexType = "MapKey"
	IndexTypeArrayIndex IndexType = "ArrayIndex"

	QueryOpCmpEqual = "="
)

type Query struct {
	op       QueryOp
	argument string
}

type Index struct {
	it IndexType
	// index is either a map key (string) or a slice index (int). When the index
	// is Query, then index will be the query key in a map.
	index string

	// query is only defined when the index is type Query
	query Query
}
