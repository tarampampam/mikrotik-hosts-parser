// Source: <https://github.com/wso2/product-apim-tooling/tree/master/import-export-cli/box>, <https://nuancesprog.ru/p/4894/amp/>
//go:generate go run resources_gen.go

package main

type resourcesBox struct {
	storage map[string][]byte
}

type ResourcesBox interface {
	Get(file string) ([]byte, bool)
	Add(file string, content []byte)
	Has(file string) bool
}

// Resource expose
var Resources = newResourceBox()

func newResourceBox() ResourcesBox {
	return &resourcesBox{storage: make(map[string][]byte)}
}

// Find for a file
func (r *resourcesBox) Has(file string) bool {
	if _, ok := r.storage[file]; ok {
		return true
	}
	return false
}

// Get file's content
// Always use / for looking up
// For example: /init/README.md is actually resources/init/README.md
func (r *resourcesBox) Get(file string) ([]byte, bool) {
	if f, ok := r.storage[file]; ok {
		return f, ok
	}
	return nil, false
}

// Add a file to box
func (r *resourcesBox) Add(file string, content []byte) {
	r.storage[file] = content
}
