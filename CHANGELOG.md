# Changelog

All notable changes to the Chart Markup Language project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial EBNF grammar specification
- Complete documentation with examples
- Support for meta section with title, author, description, and creation date
- Support for bars section with OHLC data
- Support for drawings section with multiple drawing types:
  - Rectangle areas
  - Lines with optional arrows
  - Continuous trend lines
  - Uptick and downtick triangles
  - Under and over circles
  - Under and over notes
- Support for indicators section with parameterized indicators
- Comprehensive styling system with:
  - Color support (3-digit and 6-digit hex)
  - Line styles (solid, dashed, dotted)
  - Arrow types (none, dot, arrow)
  - Style properties (colors, opacity, line width, font settings)
- DateTime format support (YYYY/DD/MM HH:MM[:SS])
- Number and string data types
- Example files demonstrating various use cases

### Grammar Features
- EBNF-compliant grammar specification
- Support for optional sections
- Flexible parameter system for indicators
- Comprehensive style property definitions
- Clear separation of drawing types and their specific styling options

## [1.0.0] - 2025-01-15

### Added
- Initial release of Chart Markup Language
- Complete grammar specification
- Documentation and examples
- MIT License
- Contributing guidelines
