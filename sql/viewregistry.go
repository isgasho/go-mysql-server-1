package sql

import (
	"strings"
	"sync"

	"gopkg.in/src-d/go-errors.v1"
)

var (
	ErrExistingView    = errors.NewKind("the view %s.%s already exists in the registry")
	ErrNonExistingView = errors.NewKind("the view %s.%s does not exist in the registry")
)

// A View is defined by a Node and has a name.
type View struct {
	name       string
	definition Node
}

// Creates a View with the specified name and definition.
func NewView(name string, definition Node) View {
	return View{name, definition}
}

// Returns the name of the view.
func (view *View) Name() string {
	return view.name
}

// Returns the definition of the view.
func (view *View) Definition() Node {
	return view.definition
}

// Views are scoped by the databases in which they were defined, so a key in
// the view registry is a pair of names: database and view.
type viewKey struct {
	dbName, viewName string
}

// Creates a viewKey ensuring both names are lowercase.
func newViewKey(databaseName, viewName string) viewKey {
	return viewKey{strings.ToLower(databaseName), strings.ToLower(viewName)}
}

// ViewRegistry is a map of viewKey to View whose access is protected by a
// RWMutex.
type ViewRegistry struct {
	mutex sync.RWMutex
	views map[viewKey]View
}

// Creates an empty ViewRegistry.
func NewViewRegistry() *ViewRegistry {
	return &ViewRegistry{
		views: make(map[viewKey]View),
	}
}

// Adds the view specified by the pair {database, view.Name()}, returning
// an error if there is already an element with that key.
func (registry *ViewRegistry) Register(database string, view View) error {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	key := newViewKey(database, view.Name())

	if _, ok := registry.views[key]; ok {
		return ErrExistingView.New(database, view.Name())
	}

	registry.views[key] = view
	return nil
}

// Deletes the view specified by the pair {databaseName, viewName}, returning
// an error if it does not exist.
func (registry *ViewRegistry) Delete(databaseName, viewName string) error {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	key := newViewKey(databaseName, viewName)

	if _, ok := registry.views[key]; !ok {
		return ErrNonExistingView.New(databaseName, viewName)
	}

	delete(registry.views, key)
	return nil
}

// Returns a pointer to the view specified by the pair {databaseName,
// viewName}, returning an error if it does not exist.
func (registry *ViewRegistry) View(databaseName, viewName string) (*View, error) {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	key := newViewKey(databaseName, viewName)

	if view, ok := registry.views[key]; ok {
		return &view, nil
	}

	return nil, ErrNonExistingView.New(databaseName, viewName)
}

// Returns the map of all views in the registry.
func (registry *ViewRegistry) AllViews() map[viewKey]View {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	return registry.views
}

// Returns an array of all the views registered under the specified database.
func (registry *ViewRegistry) ViewsInDatabase(databaseName string) (views []View) {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	for key, value := range registry.views {
		if key.dbName == databaseName {
			views = append(views, value)
		}
	}

	return views
}
