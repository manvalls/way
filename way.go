package way

import "net/url"

type pathPart struct {
	children   map[string]*pathPart
	match      []uint
	parameters []string
}

type routePart struct {
	children map[uint]*routePart
	path     []*matchedPathBit
}

type matchedPathBit struct {
	part    string
	isParam bool
}

// Router holds the list of routings and mappings
type Router struct {
	pathRoot  *pathPart
	routeRoot *routePart
}

// Add adds a route to the router
func (r Router) Add(path string, route ...uint) error {
	pathParent := r.pathRoot
	currentPart := ""
	currentParam := ""
	isParam := false
	matched := []*matchedPathBit{}
	params := []string{}

	flush := func() (err error) {
		next := ""

		if isParam {
			if currentParam == "" {
				return
			}

			currentParam, err = url.QueryUnescape(currentParam)
			if err != nil {
				return
			}

			matched = append(matched, &matchedPathBit{currentParam, true})
			params = append(params, currentParam)
		} else {
			if currentPart == "" {
				return
			}

			currentPart, err = url.QueryUnescape(currentPart)
			if err != nil {
				return
			}

			matched = append(matched, &matchedPathBit{currentPart, false})
			next = currentPart
		}

		nextParent := pathParent.children[next]
		if nextParent == nil {
			nextParent = &pathPart{children: map[string]*pathPart{}}
			pathParent.children[next] = nextParent
		}

		pathParent = nextParent
		return
	}

	for c := range path {
		switch c {
		case '/':
			err := flush()
			if err != nil {
				return err
			}

		case ':':
			if currentParam == "" && currentPart == "" {
				isParam = true
			} else {
				if isParam {
					currentParam += string(c)
				} else {
					currentPart += string(c)
				}
			}
		case '?':
			err := flush()
			if err != nil {
				return err
			}

			break
		default:
			if isParam {
				currentParam += string(c)
			} else {
				currentPart += string(c)
			}
		}
	}

	err := flush()
	if err != nil {
		return err
	}

	pathParent.match = route
	pathParent.parameters = params

	routeParent := r.routeRoot
	for _, routeBit := range route {
		nextParent := routeParent.children[routeBit]
		if routeParent == nil {
			nextParent = &routePart{children: map[uint]*routePart{}}
			routeParent.children[routeBit] = nextParent
		}

		routeParent = nextParent
	}

	routeParent.path = matched
	return nil
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
