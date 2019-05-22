# asyncapi-flogo
[AsynAPI](https://github.com/asyncapi/asyncapi) to flogo app converter tool converts given AsynAPI spec to its implementation based on flogo api/descriptor model.

Currently this tool accepts below arguments.
```sh
Usage of asyncapi-flogo:
  -type string
        conversion type like flogoapiapp or flogodescriptor (default "flogoapiapp")
  -input string
        input async api file (default "asyncapi.yml")
  -output string
        path to store generated file (default ".")
```

## Setup
To install the tool, simply open a terminal and enter the below commands
```sh
git clone https://github.com/pointlander/asyncapi-flogo.git
cd asyncapi-flogo/
go install
```

## Usage
### Flogo app api model.
```sh
cd asyncapi-flogo/
mdkir test
asyncapi-flogo -input examples/http/asyncapi.yml -type flogoapiapp -output test/
```
The resulting output is `app.go` which can be built into a working flogo application:
```sh
cd test
go build
./test
```

### Flogo app descriptor model.
```sh
cd asyncapi-flogo/
asyncapi-flogo -input examples/http/asyncapi.yml -type flogodescriptor
```
The resulting output is `flogo.json` which can be built into a working flogo application:
```sh
flogo create -f flogo.json flogoapp
cd urn:com:http:server
flogo build
./bin/urn:com:http:server
```

## Flogo Plugin Support
This tool can be integrated into [flogocli](https://github.com/project-flogo/cli).
```sh
# Install your plugin
$ flogo plugin install github.com/pointlander/asyncapi-flogo/cmd

# Run your new plugin command for api app model
$ flogo asyncapi -i asyncapi.yml -t flogoapiapp  -o test/

# Run your new plugin command for descriptor app model
$ flogo asyncapi -i asyncapi.yml -t flogodescriptor

# Remove your plugin
$ flogo plugin remove github.com/pointlander/asyncapi-flogo/cmd
```
