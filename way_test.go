package way

import (
	"testing"
)

func getRouter() Router {
	r := NewRouter()
	r.Add("/", 0)
	r.Add("/foo/:bar", 0, 1)
	r.Add("/:foo/bar", 0, 2)
	r.Add("/:foo/bar/:suffix*", 1)
	return r
}

func assertPath(actual string, expected string, t *testing.T) {
	if actual != expected {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

func assertParams(actual map[string]string, expected map[string]string, t *testing.T) {
	if len(actual) != len(expected) {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}

	for key := range actual {
		if actual[key] != expected[key] {
			t.Fatalf("Expected %v, got %v", expected, actual)
		}
	}
}

func assertRoute(actual []uint, expected []uint, t *testing.T) {
	if len(actual) != len(expected) {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}

	for i := range actual {
		if actual[i] != expected[i] {
			t.Fatalf("Expected %v, got %v", expected, actual)
		}
	}
}

func assertNoError(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetRoute(t *testing.T) {
	r := getRouter()

	params, route, err := r.GetRoute("/")
	assertNoError(err, t)
	assertParams(params, map[string]string{}, t)
	assertRoute(route, []uint{0}, t)

	params, route, err = r.GetRoute("/foo/baz")
	assertNoError(err, t)
	assertParams(params, map[string]string{"bar": "baz"}, t)
	assertRoute(route, []uint{0, 1}, t)

	params, route, err = r.GetRoute("/foo/ba+r")
	assertNoError(err, t)
	assertParams(params, map[string]string{"bar": "ba r"}, t)
	assertRoute(route, []uint{0, 1}, t)

	params, route, err = r.GetRoute("/fooo/bar")
	assertNoError(err, t)
	assertParams(params, map[string]string{"foo": "fooo"}, t)
	assertRoute(route, []uint{0, 2}, t)

	params, route, err = r.GetRoute("/test/bar/an/interesting/suffix")
	assertNoError(err, t)
	assertParams(params, map[string]string{"foo": "test", "suffix": "an/interesting/suffix"}, t)
	assertRoute(route, []uint{1}, t)

	_, _, err = r.GetRoute("/faasdasd")
	if err != ErrNotFound {
		t.Fatal("Expected error to be returned")
	}
}

func TestGetPath(t *testing.T) {
	r := getRouter()

	path, err := r.GetPath(nil, 0)
	assertNoError(err, t)
	assertPath(path, "/", t)

	path, err = r.GetPath(map[string]string{"bar": "ba z"}, 0, 1)
	assertNoError(err, t)
	assertPath(path, "/foo/ba+z", t)

	path, err = r.GetPath(map[string]string{"foo": "fooo"}, 0, 2)
	assertNoError(err, t)
	assertPath(path, "/fooo/bar", t)

	path, err = r.GetPath(map[string]string{"foo": "test", "suffix": "an/interesting/suffix"}, 1)
	assertNoError(err, t)
	assertPath(path, "/test/bar/an/interesting/suffix", t)

	_, err = r.GetPath(nil, 5)
	if err != ErrNotFound {
		t.Fatal("Expected error to be returned")
	}

	_, err = r.GetPath(nil, 0, 2)
	if err != ErrMissingParam {
		t.Fatal("Expected error to be returned")
	}
}
