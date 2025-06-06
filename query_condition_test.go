package tiledb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testAttributeValues = struct {
	Attribute1             []int32
	Attribute1Buffer       []int32
	Attribute2             []byte
	Attribute2Offset       []uint64
	Attribute2DataBuffer   []byte
	Attribute2OffsetBuffer []uint64
	Attribute3             []int64
	Attribute3Buffer       []int64
}{
	Attribute1:             []int32{1, 2, 3},
	Attribute1Buffer:       make([]int32, 3),
	Attribute2:             []byte("iamastring"),
	Attribute2Offset:       []uint64{0, 1, 4},
	Attribute2DataBuffer:   make([]byte, 10),
	Attribute2OffsetBuffer: make([]uint64, 3),
	Attribute3:             []int64{1623763941, 1623762932, 1623765583},
	Attribute3Buffer:       make([]int64, 3),
}

func TestQueryCondition(t *testing.T) {
	array, err := createBasicTestArray(t)
	if err != nil {
		t.Errorf("failed to create basic test array: %s", err)
	}

	if err := array.Open(TILEDB_READ); err != nil {
		t.Errorf("failed to open test array for reading: %s", err)
	}

	testQueryConditionInt32(t, array)
	testQueryConditionBytes(t, array)
	testQueryConditionTime(t, array)
}

func testQueryConditionInt32(t *testing.T, array *Array) {
	a1Cases := []struct {
		name           string
		opValue        int32
		op             QueryConditionOp
		negate         bool
		expectedValues []int32
	}{
		{"GreaterThan1", 1, TILEDB_QUERY_CONDITION_GT, false, []int32{2, 3}},
		{"NotGreaterThan1", 1, TILEDB_QUERY_CONDITION_GT, true, []int32{1}},
		{"LessThan3", 3, TILEDB_QUERY_CONDITION_LT, false, []int32{1, 2}},
		{"NotLessThan3", 3, TILEDB_QUERY_CONDITION_LT, true, []int32{3}},
		{"EqualTo2", 2, TILEDB_QUERY_CONDITION_EQ, false, []int32{2}},
		{"NotEqualTo2", 2, TILEDB_QUERY_CONDITION_EQ, true, []int32{1, 3}},
	}
	for _, c := range a1Cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			a1DataRead := make([]int32, 3)
			// Prepare the query
			query, err := NewQuery(array.context, array)
			require.NoError(t, err)
			assert.NotNil(t, query)

			_, err = query.SetDataBuffer("a1", a1DataRead)
			require.NoError(t, err)
			assert.NotNil(t, query)

			qc, err := NewQueryCondition(array.context, "a1", c.op, c.opValue)
			require.NoError(t, err)

			if c.negate {
				qc, err = NewQueryConditionNegated(array.context, qc)
				require.NoError(t, err)
			}

			err = query.SetQueryCondition(qc)
			require.NoError(t, err)

			err = query.SetLayout(TILEDB_ROW_MAJOR)
			require.NoError(t, err)

			// Submit the query
			err = query.Submit()
			require.NoError(t, err)
			// compare the elements in the buffer to the expected values
			elements, err := query.ResultBufferElements()
			require.NoError(t, err)

			for i := uint64(0); i < elements["a1"][1]; i++ {
				assert.Equal(t, a1DataRead[i], c.expectedValues[i])
			}
			query.Free()
		})
	}

	a1CombinationCases := []struct {
		name                 string
		op1                  QueryConditionOp
		op1Value             int32
		op2                  QueryConditionOp
		op2Value             int32
		combinationCondition QueryConditionCombinationOp
		negate               bool
		expectedValues       []int32
	}{
		{
			name:                 "GreaterThan1AndLessThan3",
			op1:                  TILEDB_QUERY_CONDITION_GT,
			op1Value:             1,
			op2:                  TILEDB_QUERY_CONDITION_LT,
			op2Value:             3,
			combinationCondition: TILEDB_QUERY_CONDITION_AND,
			negate:               false,
			expectedValues:       []int32{2},
		},
		{
			name:                 "NotGreaterThan1AndLessThan3",
			op1:                  TILEDB_QUERY_CONDITION_GT,
			op1Value:             1,
			op2:                  TILEDB_QUERY_CONDITION_LT,
			op2Value:             3,
			combinationCondition: TILEDB_QUERY_CONDITION_AND,
			negate:               true,
			expectedValues:       []int32{1, 3},
		},
	}
	for _, c := range a1CombinationCases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			a1DataRead := make([]int32, 3)
			// Prepare the query
			query, err := NewQuery(array.context, array)
			require.NoError(t, err)
			assert.NotNil(t, query)

			_, err = query.SetDataBuffer("a1", a1DataRead)
			require.NoError(t, err)
			assert.NotNil(t, query)

			qc1, err := NewQueryCondition(array.context, "a1", c.op1, c.op1Value)
			require.NoError(t, err)

			qc2, err := NewQueryCondition(array.context, "a1", c.op2, c.op2Value)
			require.NoError(t, err)

			qc, err := NewQueryConditionCombination(array.context, qc1, c.combinationCondition, qc2)
			require.NoError(t, err)

			if c.negate {
				qc, err = NewQueryConditionNegated(array.context, qc)
				require.NoError(t, err)
			}

			err = query.SetQueryCondition(qc)
			require.NoError(t, err)

			err = query.SetLayout(TILEDB_ROW_MAJOR)
			require.NoError(t, err)

			// Submit the query
			err = query.Submit()
			require.NoError(t, err)
			// compare the elements in the buffer to the expected values
			elements, err := query.ResultBufferElements()
			require.NoError(t, err)

			for i := uint64(0); i < elements["a1"][1]; i++ {
				assert.Equal(t, a1DataRead[i], c.expectedValues[i])
			}
			query.Free()
		})
	}
}

