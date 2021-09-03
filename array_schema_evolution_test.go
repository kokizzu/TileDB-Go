//go:build experimental
// +build experimental

package tiledb

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const array_schema_evolution_name = "array_schema_evolution"

func TestArraySchemaEvolution(t *testing.T) {
	config, err := NewConfig()
	require.NoError(t, err)
	assert.NotNil(t, config)

	// Test context with config
	context, err := NewContext(config)
	require.NoError(t, err)
	assert.NotNil(t, context)

	dimension, err := NewDimension(context, "test", TILEDB_INT32, []int32{1, 10}, int32(5))
	require.NoError(t, err)
	assert.NotNil(t, dimension)

	domain, err := NewDomain(context)
	require.NoError(t, err)
	assert.NotNil(t, domain)

	// Add dimension to domain
	err = domain.AddDimensions(dimension)
	require.NoError(t, err)

	arraySchema, err := NewArraySchema(context, TILEDB_DENSE)
	require.NoError(t, err)
	assert.NotNil(t, arraySchema)

	a1, err := NewAttribute(context, "a1", TILEDB_INT32)
	require.NoError(t, err)
	assert.NotNil(t, a1)

	a2, err := NewAttribute(context, "a2", TILEDB_INT32)
	require.NoError(t, err)
	assert.NotNil(t, a2)

	err = arraySchema.AddAttributes(a1, a2)
	require.NoError(t, err)

	err = arraySchema.SetCapacity(100)
	require.NoError(t, err)

	err = arraySchema.SetDomain(domain)
	require.NoError(t, err)

	err = arraySchema.Check()
	require.NoError(t, err)

	// tmpArrayPath is the array URI
	tmpArrayPath := path.Join(os.TempDir(), array_schema_evolution_name)
	if _, err = os.Stat(tmpArrayPath); err == nil {
		os.RemoveAll(tmpArrayPath)
	}

	array, err := NewArray(context, tmpArrayPath)
	require.NoError(t, err)
	assert.NotNil(t, array)

	defer os.RemoveAll(tmpArrayPath)

	err = array.Create(arraySchema)
	require.NoError(t, err)

	err = array.Close()
	require.NoError(t, err)

	arraySchemaEvolution, err := NewArraySchemaEvolution(context)
	require.NoError(t, err)

	a3, err := NewAttribute(context, "a3", TILEDB_INT32)
	require.NoError(t, err)
	assert.NotNil(t, a2)

	err = arraySchemaEvolution.AddAttribute(a3)
	require.NoError(t, err)

	// Will fail when try to add an attribute which already has the name
	err = arraySchemaEvolution.AddAttribute(a3)
	require.Error(t, err)

	// Remove atrribute a1
	err = arraySchemaEvolution.DropAttribute("a1")
	require.NoError(t, err)

	buffer, err := SerializeArraySchemaEvolution(arraySchemaEvolution,
		TILEDB_CAPNP, true)
	require.NoError(t, err)

	newArraySchemaEvolution, err := DeserializeArraySchemaEvolution(buffer,
		TILEDB_CAPNP, true)
	require.NoError(t, err)

	err = newArraySchemaEvolution.Evolve(tmpArrayPath)
	require.NoError(t, err)

	// Validate schema evolution changes
	ctx, err := NewContext(nil)
	require.NoError(t, err)
	defer ctx.Free()

	// Prepare the array for reading
	arr, err := NewArray(ctx, tmpArrayPath)
	require.NoError(t, err)
	defer array.Free()

	err = arr.Open(TILEDB_READ)
	require.NoError(t, err)

	// Need to get the evolved schema
	arrAchema, err := arr.Schema()
	require.NoError(t, err)
	assert.NotNil(t, arrAchema)

	hasAttr, err := arrAchema.HasAttribute("a1")
	require.NoError(t, err)
	assert.False(t, hasAttr)

	hasAttr, err = arrAchema.HasAttribute("a2")
	require.NoError(t, err)
	assert.True(t, hasAttr)

	hasAttr, err = arrAchema.HasAttribute("a3")
	require.NoError(t, err)
	assert.True(t, hasAttr)

	attrNum, err := arrAchema.AttributeNum()
	require.NoError(t, err)
	assert.Equal(t, uint(2), attrNum)

	attrFromIndex, err := arrAchema.AttributeFromIndex(0)
	require.NoError(t, err)
	assert.NotNil(t, attrFromIndex)

	attrName, err := attrFromIndex.Name()
	require.NoError(t, err)
	assert.Equal(t, "a2", attrName)

	attrFromName, err := arrAchema.AttributeFromName("a3")
	require.NoError(t, err)
	assert.NotNil(t, attrFromName)

	attrName2, err := attrFromName.Name()
	require.NoError(t, err)
	assert.Equal(t, "a3", attrName2)

	err = arr.Close()
	require.NoError(t, err)
}
