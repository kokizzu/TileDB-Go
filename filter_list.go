package tiledb

/*
#include <tiledb/tiledb.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

type filterListHandle struct{ *capiHandle }

func freeCapiFilterList(c unsafe.Pointer) {
	C.tiledb_filter_list_free((**C.tiledb_filter_list_t)(unsafe.Pointer(&c)))
}

func newFilterListHandle(ptr *C.tiledb_filter_list_t) filterListHandle {
	return filterListHandle{newCapiHandle(unsafe.Pointer(ptr), freeCapiFilterList)}
}

func (x filterListHandle) Get() *C.tiledb_filter_list_t {
	return (*C.tiledb_filter_list_t)(x.capiHandle.Get())
}

// FilterList represents
type FilterList struct {
	tiledbFilterList filterListHandle
	context          *Context
}

func newFilterListFromHandle(context *Context, handle filterListHandle) *FilterList {
	return &FilterList{tiledbFilterList: handle, context: context}
}

// Alloc a new FilterList
func NewFilterList(context *Context) (*FilterList, error) {
	var filterListPtr *C.tiledb_filter_list_t

	ret := C.tiledb_filter_list_alloc(context.tiledbContext.Get(), &filterListPtr)
	runtime.KeepAlive(context)
	if ret != C.TILEDB_OK {
		return nil, fmt.Errorf("error creating tiledb FilterList: %w", context.LastError())
	}

	return newFilterListFromHandle(context, newFilterListHandle(filterListPtr)), nil
}

// Free releases the internal TileDB core data that was allocated on the C heap.
// It is automatically called when this object is garbage collected, but can be
// called earlier to manually release memory if needed. Free is idempotent and
// can safely be called many times on the same object; if it has already
// been freed, it will not be freed again.
func (f *FilterList) Free() {
	f.tiledbFilterList.Free()
}

// Context exposes the internal TileDB context used to initialize the filter list
func (f *FilterList) Context() *Context {
	return f.context
}

// AddFilter appends a filter to a filter list. Data is processed through
// each filter in the order the filters were added.
func (f *FilterList) AddFilter(filter *Filter) error {
	ret := C.tiledb_filter_list_add_filter(f.context.tiledbContext.Get(), f.tiledbFilterList.Get(), filter.tiledbFilter.Get())
	runtime.KeepAlive(f)
	runtime.KeepAlive(filter)
	if ret != C.TILEDB_OK {
		return fmt.Errorf("error adding filter to tiledb FilterList: %w", f.context.LastError())
	}
	return nil
}

// SetMaxChunkSize sets the maximum tile chunk size for a filter list.
func (f *FilterList) SetMaxChunkSize(maxChunkSize uint32) error {
	ret := C.tiledb_filter_list_set_max_chunk_size(f.context.tiledbContext.Get(), f.tiledbFilterList.Get(), C.uint32_t(maxChunkSize))
	runtime.KeepAlive(f)
	if ret != C.TILEDB_OK {
		return fmt.Errorf("error setting max chunk size on tiledb FilterList: %w", f.context.LastError())
	}
	return nil
}

// MaxChunkSize Gets the maximum tile chunk size for a filter list.
func (f *FilterList) MaxChunkSize() (uint32, error) {
	var cMaxChunkSize C.uint32_t
	ret := C.tiledb_filter_list_get_max_chunk_size(f.context.tiledbContext.Get(), f.tiledbFilterList.Get(), &cMaxChunkSize)
	runtime.KeepAlive(f)
	if ret != C.TILEDB_OK {
		return 0, fmt.Errorf("error fetching max chunk size from tiledb FilterList: %w", f.context.LastError())
	}
	return uint32(cMaxChunkSize), nil
}

// NFilters Retrieves the number of filters in a filter list.
func (f *FilterList) NFilters() (uint32, error) {
	var cNFilters C.uint32_t
	ret := C.tiledb_filter_list_get_nfilters(f.context.tiledbContext.Get(), f.tiledbFilterList.Get(), &cNFilters)
	runtime.KeepAlive(f)
	if ret != C.TILEDB_OK {
		return 0, fmt.Errorf("error getting number of filter for tiledb FilterList: %w", f.context.LastError())
	}
	return uint32(cNFilters), nil
}

// FilterFromIndex Retrieves a filter object from a filter list by index.
func (f *FilterList) FilterFromIndex(index uint32) (*Filter, error) {
	var filterPtr *C.tiledb_filter_t
	ret := C.tiledb_filter_list_get_filter_from_index(f.context.tiledbContext.Get(), f.tiledbFilterList.Get(), C.uint32_t(index), &filterPtr)
	runtime.KeepAlive(f)
	if ret != C.TILEDB_OK {
		return nil, fmt.Errorf("error fetching filter for index %d from tiledb FilterList: %w", index, f.context.LastError())
	}

	return newFilterFromHandle(f.context, newFilterHandle(filterPtr)), nil
}

// Filters return slice of filters applied to filter list
func (f *FilterList) Filters() ([]*Filter, error) {
	var filters []*Filter
	nfilters, err := f.NFilters()
	if err != nil {
		return nil, err
	}

	for index := uint32(0); index < nfilters; index++ {
		filter, err := f.FilterFromIndex(index)
		if err != nil {
			return nil, err
		}
		filters = append(filters, filter)
	}
	return filters, err
}