func testQueryConditionTime(t *testing.T, array *Array) {
	a3Cases := []struct {
		name           string
		opValue        int64
		op             QueryConditionOp
		expectedValues []int64
	}{
		{"EqualTo1623762932", 1623762932, TILEDB_QUERY_CONDITION_EQ, []int64{1623762932}},
		{"LessThanEqualTo1623765583", 1623765583, TILEDB_QUERY_CONDITION_LE, []int64{1623763941, 1623762932, 1623765583}},
	}
	for _, c := range a3Cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			a3DataRead := make([]int64, 3)
			// Prepare the query
			query, err := NewQuery(array.context, array)
			require.NoError(t, err)
			assert.NotNil(t, query)

			_, err = query.SetDataBuffer("a3", a3DataRead)
			require.NoError(t, err)
			assert.NotNil(t, query)
			qc, err := NewQueryCondition(array.context, "a3", c.op, c.opValue)
			require.NoError(t, err)

			err = query.SetQueryCondition(qc)
			require.NoError(t, err)

			err = query.SetLayout(TILEDB_ROW_MAJOR)
			require.NoError(t, err)

			// Submit the query
			err = query.Submit()
			require.NoError(t, err)

			// compare the elements in the buffer to the expected values
			elements, err := query.ResultBufferElements()
			require.NoError(t, err)

			for i := uint64(0); i < elements["a3"][1]; i++ {
				assert.Equal(t, a3DataRead[i], c.expectedValues[i])
			}
			query.Free()
		})
	}
}

func testQueryConditionBytes(t *testing.T, array *Array) {
	a2Cases := []struct {
		name           string
		opValue        []byte
		op             QueryConditionOp
		expectedValues []byte
	}{
		{"EqualToI", []byte("i"), TILEDB_QUERY_CONDITION_EQ, []byte("i")},
		{"NotEqualToI", []byte("i"), TILEDB_QUERY_CONDITION_NE, []byte("amastring")},
	}
	for _, c := range a2Cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			a2DataRead := make([]byte, len(c.expectedValues))
			a2OffsetRead := make([]uint64, 3)
			// Prepare the query
			query, err := NewQuery(array.context, array)
			require.NoError(t, err)
			assert.NotNil(t, query)

			_, err = query.SetDataBuffer("a2", a2DataRead)
			require.NoError(t, err)
			_, err = query.SetOffsetsBuffer("a2", a2OffsetRead)
			require.NoError(t, err)
			assert.NotNil(t, query)
			qc, err := NewQueryCondition(array.context, "a2", c.op, c.opValue)
			require.NoError(t, err)

			err = query.SetQueryCondition(qc)
			require.NoError(t, err)

			err = query.SetLayout(TILEDB_ROW_MAJOR)
			require.NoError(t, err)

			// Submit the query
			err = query.Submit()
			require.NoError(t, err)
			// compare the elements in the buffer to the expected values
			elements, err := query.ResultBufferElements()
			require.NoError(t, err)

			for i := uint64(0); i < elements["a2"][1]; i++ {
				assert.Equal(t, a2DataRead[i], c.expectedValues[i])
			}
			query.Free()
		})
	}
}

