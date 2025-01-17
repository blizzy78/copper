Template Language
=================

Copper uses a language similar to Go which should be fairly easy to use. The following
constructs are available:

Primitives, Common Expressions
------------------------------

The following should be fairly self-explanatory. All your usual programming primitives and operators.

```
// integers
let x = 123

// strings, string concatenation
let s = "foo" + "bar"

// math(s)
let x = (2 + 2) / 4 * 3 % 5 - 7

// boolean expressions
let x = y >= 5
let x = boolA || boolB && boolC

// accessing fields, methods, slice elements
let x = y.foo.bar().baz[qux]
```

Line Breaks
-----------

Copper ignores line breaks in code completely. That is to say, use line breaks to make your
code more readable. Copper itself doesn't care.

```
// you really should use line breaks like in this example
if x > 5
  foo()
else
  bar()
end

// this is equivalent, but not as readable
if x > 5 foo() else bar() end
```

Code Blocks
-----------

Any code wrapped in `<% %>` code tags is considered Copper code that should be executed.
Text outside of these tags is rendered as-is (unescaped.) Code tags can appear anywhere in
the template, even inside HTML attributes and so on. This makes it very easy to render output
dynamically.

The values of all expression statements inside the code block will be rendered as a single
string, concatenated together (without any delimiter.) `nil` values will be ignored.
If a value is itself a slice or array, its elements will again be rendered
as strings, concatenated together. `nil` elements will be ignored.

### Example ###

```
This is literal text that will be rendered as-is.

The following is a code block:

<%
html(user.name)
safe(article.introText)
%>

Dynamic HTML: <a href="<% html(article.url) %>"><% html(article.headline) %></a>

Again, this is literal text.
```

Comments
--------

Line comments start with `//` and go to the end of the line, or to the end of the code block
using `%>`

Block comments start with `/*` and end with `*/`, and they may span multiple code blocks.

### Example ###

```
<%
let x = 123  // this comment is an example
safe(x)

/* a multi-line
code block here */
%>

<% // comment one %> This text will be rendered. <% // comment two %>

<% /* %> This text will not be rendered. <% */ %>
```

Statements vs Expressions
-------------------------

Most statements can be used as expressions and vice versa. For example, the expression
`1 + 2` can be used as a standalone statement, which is usually used by the containing
statement to return it:

```
// if foo is greater than 0, x would be 3, otherwise it would be 7
let x = if foo > 0
  1 + 2
else
  3 + 4
end
```

Usually this would be written like this:

```
if foo > 0
  let x = 1 + 2
else
  let x = 3 + 4
end
```

Variable Assignment - `let`
---------------------------

**`let IDENT = EXPR`**

The `let` statement can be used to set a variable in the current scope.

### Expressions ###

`let` statements cannot be used as expressions. The following is invalid code:

```
let x = let y = 123   // invalid
```

### Example ###

```
let x = 1 + 2 * 3   // set x=7

// this works because the if statement is also an expression
// because x==7, y will be "foo"
let y = if x > 5 "foo" else "bar" end
```

Loop - `for`
------------

**`for IDENT in RANGE_EXPR ... end`**

**`for IDENT, STATUS_IDENT in RANGE_EXPR ... end`**

The `for` statement iterates over a set of values, produced by a [Ranger]. The `RANGE_EXPR`
is an expression that must be a `Ranger`. `IDENT` is the variable identifier used in the
`for` loop body for the current value the `Ranger` has produced. `STATUS_IDENT` is an
optional identifier for a variable that provides status of the current loop iteration
(see [Status].)

There is no builtin way to iterate over the elements of a slice, for example. Instead,
a helper function must be used to create a `Ranger` that produces the slice's elements
(see example.)

The `for` loop's body is ended with the `end` statement.

The `break` statement can be used to break out of the loop. The `continue` statement
can be used to stop the current iteration of the loop and start the next (if any.)

### Expressions ###

`for` statements can be used as expressions that return all values of expression
statements inside the loop's body, as a slice. That is, the loop acts as if it was
inside a `capture ... end` block.

If the returned slice would be empty, `nil` is returned instead of the slice. If the
slice would only contain a single element, the value of that element is returned instead
of the slice.

### Example ###

```
// use helper function range() to produce a ranger over stringSlice
for s in range(stringSlice)
  safe(s)
end

let sum = 0
// use helper function fromTo() to produce a ranger over integers
for i in fromTo(1, 10)
  let sum = sum + i
end

let sum = 0
// use helper function range() to produce a ranger over a hash
for e in range(hash)
  let sum = sum + e.Value
end
```

Conditionals - `if`, `elseif`, `else`
-------------------------------------

**`if BOOL_EXPR ... elseif BOOL_EXPR ... else ... end`**

`if`/`elseif`/`else` execute the enclosed statements only if `BOOL_EXPR` evaluates to
`true`. Zero or more `elseif` statements may be used, but only zero or one `else`
statement, which must be the last. The last `if`/`elseif`/`else` block (whichever comes
last) must be closed with the `end` statement.

### Expressions ###

`if` statements can be used as expressions that return all values of expression statements
inside the block that has been executed, as a slice. That is, the `if` statement acts as
if it was inside a `capture ... end` block.

If the returned slice would be empty, `nil` is returned instead of the slice. If the
slice would only contain a single element, the value of that element is returned instead
of the slice.

### Example ###

```
if x > 5
  let y = "foo"
else
  let y = "bar"
end

// or rewrite using if as an expression
let y = if x > 5
  "foo"
else
  "bar"
end

// or rewrite as a single line expression
// (just removed the line breaks)
let y = if x > 5 "foo" else "bar" end

// call a function using conditional argument
theFunc(if x > 5
  "foo"
else
  "bar"
end)
```

Capture All Expressions as Slice - `capture`
--------------------------------------------

**`capture ... end`**

Blocks only return the last expression statement's value as the value of the block.
The `capture` statement can be used to capture all expression statements' values in a
slice instead.

This is especially useful when using mixed literal output and code blocks, for example
to pass it all to another template.

If the returned slice would be empty, `nil` is returned instead of the slice. If the
slice would only contain a single element, the value of that element is returned instead
of the slice.

### Example ###

```
<%
// get article from somewhere
let article = ...

// call a function, passing it a hash containing the entries "headline" and "body"
foo({
  // pass the article's headline as the "headline" entry
  "headline": article.headline,

  // pass the "body" entry containing a mix of literal output and code blocks -
  // because these are multiple expressions, "capture" must be used to combine them
  // into a single slice
  "body": capture %>
    <p>
      <% html(article.text) %>
    </p>
  <% end
})
%>
```

Hash - `{ }`
------------

**`{ KEY_1_EXPR: EXPR_1, ... }`**

A hash expression is used to create a map of values. Key expressions must be of type `string`.
The internal type of the hash is `map[string]interface{}`

### Example ###

```
let h = {
  "foo": "bar",
  "value": 42   // note no extra comma after the last value!
}

let key = "foo"
let h = {
  // use expression as key instead of literal string
  key: "bar"
}
```




[Ranger]: https://godoc.org/github.com/blizzy78/copper/ranger#Ranger
[Status]: https://godoc.org/github.com/blizzy78/copper/ranger#Status
