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

// ConfigIter creates a config iterator object.
type ConfigIter struct {
	config           *Config
	tiledbConfigIter *C.tiledb_config_iter_t
}

// NewConfigIter creates an iterator for configuration. This can be used
// only for reading. This sets the pointer to the first search item.
func NewConfigIter(config *Config, prefix string) (*ConfigIter, error) {
	ci := ConfigIter{config: config}
	var err *C.tiledb_error_t
	cprefix := C.CString(prefix)
	defer C.free(unsafe.Pointer(cprefix))
	C.tiledb_config_iter_alloc(config.tiledbConfig, cprefix, &ci.tiledbConfigIter, &err)
	runtime.KeepAlive(config)
	if err != nil {
		defer C.tiledb_error_free(&err)
		return nil, fmt.Errorf("error creating tiledb config iter: %w", cError(err))
	}
	freeOnGC(&ci)

	return &ci, nil
}

// Free releases the internal TileDB core data that was allocated on the C heap.
// It is automatically called when this object is garbage collected, but can be
// called earlier to manually release memory if needed. Free is idempotent and
// can safely be called many times on the same object; if it has already
// been freed, it will not be freed again.
func (ci *ConfigIter) Free() {
	if ci.tiledbConfigIter != nil {
		C.tiledb_config_iter_free(&ci.tiledbConfigIter)
	}
}

// Here retrieves the param and value for the item currently pointed to by the
// iterator.
func (ci *ConfigIter) Here() (*string, *string, error) {
	var err *C.tiledb_error_t
	var cparam, cvalue *C.char // ci must be kept alive while these are being accessed.
	C.tiledb_config_iter_here(ci.tiledbConfigIter, &cparam, &cvalue, &err)
	if err != nil {
		defer C.tiledb_error_free(&err)
		return nil, nil, fmt.Errorf("error getting param, value from config iter: %w", cError(err))
	}
	param := C.GoString(cparam)
	value := C.GoString(cvalue)
	runtime.KeepAlive(ci)
	return &param, &value, nil
}

// Next moves the iterator to the next item.
func (ci *ConfigIter) Next() error {
	var err *C.tiledb_error_t
	C.tiledb_config_iter_next(ci.tiledbConfigIter, &err)
	runtime.KeepAlive(ci)
	if err != nil {
		defer C.tiledb_error_free(&err)
		return fmt.Errorf("error moving to next ConfigItem from iter: %w", cError(err))
	}
	return nil
}

// Done checks if the iterator is done.
func (ci *ConfigIter) Done() (bool, error) {
	var err *C.tiledb_error_t
	var cDone C.int32_t
	C.tiledb_config_iter_done(ci.tiledbConfigIter, &cDone, &err)
	runtime.KeepAlive(ci)
	if err != nil {
		defer C.tiledb_error_free(&err)
		return false, fmt.Errorf("error moving to next ConfigItem from iter: %w", cError(err))
	}
	return int(cDone) == 1, nil
}

// IsDone checks if the iterator is done.
func (ci *ConfigIter) IsDone() bool {
	var err *C.tiledb_error_t
	var cDone C.int32_t
	C.tiledb_config_iter_done(ci.tiledbConfigIter, &cDone, &err)
	runtime.KeepAlive(ci)
	if err != nil {
		C.tiledb_error_free(&err)
		return false
	}
	return int(cDone) == 1
}

// Reset resets the config iterator.
func (ci *ConfigIter) Reset(prefix string) error {
	var err *C.tiledb_error_t
	cprefix := C.CString(prefix)
	defer C.free(unsafe.Pointer(cprefix))
	C.tiledb_config_iter_reset(ci.config.tiledbConfig, ci.tiledbConfigIter, cprefix, &err)
	runtime.KeepAlive(ci)
	if err != nil {
		defer C.tiledb_error_free(&err)
		return fmt.Errorf("error creating tiledb config iter: %w", cError(err))
	}
	return nil
}
