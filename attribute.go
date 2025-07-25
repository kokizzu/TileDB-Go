package tiledb

/*
#include <tiledb/tiledb.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"os"
	"runtime"
	"unsafe"
)

type attributeHandle struct{ *capiHandle }

func freeCapiAttribute(c unsafe.Pointer) {
	C.tiledb_attribute_free((**C.tiledb_attribute_t)(unsafe.Pointer(&c)))
}

func newAttributeHandle(ptr *C.tiledb_attribute_t) attributeHandle {
	return attributeHandle{newCapiHandle(unsafe.Pointer(ptr), freeCapiAttribute)}
}

func (x attributeHandle) Get() *C.tiledb_attribute_t {
	return (*C.tiledb_attribute_t)(x.capiHandle.Get())
}

/*
Attribute describes an attribute of an Array cell.

An attribute specifies a name and datatype for a particular value in each array cell. There are 3 supported attribute types:

	Fundamental types, such as char, int, double, uint64, etc..
	Fixed sized arrays: [N]T or make([]T, N), where T is a fundamental type
	Variable length data: string, []T, where T is a fundamental type
*/
type Attribute struct {
	tiledbAttribute attributeHandle
	context         *Context
}

func newAttributeFromHandle(context *Context, handle attributeHandle) *Attribute {
	return &Attribute{tiledbAttribute: handle, context: context}
}

// NewAttribute allocates a new attribute.
func NewAttribute(context *Context, name string, datatype Datatype) (*Attribute, error) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	var attributePtr *C.tiledb_attribute_t
	ret := C.tiledb_attribute_alloc(context.tiledbContext.Get(), cname, C.tiledb_datatype_t(datatype), &attributePtr)
	runtime.KeepAlive(context)
	if ret != C.TILEDB_OK {
		return nil, fmt.Errorf("error creating tiledb attribute: %w", context.LastError())
	}

	return newAttributeFromHandle(context, newAttributeHandle(attributePtr)), nil
}

// Free releases the internal TileDB core data that was allocated on the C heap.
// It is automatically called when this object is garbage collected, but can be
// called earlier to manually release memory if needed. Free is idempotent and
// can safely be called many times on the same object; if it has already
// been freed, it will not be freed again.
func (a *Attribute) Free() {
	a.tiledbAttribute.Free()
}

// Context exposes the internal TileDB context used to initialize the attribute.
func (a *Attribute) Context() *Context {
	return a.context
}

// SetFilterList sets the attribute filterList.
func (a *Attribute) SetFilterList(filterlist *FilterList) error {
	ret := C.tiledb_attribute_set_filter_list(a.context.tiledbContext.Get(), a.tiledbAttribute.Get(), filterlist.tiledbFilterList.Get())
	runtime.KeepAlive(a)
	runtime.KeepAlive(filterlist)
	if ret != C.TILEDB_OK {
		return fmt.Errorf("error setting tiledb attribute filter list: %w", a.context.LastError())
	}
	return nil
}

// FilterList returns a copy of the filter list for attribute.
func (a *Attribute) FilterList() (*FilterList, error) {
	var filterListPtr *C.tiledb_filter_list_t
	ret := C.tiledb_attribute_get_filter_list(a.context.tiledbContext.Get(), a.tiledbAttribute.Get(), &filterListPtr)
	runtime.KeepAlive(a)
	if ret != C.TILEDB_OK {
		return nil, fmt.Errorf("error getting tiledb attribute filter list: %w", a.context.LastError())
	}

	return newFilterListFromHandle(a.context, newFilterListHandle(filterListPtr)), nil
}

// SetCellValNum sets the number of attribute values per cell.
// This is inferred from the type parameter of the NewAttribute
// function, but can also be set manually.
func (a *Attribute) SetCellValNum(val uint32) error {
	ret := C.tiledb_attribute_set_cell_val_num(a.context.tiledbContext.Get(),
		a.tiledbAttribute.Get(), C.uint32_t(val))
	runtime.KeepAlive(a)
	if ret != C.TILEDB_OK {
		return fmt.Errorf("error setting tiledb attribute cell val num: %w", a.context.LastError())
	}
	return nil
}

