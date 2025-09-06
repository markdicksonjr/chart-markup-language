# Python CML Renderer

A Python implementation for parsing and rendering Chart Markup Language (CML) files.

## Features

- **Parser**: Complete CML grammar parser using pyparsing
- **Renderer**: Matplotlib-based chart rendering
- **Data Structures**: Type-safe data classes for all CML elements
- **Styling**: Full support for CML styling properties

## Installation

```bash
pip install -r requirements.txt
```

## Usage

### Basic Usage

```python
from cml_parser import parse_cml_file
from cml_renderer import CMLRenderer

# Parse a CML file
chart = parse_cml_file("example.cml")

# Render to image
renderer = CMLRenderer()
renderer.render(chart, "output.png")
```

### Command Line Usage

```python
from cml_renderer import render_cml_file

# Render directly from file
render_cml_file("input.cml", "output.png")
```

## API Reference

### CMLParser

The main parser class for CML files.

```python
parser = CMLParser()
chart = parser.parse(cml_content)
```

### CMLRenderer

The main renderer class for creating visual charts.

```python
renderer = CMLRenderer(figsize=(12, 8), dpi=100)
renderer.render(chart, output_file="chart.png")
```

### Data Structures

- `Chart`: Complete chart representation
- `Bar`: OHLC price data
- `Drawing`: Base class for all drawing types
- `Rectangle`, `Line`, `Triangle`, `Circle`, `Note`: Specific drawing types
- `Indicator`: Technical indicators
- `MetaEntry`: Metadata entries

## Examples

See the `examples/` directory for sample CML files that can be rendered with this implementation.

## Dependencies

- `matplotlib`: Chart rendering
- `numpy`: Numerical operations
- `pandas`: Data manipulation
- `pyparsing`: Grammar parsing
