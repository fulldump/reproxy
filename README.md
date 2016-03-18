# Reproxy

<!-- MarkdownTOC autolink=true bracket=round depth=4 -->

- [Getting started](#getting-started)
- [Reproxy developer](#reproxy-developer)
    - [Build](#build)
    - [Static files](#static-files)

<!-- /MarkdownTOC -->

## Getting started

Run it:
```sh
./reproxy 
```

Open [http://localhost:8000/reproxy/] and configure it.


Change the default port:
```sh
./reproxy --address=9000
```

Change the configuration file:
```sh
./reproxy --filename=other_file.json
```


## Reproxy developer

### Build

How to build:
```sh
make
```

How to build for all architectures:
```sh
make all
```

### Static files

How to include static files into the binary:
```sh
go run src/genstatic/genstatic.go --dir=static/ --package=files > src/reproxy/files/data.go
```