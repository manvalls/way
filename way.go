package way

import (
	"errors"
	"net/url"
	"strings"
)

// Params hold information about provided or matched URL parameters
type Params = map[string][]string

// Route contains the matched route
type Route = []string

// RouteMap holds a list of route mappings
type RouteMap = map[string]Route

type pathPart struct {
	children   map[string]*pathPart
	suffix     *pathPart
	match      []string
	parameters []string
}

type routePart struct {
	children map[string]*routePart
	path     []*matchedPathBit
}

type matchedPathBit struct {
	part     string
	isParam  bool
	isSuffix bool
}

var reservedChars = ", ?"

// Router holds the list of routings and mappings
type Router struct {
	pathRoot  *pathPart
	routeRoot *routePart
}

// NewRouter builds a new router instance
func NewRouter() Router {
	return Router{&pathPart{children: map[string]*pathPart{}}, &routePart{children: map[string]*routePart{}}}
}

// BuildRouter builds a new router instance and adds the provided routes to it,
// panicking if there's an error adding the routes
func BuildRouter(maps ...RouteMap) Router {
	router := NewRouter()

	for _, m := range maps {
		if err := router.AddAll(m); err != nil {
			panic(err)
		}
	}

	return router
}

// Add adds a route to the router
func (r Router) Add(path string, route ...string) error {
	pathParent := r.pathRoot
	currentPart := ""
	currentParam := ""
	isParam := false
	matched := []*matchedPathBit{}
	params := []string{}
	hasSuffix := false

	flush := func() (err error) {
		next := ""
		isSuffix := false

		if isParam {
			if currentParam == "" {
				return
			}

			if hasSuffix {
				return ErrMiddleSuffix
			}

			lastPos := len(currentParam) - 1
			if currentParam[lastPos] == '*' {
				currentParam = currentParam[:lastPos]
				isSuffix = true
				hasSuffix = true
			}

			currentParam, err = url.QueryUnescape(currentParam)
			if err != nil {
				return
			}

			matched = append(matched, &matchedPathBit{currentParam, true, isSuffix})
			params = append(params, currentParam)
		} else {
			if currentPart == "" {
				return
			}

			if hasSuffix {
				return ErrMiddleSuffix
			}

			currentPart, err = url.QueryUnescape(currentPart)
			if err != nil {
				return
			}

			matched = append(matched, &matchedPathBit{currentPart, false, false})
			next = currentPart
		}

		if !isSuffix {
			nextParent := pathParent.children[next]
			if nextParent == nil {
				nextParent = &pathPart{children: map[string]*pathPart{}}
				pathParent.children[next] = nextParent
			}
			pathParent = nextParent
		} else {
			nextParent := pathParent.suffix
			if nextParent == nil {
				nextParent = &pathPart{children: map[string]*pathPart{}}
				pathParent.suffix = nextParent
			}
			pathParent = nextParent
		}

		currentParam = ""
		currentPart = ""
		isParam = false
		return
	}

	for _, c := range path {
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
		if strings.ContainsAny(routeBit, reservedChars) {
			return ErrInvalidChars
		}

		nextParent := routeParent.children[routeBit]
		if nextParent == nil {
			nextParent = &routePart{children: map[string]*routePart{}}
			routeParent.children[routeBit] = nextParent
		}

		routeParent = nextParent
	}

	routeParent.path = matched
	return nil
}

// AddAll adds all provided routes to the router
func (r Router) AddAll(m RouteMap) error {
	for key, value := range m {
		if err := r.Add(key, value...); err != nil {
			return err
		}
	}

	return nil
}

// ErrNotFound is returned when the requested route could not be found
var ErrNotFound = errors.New("Requested route not found")

// ErrMissingParam is returned when there is a missing parameter
var ErrMissingParam = errors.New("Missing parameter")

// ErrMiddleSuffix is returned when the provided path contains a suffix not located at the end
var ErrMiddleSuffix = errors.New("Suffix parameters can only happen at the end of the path")

