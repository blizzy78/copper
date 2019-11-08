Template Language
=================

Copper uses a language similar to Go which should be fairly easy to use. The following
constructs are available:

Code blocks
-----------

Any code wrapped in `<% %>` code tags is considered Copper code that should be executed.
Text outside of these tags is rendered as-is (unescaped.) Code tags can appear anywhere in
the template, even inside HTML attributes and so on. This makes it very easy to render output
dynamically.

**Example:**

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
