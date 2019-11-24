// +build gofuzz

package lexer

import (
	"bytes"
	"strings"

	"github.com/blizzy78/copper/evaluator"
	"github.com/blizzy78/copper/helpers"
	"github.com/blizzy78/copper/lexer"
	"github.com/blizzy78/copper/parser"
	"github.com/blizzy78/copper/ranger"
	"github.com/blizzy78/copper/scope"
	"github.com/blizzy78/copper/template"

	// required for go-fuzz-build
	_ "github.com/dvyukov/go-fuzz/go-fuzz-dep"
)

func Fuzz(data []byte) (ret int) {
	if len(data) <= 2 {
		return 0
	}

	r := bytes.NewReader(data)
	w := bytes.Buffer{}
	d := map[string]interface{}{}
	s := scope.Scope{}

	s.Set("range", ranger.NewInt)
	s.Set("fromTo", ranger.NewFromTo)

	s.Set("safe", helpers.Safe)
	s.Set("hasPrefix", helpers.HasPrefix)
	s.Set("hasSuffix", helpers.HasSuffix)
	s.Set("len", helpers.Len)
	s.Set("has", helpers.Has)
	s.Set("html", helpers.HTML)

	defer func() {
		if er := toError(recover()); er != nil {
			msg := er.Error()
			switch msg {
			case "maxExclusive must be greater than minInclusive":
				fallthrough
			case "cannot get length of nil":
				// okay
			default:
				if strings.HasPrefix(msg, "cannot get length of unsupported type:") {
					// okay
					return
				}

				panic(er)
			}
		}
	}()

	if err := template.Render(r, &w, d, &s); err != nil {
		if !lexer.IsParseError(err) && !parser.IsParseError(err) && !evaluator.IsEvaluationError(err) {
			panic(err)
		}

		return 0
	}

	return 1
}

func toError(v interface{}) error {
	if err, ok := v.(error); ok {
		return err
	}

	return nil
}
