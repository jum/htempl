# htempl

Each .htempl file contains a yaml fragment with variable definitions. A few variables have a special meaning:

* include: include yaml from the named file.
* includes: include the yaml from all of the named files.
* template: Include templates from the named file.
* templates: Include the template from all of the named files.

All variables are available as "." variables while executing the template.

The remainder of the file after the yaml block is appended to the other templates to build the final template. See the package documentation for more details on how to use the library:

https://pkg.go.dev/github.com/jum/htempl

The utility htempl in cmd/htempl can be used to statically generate HTML
pages from templates as a kind of static site generator.

Example:

```yaml
---
title: This is a Test
template: site.templ
...
{{.title}}
template text appended to site.templ and executed with the yaml data as dot.
{{template "site" .}}
```

There are a few special functions in the FuncMap of the template that can be used:

* safeattr(string) convert string into a HTML attribute for dynamically constructing HTML attributes.
* safehtml(sring) convert string to HTML that is not escaped by the default html/template escaping.
* safejs(string) convert string to javascript that is not escaped.
* safecss(string) convert string to css that is not escaped.
* safeurl(string) convert string to an url that is not escaped.
* md2html(string) convert string in markdown syntax into HTML using the gomarkdown markdown parser.

A recent talk on htempl for the Hannover golang meeting in the subdirctory slides. To view the slides online: https://go-talks.appspot.com/github.com/jum/htempl/slides/htempl.slide