func createBasicTestArray(t testing.TB) (*Array, error) {
	// Create configuration
	config, err := NewConfig()
	if err != nil {
		return nil, err
	}

	context, err := NewContext(config)
	if err != nil {
		return nil, err
	}

	domain, err := createDomain(context)
	if err != nil {
		return nil, err
	}

	attributes, err := createAttributes(context)
	if err != nil {
		return nil, err
	}

	arraySchema, err := createSchema(context, domain, attributes)
	if err != nil {
		return nil, err
	}

	// create temp group name
	tmpArrayPath := t.TempDir()

	// Prepare some data for the array
	buffD1 := []int32{1, 2, 2}
	buffD2 := []int32{1, 1, 2}

	// Create array on disk
	if err = CreateArray(context, tmpArrayPath, arraySchema); err != nil {
		return nil, err
	}

	// Create new array struct
	array, err := NewArray(context, tmpArrayPath)
	if err != nil {
		return nil, err
	}

	if err := array.Open(TILEDB_WRITE); err != nil {
		return nil, err
	}
	query, err := NewQuery(context, array)
	if err != nil {
		return nil, err
	}
	if err := query.SetLayout(TILEDB_UNORDERED); err != nil {
		return nil, err
	}
	if _, err = query.SetDataBuffer("a1", testAttributeValues.Attribute1); err != nil {
		return nil, err
	}
	if _, err = query.SetDataBuffer("a2", testAttributeValues.Attribute2); err != nil {
		return nil, err
	}
	if _, err = query.SetOffsetsBuffer("a2", testAttributeValues.Attribute2Offset); err != nil {
		return nil, err
	}
	if _, err = query.SetDataBuffer("a3", testAttributeValues.Attribute3); err != nil {
		return nil, err
	}
	if _, err := query.SetDataBuffer("rows", buffD1); err != nil {
		return nil, err
	}
	if _, err := query.SetDataBuffer("cols", buffD2); err != nil {
		return nil, err
	}

	// Perform the write, finalize and close the array.
	if err := query.Submit(); err != nil {
		return nil, err
	}
	if err = query.Finalize(); err != nil {
		return nil, err
	}
	if err = array.Close(); err != nil {
		return nil, err
	}
	return array, nil
}

func createSchema(context *Context, domain *Domain, attributes []*Attribute) (*ArraySchema, error) {
	// Create array schema
	arraySchema, err := NewArraySchema(context, TILEDB_SPARSE)
	if err != nil {
		return nil, err
	}

	if err := arraySchema.SetCellOrder(TILEDB_ROW_MAJOR); err != nil {
		return nil, err
	}

	if err := arraySchema.SetTileOrder(TILEDB_ROW_MAJOR); err != nil {
		return nil, err
	}

	// Add Attribute
	if err := arraySchema.AddAttributes(attributes...); err != nil {
		return nil, err
	}

	// Set Domain
	if err := arraySchema.SetDomain(domain); err != nil {
		return nil, err
	}

	// Validate Schema
	if err := arraySchema.Check(); err != nil {
		return nil, err
	}

	return arraySchema, nil
}

func createAttributes(context *Context) ([]*Attribute, error) {
	// Create attribute to add to schema
	a1, err := NewAttribute(context, "a1", TILEDB_INT32)
	if err != nil {
		return nil, err
	}
	a2, err := NewAttribute(context, "a2", TILEDB_STRING_ASCII)
	if err != nil {
		return nil, err
	}
	err = a2.SetCellValNum(TILEDB_VAR_NUM)
	if err != nil {
		return nil, err
	}
	a3, err := NewAttribute(context, "a3", TILEDB_DATETIME_SEC)
	if err != nil {
		return nil, err
	}

	return []*Attribute{a1, a2, a3}, nil
}

func createDomain(context *Context) (*Domain, error) {
	// Test create row dimension
	rowDim, err := NewDimension(context, "rows", TILEDB_INT32, []int32{1, 4}, int32(2))
	if err != nil {
		return nil, err
	}

	// Test create row dimension
	colDim, err := NewDimension(context, "cols", TILEDB_INT32, []int32{1, 4}, int32(2))
	if err != nil {
		return nil, err
	}

	// Test creating domain
	domain, err := NewDomain(context)
	if err != nil {
		return nil, err
	}

	// Add dimensions
	if err := domain.AddDimensions(rowDim, colDim); err != nil {
		return nil, err
	}
	return domain, nil
}
