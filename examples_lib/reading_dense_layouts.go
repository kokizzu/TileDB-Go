package examples_lib

import (
	"fmt"

	tiledb "github.com/TileDB-Inc/TileDB-Go"
)

func createReadingDenseLayoutsArray(dir string) {
	// Create a TileDB context.
	ctx, err := tiledb.NewContext(nil)
	checkError(err)
	defer ctx.Free()

	// The array will be 4x4 with dimensions "rows" and "cols", with domain [1,4].
	domain, err := tiledb.NewDomain(ctx)
	checkError(err)
	defer domain.Free()

	rowDim, err := tiledb.NewDimension(ctx, "rows", tiledb.TILEDB_INT32, []int32{1, 4}, int32(2))
	checkError(err)
	defer rowDim.Free()

	colDim, err := tiledb.NewDimension(ctx, "cols", tiledb.TILEDB_INT32, []int32{1, 4}, int32(2))
	checkError(err)
	defer colDim.Free()

	err = domain.AddDimensions(rowDim, colDim)
	checkError(err)

	// The array will be dense.
	schema, err := tiledb.NewArraySchema(ctx, tiledb.TILEDB_DENSE)
	checkError(err)
	defer schema.Free()

	err = schema.SetDomain(domain)
	checkError(err)
	err = schema.SetCellOrder(tiledb.TILEDB_ROW_MAJOR)
	checkError(err)
	err = schema.SetTileOrder(tiledb.TILEDB_ROW_MAJOR)
	checkError(err)

	// Add a single attribute "a" so each (i,j) cell can store an integer.
	a, err := tiledb.NewAttribute(ctx, "a", tiledb.TILEDB_INT32)
	checkError(err)
	defer a.Free()

	err = schema.AddAttributes(a)
	checkError(err)

	// Create the (empty) array on disk.
	err = tiledb.CreateArray(ctx, dir, schema)
	checkError(err)
}

func writeReadingDenseLayoutsArray(dir string) {
	ctx, err := tiledb.NewContext(nil)
	checkError(err)
	defer ctx.Free()

	// Prepare some data for the array
	data := []int32{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	// Open the array for writing and create the query.
	array, err := tiledb.NewArray(ctx, dir)
	checkError(err)
	defer array.Free()

	err = array.Open(tiledb.TILEDB_WRITE)
	checkError(err)
	defer array.Close()

	query, err := tiledb.NewQuery(ctx, array)
	checkError(err)
	defer query.Free()

	err = query.SetLayout(tiledb.TILEDB_GLOBAL_ORDER)
	checkError(err)
	_, err = query.SetDataBuffer("a", data)
	checkError(err)

	// Perform the write and close the array.
	err = query.Submit()
	checkError(err)

	err = query.Finalize()
	checkError(err)
}

func readReadingDenseLayoutsArray(dir string) {
	ctx, err := tiledb.NewContext(nil)
	checkError(err)
	defer ctx.Free()

	// Prepare the array for reading
	array, err := tiledb.NewArray(ctx, dir)
	checkError(err)
	defer array.Free()

	err = array.Open(tiledb.TILEDB_READ)
	checkError(err)
	defer array.Close()

	// Non-empty domain: [1,4], [1,4]
	x, isEmpty, err := array.NonEmptyDomain()
	checkError(err)
	if !isEmpty {
		rows := x[0].Bounds.([]int32)
		cols := x[1].Bounds.([]int32)
		fmt.Printf("Non-empty domain: [%d,%d], [%d,%d]\n",
			rows[0], rows[1], cols[0], cols[1])
	}

	// Slice only rows 1, 2 and cols 2, 3, 4
	subArray := []int32{1, 2, 2, 4}

	// Prepare the vector that will hold the result (of size 6 elements)
	data := make([]int32, 6)
	rows := make([]int32, 6)
	cols := make([]int32, 6)

	// Prepare the query
	query, err := tiledb.NewQuery(ctx, array)
	checkError(err)
	defer query.Free()

	// Prepare the subarray
	subarray, err := array.NewSubarray()
	checkError(err)
	defer subarray.Free()

	err = subarray.SetSubArray(subArray)
	checkError(err)
	err = query.SetSubarray(subarray)
	checkError(err)
	err = query.SetLayout(tiledb.TILEDB_ROW_MAJOR)
	checkError(err)
	_, err = query.SetDataBuffer("a", data)
	checkError(err)
	_, err = query.SetDataBuffer("rows", rows)
	checkError(err)
	_, err = query.SetDataBuffer("cols", cols)
	checkError(err)

	// Submit the query and close the array.
	err = query.Submit()
	checkError(err)

	// Print out the results.
	elements, err := query.ResultBufferElements()
	checkError(err)
	resultNum := elements["a"][1]
	for r := 0; r < int(resultNum); r++ {
		i := rows[r]
		j := cols[r]
		a := data[r]
		fmt.Printf("Cell (%d, %d) has data %d\n", i, j, a)
	}

	err = query.Finalize()
	checkError(err)
}

func RunReadingDenseLayouts() {
	tmpDir := temp("reading_dense_layouts")
	defer cleanup(tmpDir)

	createReadingDenseLayoutsArray(tmpDir)
	writeReadingDenseLayoutsArray(tmpDir)
	readReadingDenseLayoutsArray(tmpDir)
}
