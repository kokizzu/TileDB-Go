/**
 * @file   rerading_range_test.go
 *
 * @section LICENSE
 *
 * The MIT License
 *
 * @copyright Copyright (c) 2021 TileDB, Inc.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 *
 * @section DESCRIPTION
 *
 * This is a part of the TileDB reading_range tutorial:
 *   https://docs.tiledb.io/en/latest/reading_range.html
 *
 * When run, this program will create a simple 2D dense array, write some data
 * to it, and read a slice of the data back in the layout of the user's choice
 * (passed as an argument to the program: "row", "col", or "global").
 *
 */

package examples

import "github.com/TileDB-Inc/TileDB-Go/examples_lib"

// ExampleRunReadRangeArray shows and example creation, writing and range reading
// of a dense array
func ExampleRunReadRangeArray() {
	examples_lib.RunReadRangeArray()

	// Output: Num of Ranges: 2
	// Range for dimension: 0, start: 1, end: 1
	// Range for dimension: 0, start: 3, end: 4
	// Ranges: map[cols:[{1 4}] rows:[{1 1} {3 4}]]
	// Num of Ranges: 2
	// Range for dimension: 1, start: 1, end: 1
	// Range for dimension: 1, start: 3, end: 4
	// Ranges: map[cols:[{1 1} {3 4}] rows:[{1 4}]]
}