// ErrInvalidChars is returned when invalid characters are used in the route
var ErrInvalidChars = errors.New("Invalid characters")

// Merge builds new parameters after merging provided ones
func Merge(params ...Params) Params {
	result := make(Params)

	for _, p := range params {
		for paramName, paramValues := range p {
			for _, value := range paramValues {
				result[paramName] = append(result[paramName], value)
			}
		}
	}

	return result
}

// GetURL gets the URL from the given route and parameters
func (r Router) GetURL(originalParams Params, route ...string) (string, error) {
	parent := r.routeRoot
	for _, i := range route {
		if strings.ContainsAny(i, reservedChars) {
			return "", ErrInvalidChars
		}

		parent = parent.children[i]
		if parent == nil {
			return "", ErrNotFound
		}
	}

	if parent.path == nil {
		return "", ErrNotFound
	}

	params := Merge(originalParams)
	path := ""

	for _, bit := range parent.path {
		if bit.isParam {
			param := params[bit.part]
			if len(param) == 0 {
				return "", ErrMissingParam
			}

			p := param[0]
			params[bit.part] = param[1:]

			if !bit.isSuffix {
				p = url.QueryEscape(p)
			}

			path += "/" + p
		} else {
			path += "/" + bit.part
		}
	}

	if path == "" {
		path = "/"
	}

	var values url.Values
	values = params
	query := values.Encode()

	if query != "" {
		path += "?" + query
	}

	return path, nil
}

func match(parts []string, params []string, parent *pathPart) ([]string, *pathPart, error) {
	if len(parts) == 0 {
		if parent.match == nil {
			return nil, nil, ErrNotFound
		}

		return params, parent, nil
	}

	keys := []string{parts[0], ""}
	nextParts := parts[1:]

	for _, key := range keys {
		child := parent.children[key]
		if child != nil {
			nextParams := params
			if key == "" {
				nextParams = append(params, parts[0])
			}

			p, m, err := match(nextParts, nextParams, child)
			if err == nil {
				return p, m, nil
			}
		}
	}

	if parent.suffix != nil {
		nextParams := append(params, strings.Join(parts, "/"))
		return match([]string{}, nextParams, parent.suffix)
	}

	return nil, nil, ErrNotFound
}

// GetRoute gets the route and params for the given URL
func (r Router) GetRoute(urlToMatch *url.URL) (Params, Route, error) {
	currentPart := ""
	parts := []string{}
	path := urlToMatch.Path

	flush := func() (err error) {
		if currentPart == "" {
			return
		}

		currentPart, err = url.QueryUnescape(currentPart)
		if err != nil {
			return err
		}

		parts = append(parts, currentPart)
		currentPart = ""
		return
	}

	for _, c := range path {
		switch c {
		case '/':
			err := flush()
			if err != nil {
				return nil, nil, err
			}

		default:
			currentPart += string(c)
		}
	}

	err := flush()
	if err != nil {
		return nil, nil, err
	}

	paramList, matchedPart, err := match(parts, []string{}, r.pathRoot)
	if err != nil {
		return nil, nil, err
	}

	params := make(Params)
	for i, p := range matchedPart.parameters {
		params[p] = append(params[p], paramList[i])
	}

	return Merge(params, urlToMatch.Query()), Clone(matchedPart.match), nil
}

// MustGetURL gets the URL from the given route and parameters and panics if there's an error
func (r Router) MustGetURL(originalParams Params, route ...string) string {
	url, err := r.GetURL(originalParams, route...)
	if err != nil {
		panic(err)
	}

	return url
}

// MustGetRoute gets the URL from the given route and parameters and panics if there's an error
func (r Router) MustGetRoute(urlToMatch *url.URL) (Params, Route) {
	params, route, err := r.GetRoute(urlToMatch)
	if err != nil {
		panic(err)
	}

	return params, route
}

// Clone returns a copy of the provided route
func Clone(route Route) Route {
	result := make(Route, len(route))
	copy(result, route)
	return result
}
