# Go CML Renderer

A Go implementation for parsing and rendering Chart Markup Language (CML) files.

## Features

- **Parser**: Complete CML grammar parser with regex-based parsing
- **Renderer**: Image-based chart rendering using gg graphics library
- **Data Structures**: Type-safe structs for all CML elements
- **Styling**: Full support for CML styling properties

## Installation

```bash
go mod tidy
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    // Parse a CML file
    parser := NewCMLParser()
    chart, err := parser.Parse(cmlContent)
    if err != nil {
        log.Fatal(err)
    }

    // Render to image
    renderer := NewCMLRenderer(800, 600)
    err = renderer.Render(chart, "output.png")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Command Line Usage

```bash
go run . example.cml output.png
```

## API Reference

### CMLParser

The main parser struct for CML files.

```go
parser := NewCMLParser()
chart, err := parser.Parse(cmlContent)
```

### CMLRenderer

The main renderer struct for creating visual charts.

```go
renderer := NewCMLRenderer(800, 600)
err := renderer.Render(chart, "chart.png")
```

### Data Structures

- `Chart`: Complete chart representation
- `Bar`: OHLC price data
- `Drawing`: Interface for all drawing types
- `Rectangle`, `Line`, `Triangle`, `Circle`, `Note`: Specific drawing types
- `Indicator`: Technical indicators
- `MetaEntry`: Metadata entries

## Examples

See the `examples/` directory for sample CML files that can be rendered with this implementation.

## Dependencies

- `github.com/fogleman/gg`: Graphics rendering
- `golang.org/x/image`: Image processing
- Standard library: File I/O, regex, time parsing
