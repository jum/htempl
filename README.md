# htempl

A simple command to generate HTML files using go templates.

```sh
htempl -d destdir *.htempl
```

Each .htempl file contains a yaml fragment with variable definitions. A few variables have a special meaning:

* include: include yaml from the named file.
* includes: include the yaml from all of the named files.
* template: Include templates from the named file.
* templates: Include the template from all of the named files.

All variables are available as "." variables while executing the template. Please note that include and includes only work in the top level yaml fragment, not in included ones.

The remainder of the file after the yaml block is appended to the other templates to build the final template and write a file with the same name but using the suffix .html. The destdir parameter can be used to place the generated html files in a different directory.

Example:

```yaml
---
title: This is a Test
template: site.templ
...
{{.title}}
template text appended to site.templ and executed with the yaml data as dot.
{{template "site"}}
```

There are a few special functions in the FuncMap of the template that can be used:

* htmlattr(string) convert string into a HTML attribute for dynamically constructing HTML attributes.
* md2html(string) convert string in markdown syntax into HTML using the blackfriday markdown parser.
