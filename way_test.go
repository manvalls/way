package way

import (
	"net/url"
	"testing"
)

func getRouter() Router {
	r := BuildRouter(RouteMap{
		"/:foo/bar/:suffix*": Route{1},
		"/:foo/:foo/foo":     Route{1, 2, 3},
	})

	r.Add("/", 0)
	r.Add("/foo/:bar", 0, 1)
	r.Add("/:foo/bar", 0, 2)
	return r
}

func assertPath(actual string, expected string, t *testing.T) {
	if actual != expected {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

func assertParams(actual Params, expected Params, t *testing.T) {
	if len(actual) != len(expected) {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}

	for key := range actual {
		if len(actual[key]) != len(expected[key]) {
			t.Fatalf("Expected %v, got %v", expected, actual)
		}

		for subkey := range actual[key] {
			if actual[key][subkey] != expected[key][subkey] {
				t.Fatalf("Expected %v, got %v", expected, actual)
			}
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

	u, _ := url.Parse("/")
	params, route, err := r.GetRoute(u)
	assertNoError(err, t)
	assertParams(params, Params{}, t)
	assertRoute(route, []uint{0}, t)

	u, _ = url.Parse("/foo/baz")
	params, route, err = r.GetRoute(u)
	assertNoError(err, t)
	assertParams(params, Params{"bar": []string{"baz"}}, t)
	assertRoute(route, []uint{0, 1}, t)

	u, _ = url.Parse("/foo/ba+r")
	params, route, err = r.GetRoute(u)
	assertNoError(err, t)
	assertParams(params, Params{"bar": []string{"ba r"}}, t)
	assertRoute(route, []uint{0, 1}, t)

	u, _ = url.Parse("/fooo/bar")
	params, route, err = r.GetRoute(u)
	assertNoError(err, t)
	assertParams(params, Params{"foo": []string{"fooo"}}, t)
	assertRoute(route, []uint{0, 2}, t)

	u, _ = url.Parse("/test/bar/an/interesting/suffix?foo=bar")
	params, route, err = r.GetRoute(u)
	assertNoError(err, t)
	assertParams(params, Params{"foo": []string{"test", "bar"}, "suffix": []string{"an/interesting/suffix"}}, t)
	assertRoute(route, []uint{1}, t)

	u, _ = url.Parse("/one/two/foo")
	params, route, err = r.GetRoute(u)
	assertNoError(err, t)
	assertParams(params, Params{"foo": []string{"one", "two"}}, t)
	assertRoute(route, []uint{1, 2, 3}, t)

	u, _ = url.Parse("/faasdasd")
	_, _, err = r.GetRoute(u)
	if err != ErrNotFound {
		t.Fatal("Expected error to be returned")
	}
}

func TestGetURL(t *testing.T) {
	r := getRouter()

	path, err := r.GetURL(nil, 0)
	assertNoError(err, t)
	assertPath(path, "/", t)

	path, err = r.GetURL(Params{"bar": []string{"ba z", "bar"}}, 0, 1)
	assertNoError(err, t)
	assertPath(path, "/foo/ba+z?bar=bar", t)

	path, err = r.GetURL(Params{"foo": []string{"fooo"}}, 0, 2)
	assertNoError(err, t)
	assertPath(path, "/fooo/bar", t)

	path, err = r.GetURL(Params{"foo": []string{"test"}, "suffix": []string{"an/interesting/suffix"}}, 1)
	assertNoError(err, t)
	assertPath(path, "/test/bar/an/interesting/suffix", t)

	_, err = r.GetURL(nil, 5)
	if err != ErrNotFound {
		t.Fatal("Expected error to be returned")
	}

	_, err = r.GetURL(nil, 0, 2)
	if err != ErrMissingParam {
		t.Fatal("Expected error to be returned")
	}
}
