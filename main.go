package main

import (
	"flag"

	"github.com/pointlander/asyncapi-flogo/transform"
)

func main() {
	input := flag.String("input", "asyncapi.yml", "input async api file")
	conversionType := flag.String("type", "flogoapiapp", "conversion type like flogoapiapp or flogodescriptor")
	output := flag.String("output", ".", "path to store generated file")

	flag.Parse()
	transform.Transform(*input, *output, *conversionType)
}
