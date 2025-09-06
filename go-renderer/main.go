package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

// Version information set at build time
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitRef    = "unknown"
)

func main() {
	fmt.Printf("DEBUG: Main function started\n")
	if len(os.Args) < 2 {
		fmt.Println("Usage: cml-renderer <input.cml> [output.png]")
		fmt.Println("Example: cml-renderer example.cml chart.png")
		fmt.Println("")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Ref: %s\n", GitRef)
		os.Exit(1)
	}

	// Handle version flag
	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		fmt.Printf("cml-renderer version %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Ref: %s\n", GitRef)
		os.Exit(0)
	}

	inputFile := os.Args[1]
	outputFile := "output.png"

	if len(os.Args) > 2 {
		outputFile = os.Args[2]
	}

	// Read the CML file
	content, err := ioutil.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", inputFile, err)
		os.Exit(1)
	}

	// Parse the CML content
	parser := NewCMLParser()
	chart, err := parser.Parse(string(content))
	if err != nil {
		fmt.Printf("Error parsing CML: %v\n", err)
		os.Exit(1)
	}

	// Render the chart
	renderer := NewCMLRenderer(800, 600)
	err = renderer.Render(chart, outputFile)
	if err != nil {
		fmt.Printf("Error rendering chart: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Chart rendered successfully to %s\n", outputFile)
}
