[![Build Status](https://travis-ci.org/blizzy78/copper.svg?branch=master)](https://travis-ci.org/blizzy78/copper) [![GoDoc](https://godoc.org/github.com/blizzy78/copper?status.svg)](https://godoc.org/github.com/blizzy78/copper) [![Coverage Status](https://coveralls.io/repos/github/blizzy78/copper/badge.svg?branch=master)](https://coveralls.io/github/blizzy78/copper?branch=master)


Copper Template Rendering Engine
================================

Copper is a template rendering engine written in Go. It is similar to Buffalo's [Plush],
but has more capabilities.

Copper is agnostic of any HTTP router or any other framework and can also be used
standalone (for example, to render e-mail messages or any other text.)

See [Usage Example](#usage-example) for a simple standalone example of using Copper.

Template Language
-----------------

Copper uses a language similar to Go which should be fairly easy to use.

See [SYNTAX.md] for a full list of language constructs.

No Auto-Escaping Text
---------------------

Copper never escapes text automatically. Every time text should be rendered, it must be
passed through one of the provided (or a custom) [helper function] that marks the text to be
safe for output, optionally escaping it. While this may sound tedious when coming from other
rendering engines, it prevents Copper from guessing (and mis-guessing) the correct escaping
mechanism to use. Instead, the responsibility is explicitly shifted to the author of the
template.

Usage Example
-------------

This is a simple, but full example of how to use Copper. It shows how to render a simple
template that outputs some dynamic variables.

See [Copper Example] for a more complete example that integrates into [net/http], [Gorilla],
[Chi], or [httprouter].

```go
package main

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/blizzy78/copper/helpers"
	"github.com/blizzy78/copper/template"
)

const (
	// the template to render
	tmpl = `<html>
<body>
<p>Hello, <% html(who) %>!</p>
<% safe(someHTML) %>
</body>
</html>`
)

func main() {
	// load a template by name -
	// in this example, we ignore the name and always return the same template
	loader := template.LoaderFunc(func(name string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(tmpl)), nil
	})

	// construct a new renderer
	r := template.NewRenderer(loader,
		// html will be a global template function that HTML-escapes strings and
		// marks them as safe for output
		template.WithScopeData("html", helpers.HTML),

		// safe will be a global template function that marks strings as safe
		// for output
		template.WithScopeData("safe", helpers.Safe))

	// a context that can be used by template helper functions -
	// the renderer does not use it
	ctx := context.Background()

	// let's render into this buffer - any io.Writer is fine
	buf := bytes.Buffer{}

	// the name of the template to render -
	// in this example, the name is ignored
	name := "myTemplate"

	// data provided to the template being rendered
	data := map[string]interface{}{
		"who":      "World",
		"someHTML": "<p>This is HTML</p>",
	}

	// parse and render the template
	err := r.Render(ctx, &buf, name, data)
	if err != nil {
		panic(err)
	}

	// output buffer contents
	println(buf.String())
}
```

Usage on Command Line
---------------------

You can use the [coppercli] tool to render text on the command line.

License
-------

Copper is licensed under the MIT license.



[Plush]: https://github.com/gobuffalo/plush
[Copper Example]: https://github.com/blizzy78/copperexample
[net/http]: https://golang.org/pkg/net/http/
[Gorilla]: https://github.com/gorilla/mux
[Chi]: https://github.com/go-chi/chi
[httprouter]: https://github.com/julienschmidt/httprouter
[SYNTAX.md]: SYNTAX.md
[helper function]: https://godoc.org/github.com/blizzy78/copper/helpers
[coppercli]: https://github.com/blizzy78/coppercli