// CellValNum returns number of values of one cell on this attribute.
// For variable-sized attributes returns TILEDB_VAR_NUM.
func (a *Attribute) CellValNum() (uint32, error) {
	var cellValNum C.uint32_t
	ret := C.tiledb_attribute_get_cell_val_num(a.context.tiledbContext.Get(), a.tiledbAttribute.Get(), &cellValNum)
	runtime.KeepAlive(a)
	if ret != C.TILEDB_OK {
		return 0, fmt.Errorf("error getting tiledb attribute cell val num: %w", a.context.LastError())
	}

	return uint32(cellValNum), nil
}

// CellSize gets the attribute cell size.
func (a *Attribute) CellSize() (uint64, error) {
	var cellSize C.uint64_t
	ret := C.tiledb_attribute_get_cell_size(a.context.tiledbContext.Get(), a.tiledbAttribute.Get(), &cellSize)
	runtime.KeepAlive(a)
	if ret != C.TILEDB_OK {
		return 0, fmt.Errorf("error getting tiledb attribute cell size: %w", a.context.LastError())
	}

	return uint64(cellSize), nil
}

// SetFillValue sets the default fill value for the input attribute. This value will
// be used for the input attribute whenever querying (1) an empty cell in
// a dense array, or (2) a non-empty cell (in either dense or sparse array)
// when values on the input attribute are missing (e.g., if the user writes
// a subset of the attributes in a write operation).
// Applicable to var-sized attributes.
// @note A call to `tiledb_attribute_cell_val_num` sets the fill value
//
//	of the attribute to its default. Therefore, make sure you invoke
//	`tiledb_attribute_set_fill_value` after deciding on the number
//	of values this attribute will hold in each cell.
//
// @note For fixed-sized attributes, the input `size` should be equal
//
//	to the cell size.
func (a *Attribute) SetFillValue(value interface{}) error {
	switch value := value.(type) {
	case int:
		return attributeSetFillValue(a, value)
	case int8:
		return attributeSetFillValue(a, value)
	case int16:
		return attributeSetFillValue(a, value)
	case int32:
		return attributeSetFillValue(a, value)
	case int64:
		return attributeSetFillValue(a, value)
	case uint:
		return attributeSetFillValue(a, value)
	case uint8:
		return attributeSetFillValue(a, value)
	case uint16:
		return attributeSetFillValue(a, value)
	case uint32:
		return attributeSetFillValue(a, value)
	case uint64:
		return attributeSetFillValue(a, value)
	case float32:
		return attributeSetFillValue(a, value)
	case float64:
		return attributeSetFillValue(a, value)
	case bool:
		return attributeSetFillValue(a, value)
	case string:
		cValue := unsafe.Pointer(C.CString(value))
		defer C.free(cValue)
		return attributeSetFillValueInternal(a, cValue, uint64(len(value)))
	}
	return fmt.Errorf("unrecognized fill value type %T", value)
}

func attributeSetFillValue[T scalarType](a *Attribute, value T) error {
	valNum, err := a.CellValNum()
	if err != nil {
		return err
	}
	dataType, err := a.Type()
	if err != nil {
		return err
	}
	valueSize := uint64(unsafe.Sizeof(value))
	if valNum != TILEDB_VAR_NUM {
		valueSize = dataType.Size() * uint64(valNum)
	}
	return attributeSetFillValueInternal(a, unsafe.Pointer(&value), valueSize)
}

func attributeSetFillValueInternal(a *Attribute, value unsafe.Pointer, valueSize uint64) error {
	ret := C.tiledb_attribute_set_fill_value(
		a.context.tiledbContext.Get(),
		a.tiledbAttribute.Get(),
		value,
		C.uint64_t(valueSize),
	)
	runtime.KeepAlive(a)
	if ret != C.TILEDB_OK {
		return fmt.Errorf("could not set attribute fill value: %w", a.context.LastError())
	}
	return nil
}

