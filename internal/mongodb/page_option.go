package mongodb

import "sync"

// FindPageOptionSyncPool is a sync.Pool that stores FindPageOption objects.
// sync.Pool is a type-safe, reusable object pool provided by Go's standard library.
var FindPageOptionSyncPool = sync.Pool{
	New: func() interface{} {
		// New is a function that creates a new FindPageOption when the pool is empty.
		return new(FindPageOption)
	},
}

// acquireFindPageOption is a function that gets a FindPageOption object from the pool.
func acquireFindPageOption() *FindPageOption {
	return FindPageOptionSyncPool.Get().(*FindPageOption)
}

// releaseFindPageOption is a function that puts a FindPageOption object back into the pool.
// Before putting it back, it resets the selector and fields of the FindPageOption to nil.
func releaseFindPageOption(m *FindPageOption) {
	m.selector = nil
	m.fields = nil
	FindPageOptionSyncPool.Put(m)
}

// FindPageOption is a struct that holds the fields and selector for a page.
type FindPageOption struct {
	fields   []string // fields is a slice of strings that represents the fields to sort by.
	selector any      // selector is an interface that represents the fields to select.
}

// SetSortField is a method on FindPageOption that sets the fields to sort by.
// It takes a variadic parameter of strings, allowing multiple fields to be passed.
// It returns a pointer to the FindPageOption for method chaining.
func (o *FindPageOption) SetSortField(field ...string) *FindPageOption {
	o.fields = field
	return o
}

// SetSelectField is a method on FindPageOption that sets the selector.
// It takes an interface as a parameter, allowing any type to be passed.
// It returns a pointer to the FindPageOption for method chaining.
func (o *FindPageOption) SetSelectField(bson interface{}) *FindPageOption {
	o.selector = bson
	return o
}

// NewFindPageOption is a function that creates a new FindPageOption.
// It does this by acquiring a FindPageOption from the pool.
func NewFindPageOption() *FindPageOption {
	return acquireFindPageOption()
}
