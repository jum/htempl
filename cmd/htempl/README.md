# htempl

A simple command to generate HTML files using go templates combined with
data from YAML files. Please note that all references to files are relative
to the working directory of the command.

```sh
htempl -dest destdir -destsuffix html *.htempl
```

You can also process a tree of files:

```sh
htempl -dest destdir -src srcdir -destsuffix html -srcsuffix htempl
```

Please see the README.md of the library for details of the file format.
