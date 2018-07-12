package way

type pathPart struct {
	children   map[string]*pathPart
	match      []uint
	parameters []string
}

type routePart struct {
	children map[uint]*routePart
	path     string
}

// Router holds the list of routings and mappings
type Router struct {
	pathRoot  *pathPart
	routeRoot *routePart
}

// Add adds a route to the router
func (r Router) Add(path string, route ...uint) {

}

// RmPath removes a path from the router
func (r Router) RmPath(path string) {

}

// RmRoute removes a route from the router
func (r Router) RmRoute(route ...uint) {

}

// GetPath gets the path from the given route and parameters
func (r Router) GetPath(params map[string]string, route ...uint) (path string, err error) {
	return "", nil
}

// GetRoute gets the route and params for the given path
func (r Router) GetRoute(path string) (params map[string]string, route []uint, err error) {
	return nil, nil, nil
}
