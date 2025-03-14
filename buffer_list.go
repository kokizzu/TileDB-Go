package tiledb

/*
#include <tiledb/tiledb.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"io"
	"runtime"
)

// BufferList A list of TileDB BufferList objects
type BufferList struct {
	tiledbBufferList *C.tiledb_buffer_list_t
	context          *Context
}

// NewBufferList Allocs a new buffer list
func NewBufferList(context *Context) (*BufferList, error) {
	bufferList := BufferList{context: context}

	ret := C.tiledb_buffer_list_alloc(bufferList.context.tiledbContext, &bufferList.tiledbBufferList)
	runtime.KeepAlive(context)
	if ret != C.TILEDB_OK {
		return nil, fmt.Errorf("error creating tiledb buffer list: %w", bufferList.context.LastError())
	}
	freeOnGC(&bufferList)

	return &bufferList, nil
}

// Free releases the internal TileDB core data that was allocated on the C heap.
// It is automatically called when this object is garbage collected, but can be
// called earlier to manually release memory if needed. Free is idempotent and
// can safely be called many times on the same object; if it has already
// been freed, it will not be freed again.
func (b *BufferList) Free() {
	if b.tiledbBufferList != nil {
		C.tiledb_buffer_list_free(&b.tiledbBufferList)
	}
}

// Context exposes the internal TileDB context used to initialize the buffer list.
func (b *BufferList) Context() *Context {
	return b.context
}

// WriteTo writes the contents of a BufferList to an io.Writer.
func (b *BufferList) WriteTo(w io.Writer) (int64, error) {
	nbuffs, err := b.NumBuffers()
	if err != nil {
		return 0, err
	}

	written := int64(0)

	for i := uint(0); i < uint(nbuffs); i++ {
		buff, err := b.GetBuffer(i)
		if err != nil {
			return 0, err
		}
		n, err := buff.WriteTo(w)
		written += n

		buff.Free()

		if err != nil {
			return written, err
		}
	}

	return written, nil
}

// Static assert that BufferList implements io.WriterTo.
var _ io.WriterTo = (*BufferList)(nil)

// NumBuffers returns number of buffers in the list.
func (b *BufferList) NumBuffers() (uint64, error) {
	var numBuffers C.uint64_t
	ret := C.tiledb_buffer_list_get_num_buffers(b.context.tiledbContext, b.tiledbBufferList, &numBuffers)
	runtime.KeepAlive(b)

	if ret != C.TILEDB_OK {
		return 0, fmt.Errorf("error getting tiledb bufferList num buffers: %w", b.context.LastError())
	}

	return uint64(numBuffers), nil
}

// GetBuffer returns a Buffer at the given index in the list.
func (b *BufferList) GetBuffer(bufferIndex uint) (*Buffer, error) {
	buffer := Buffer{context: b.context}

	ret := C.tiledb_buffer_list_get_buffer(b.context.tiledbContext, b.tiledbBufferList, C.uint64_t(bufferIndex), &buffer.tiledbBuffer)
	runtime.KeepAlive(b)
	if ret != C.TILEDB_OK {
		return nil, fmt.Errorf("error getting tiledb buffer index %d from buffer list: %w", bufferIndex, b.context.LastError())
	}
	freeOnGC(&buffer)

	return &buffer, nil
}

// TotalSize returns the total number of bytes in the buffers in the list.
func (b *BufferList) TotalSize() (uint64, error) {
	var totalSize C.uint64_t
	ret := C.tiledb_buffer_list_get_total_size(b.context.tiledbContext, b.tiledbBufferList, &totalSize)
	runtime.KeepAlive(b)

	if ret != C.TILEDB_OK {
		return 0, fmt.Errorf("error getting tiledb bufferList num buffers: %w", b.context.LastError())
	}

	return uint64(totalSize), nil
}

// Flatten copies and concatenates all buffers in the list into a new buffer.
//
// Deprecated: Use WriteTo instead for increased performance.
func (b *BufferList) Flatten() (*Buffer, error) {
	buffer := Buffer{context: b.context}

	ret := C.tiledb_buffer_list_flatten(b.context.tiledbContext, b.tiledbBufferList, &buffer.tiledbBuffer)
	runtime.KeepAlive(b)

	if ret != C.TILEDB_OK {
		return nil, fmt.Errorf("error getting tiledb bufferList num buffers: %w", b.context.LastError())
	}
	freeOnGC(&buffer)

	return &buffer, nil
}