// SetFillValueNullable sets the default fill value for the input attribute. This value will
// be used for the input attribute whenever querying (1) an empty cell in
// a dense array, or (2) a non-empty cell (in either dense or sparse array)
// when values on the input attribute are missing (e.g., if the user writes
// a subset of the attributes in a write operation).
// Applicable to var-sized attributes.
// @note A call to `tiledb_attribute_cell_val_num` sets the fill value
//
//	of the attribute to its default. Therefore, make sure you invoke
//	`tiledb_attribute_set_fill_value` after deciding on the number
//	of values this attribute will hold in each cell.
//
// @note For fixed-sized attributes, the input `size` should be equal
//
//	to the cell size.
func (a *Attribute) SetFillValueNullable(value interface{}, valid bool) error {
	switch value := value.(type) {
	case int:
		return attributeSetFillValueNullable(a, value, valid)
	case int8:
		return attributeSetFillValueNullable(a, value, valid)
	case int16:
		return attributeSetFillValueNullable(a, value, valid)
	case int32:
		return attributeSetFillValueNullable(a, value, valid)
	case int64:
		return attributeSetFillValueNullable(a, value, valid)
	case uint:
		return attributeSetFillValueNullable(a, value, valid)
	case uint8:
		return attributeSetFillValueNullable(a, value, valid)
	case uint16:
		return attributeSetFillValueNullable(a, value, valid)
	case uint32:
		return attributeSetFillValueNullable(a, value, valid)
	case uint64:
		return attributeSetFillValueNullable(a, value, valid)
	case float32:
		return attributeSetFillValueNullable(a, value, valid)
	case float64:
		return attributeSetFillValueNullable(a, value, valid)
	case bool:
		return attributeSetFillValueNullable(a, value, valid)
	case string:
		cValue := unsafe.Pointer(C.CString(value))
		defer C.free(cValue)
		return attributeSetFillValueNullableInternal(a, cValue, uint64(len(value)), valid)
	}
	return fmt.Errorf("unrecognized fill value type %T", value)
}

func attributeSetFillValueNullable[T scalarType](a *Attribute, value T, valid bool) error {
	valNum, err := a.CellValNum()
	if err != nil {
		return err
	}
	dataType, err := a.Type()
	if err != nil {
		return err
	}
	valueSize := uint64(unsafe.Sizeof(value))
	if valNum != TILEDB_VAR_NUM {
		valueSize = dataType.Size() * uint64(valNum)
	}
	return attributeSetFillValueNullableInternal(a, unsafe.Pointer(&value), valueSize, valid)
}

func attributeSetFillValueNullableInternal(a *Attribute, value unsafe.Pointer, valueSize uint64, valid bool) error {
	cValid := C.uint8_t(0)
	if valid {
		cValid = 1
	}
	ret := C.tiledb_attribute_set_fill_value_nullable(
		a.context.tiledbContext.Get(),
		a.tiledbAttribute.Get(),
		value,
		C.uint64_t(valueSize),
		cValid,
	)
	runtime.KeepAlive(a)
	if ret != C.TILEDB_OK {
		return fmt.Errorf("could not set attribute fill value: %w", a.context.LastError())
	}
	return nil
}

// GetFillValue gets the default fill value for the input attribute. This value will
// be used for the input attribute whenever querying (1) an empty cell in
// a dense array, or (2) a non-empty cell (in either dense or sparse array)
// when values on the input attribute are missing (e.g., if the user writes
// a subset of the attributes in a write operation).
// Applicable to both fixed-sized and var-sized attributes.
func (a *Attribute) GetFillValue() (interface{}, uint64, error) {
	var fillValueSize C.uint64_t
	var cvalue unsafe.Pointer // a must be kept alive while cvalue is being accessed.

	ret := C.tiledb_attribute_get_fill_value(a.context.tiledbContext.Get(), a.tiledbAttribute.Get(), &cvalue, &fillValueSize)
	if ret != C.TILEDB_OK {
		return nil, 0, fmt.Errorf("error getting tiledb attribute fill value: %w", a.context.LastError())
	}

	attrDataType, err := a.Type()
	if err != nil {
		return nil, 0, fmt.Errorf("error getting tiledb attribute fill value: %w", a.context.LastError())
	}

	value, err := attrDataType.GetValue(1, cvalue)
	runtime.KeepAlive(a)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting tiledb attribute fill value: %w", a.context.LastError())
	}

	return value, uint64(fillValueSize), nil
}

