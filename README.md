# cfgutil

Utility for generating  cfg files from openapi specifications. 

JSON mode (`-json`) emits valid JSON representing a cfg file. 

Mk mode (`-mk`) generates a valid [cfg](https://github.com/seh-msft/cfg) file with identifiers for one or more OpenAPI JSON specification files. 

## Build

	go build

## Install

	go install

## Usage

For mk mode:

If `-api` is not specified, one or more OpenAPI specification files should be passed as commandline arguments. 

In `bash(1)`, it suffices to glob astericks on a directory. 

For JSON mode:

If `-cfg` is not specified, a cfg file must be passed as a commandline argument. 

```
Usage of mkcfg:
  -all
        Include every parameter in the output (mk)
  -api string
        Input .json OpenAPI specification file (mk)
  -cautious

  -cfg string
        Input .cfg file (json)
  -json
        Convert a cfg file to JSON
  -minimal
        If not in strict mode, do not emit exclusivity parameters (mk)
  -mk
        Generate a new cfg file (default)
  -o string
        Output file
  -single
        Force usage of single quoting
  -strict
        Generate a strict cfg allowlisting explicit path:title combinations (mk)
```

## Examples

Generate a loose cfg for a directory of two JSON specifications:

```
$ go run cfg.go ./jsons/*
# Identifiers for the API "My API":

"accountId"=
        disallow path=".*" title=".*"
        permit title="My API"

"Authorization"=
        disallow path=".*" title=".*"
        permit title="My API"

# Identifiers for the API "Your API":

"objectType"=
        disallow path=".*" title=".*"
        permit title="Your API"

"objectId"=
        disallow path=".*" title=".*"
        permit title="Your API"

$
```

Note: Even loose mode constrains an identifier to its original API, by default. This option is configurable with the `-noapi` flag. 
