/*
Package htempl combines go html/templates with YAML configuration in one
file. The constants yamlStart and yamlEnd delimit the YAML header in front
of the go template section:

	---
	hello: "world"
	...
	{{.hello}}

# Predefined YAML values

The predefined YAML values are all expected to be filenames that are to be
included in the final htmpl document.

	include
		The named file is included in the YAML section to define more YAML
		data elements.
	includes
		The list of named files is included in the YAML section to define
		more YAML data elements.
	template
		The named file is included in the go template section for more
		template data.
	templates
		The list of named files is included in the go template section for
		more template data.

# Functions

The variable DefaultTemplateFunc is the default FuncMap that is installed and
supports the following template functions:

	map
		The arguments are expected to be pairs of names and interfaces
		and are returned as a new map.
	withDefault
		The first argument is a map, and the second a name and the third
		an interfae. The returned map makes sure the name and interface
		are present in the output map in case the input map does not have
		a value for name.
	md2html
		Returns the result of converting the argument from Markdown to
		HTML.
	safeattr
		Returns the argument as an template.HTMLAttr to avoid escaping in
		HTML attribute argumentsq.
	safehtml
		Returns the argument as template.HTML to avoid HTML escaping.
	safejs
		Returns the argument as template.JS to avoid javascript escaping.
	safecss
		Returns the argument as template.CSS to avoid CSS escaping.
	safeurl
		Returns the argument as template.URL to avoid URL escaping.
*/
package htempl