// GetFillValueNullable gets the default fill value for the input attribute. This value will
// be used for the input attribute whenever querying (1) an empty cell in
// a dense array, or (2) a non-empty cell (in either dense or sparse array)
// when values on the input attribute are missing (e.g., if the user writes
// a subset of the attributes in a write operation).
// Applicable to both fixed-sized and var-sized attributes.
func (a *Attribute) GetFillValueNullable() (interface{}, uint64, bool, error) {
	var fillValueSize C.uint64_t
	var cvalue unsafe.Pointer // a must be kept alive while cvalue is being accessed.
	var cvalid C.uint8_t

	ret := C.tiledb_attribute_get_fill_value_nullable(a.context.tiledbContext.Get(), a.tiledbAttribute.Get(), &cvalue, &fillValueSize, &cvalid)
	if ret != C.TILEDB_OK {
		return nil, 0, false, fmt.Errorf("error getting tiledb attribute fill value: %w", a.context.LastError())
	}

	attrDataType, err := a.Type()
	if err != nil {
		return nil, 0, false, fmt.Errorf("error getting tiledb attribute fill value: %w", a.context.LastError())
	}

	value, err := attrDataType.GetValue(1, cvalue)
	runtime.KeepAlive(a)
	if err != nil {
		return nil, 0, false, fmt.Errorf("error getting tiledb attribute fill value: %w", a.context.LastError())
	}

	return value, uint64(fillValueSize), cvalid == 1, nil
}

// Name returns the name of the attribute.
func (a *Attribute) Name() (string, error) {
	var cName *C.char // a must be kept alive while cName is being accessed.
	ret := C.tiledb_attribute_get_name(a.context.tiledbContext.Get(), a.tiledbAttribute.Get(), &cName)
	if ret != C.TILEDB_OK {
		return "", fmt.Errorf("error getting tiledb attribute name: %w", a.context.LastError())
	}

	name := C.GoString(cName)
	runtime.KeepAlive(a)

	return name, nil
}

// Type returns the attribute datatype.
func (a *Attribute) Type() (Datatype, error) {
	var attrType C.tiledb_datatype_t
	ret := C.tiledb_attribute_get_type(a.context.tiledbContext.Get(), a.tiledbAttribute.Get(), &attrType)
	runtime.KeepAlive(a)
	if ret != C.TILEDB_OK {
		return 0, fmt.Errorf("error getting tiledb attribute type: %w", a.context.LastError())
	}
	return Datatype(attrType), nil
}

// DumpToString returns the attribute in ASCII format as a string.
func (a *Attribute) DumpToString() (string, error) {
	var cStr *C.tiledb_string_t
	ret := C.tiledb_attribute_dump_str(a.context.tiledbContext.Get(), a.tiledbAttribute.Get(), &cStr)
	runtime.KeepAlive(a)
	if ret != C.TILEDB_OK {
		return "", fmt.Errorf("error dumping attribute to string: %w", a.context.LastError())
	}
	defer C.tiledb_string_free(&cStr)

	goStr, err := stringHandleToString(cStr)
	if err != nil {
		return "", fmt.Errorf("error converting attribute dump to string: %w", err)
	}
	return goStr, nil
}

// DumpSTDOUT dumps the attribute in ASCII format to stdout.
func (a *Attribute) DumpSTDOUT() error {
	goStr, err := a.DumpToString()
	if err != nil {
		return err
	}
	fmt.Print(goStr)
	return nil
}

// Dump dumps the attribute in ASCII format to the given path.
func (a *Attribute) Dump(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("error path already %s exists", path)
	}
	goStr, err := a.DumpToString()
	if err != nil {
		return err
	}
	err = os.WriteFile(path, []byte(goStr), 0644)
	if err != nil {
		return fmt.Errorf("error writing attribute dump to file %s: %w", path, err)
	}
	return nil
}

// SetNullable sets if the attribute is nullable or not.
func (a *Attribute) SetNullable(nullable bool) error {
	var cNullable C.uint8_t
	if nullable {
		cNullable = 1
	}
	ret := C.tiledb_attribute_set_nullable(a.context.tiledbContext.Get(),
		a.tiledbAttribute.Get(), cNullable)
	runtime.KeepAlive(a)
	if ret != C.TILEDB_OK {
		return fmt.Errorf("error setting tiledb attribute nullable: %w", a.context.LastError())
	}
	return nil
}

// Nullable returns if the attribute is nullable or not.
func (a *Attribute) Nullable() (bool, error) {
	var nullable C.uint8_t
	ret := C.tiledb_attribute_get_nullable(a.context.tiledbContext.Get(), a.tiledbAttribute.Get(), &nullable)
	runtime.KeepAlive(a)
	if ret != C.TILEDB_OK {
		return false, fmt.Errorf("error getting tiledb attribute nullable: %w", a.context.LastError())
	}

	return nullable == 1, nil
}
