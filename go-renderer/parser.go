package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Chart represents a complete CML chart
type Chart struct {
	Meta       []MetaEntry
	Settings   []SettingsEntry
	Bars       []Bar
	Drawings   []Drawing
	Indicators []Indicator
}

// GetBarType returns the bar type from settings, defaulting to "candlestick"
func (c *Chart) GetBarType() string {
	for _, entry := range c.Settings {
		if entry.Key == "bar-type" {
			if str, ok := entry.Value.(string); ok {
				return str
			}
		}
	}
	return "candlestick"
}

// GetGridConfig returns the grid configuration from meta, with defaults
func (c *Chart) GetGridConfig() GridConfig {
	defaultConfig := GridConfig{
		Enabled:   true,
		LineWidth: 0.5,
		Color:     "#000000",
		Opacity:   1.0,
	}

	for _, entry := range c.Settings {
		if entry.Key == "grid" {
			if config, ok := entry.Value.(GridConfig); ok {
				// Apply defaults for missing values
				if config.LineWidth == 0 {
					config.LineWidth = defaultConfig.LineWidth
				}
				if config.Color == "" {
					config.Color = defaultConfig.Color
				}
				if config.Opacity == 0 {
					config.Opacity = defaultConfig.Opacity
				}
				return config
			}
		}
	}
	return defaultConfig
}

// GetYAxisConfig returns the Y-axis configuration from settings, with defaults
func (c *Chart) GetYAxisConfig() YAxisConfig {
	defaultConfig := YAxisConfig{
		Precision: 2, // Default 2 decimal places
	}

	for _, entry := range c.Settings {
		if entry.Key == "y-axis-precision" {
			if config, ok := entry.Value.(YAxisConfig); ok {
				// Apply defaults for missing values
				if config.Precision == 0 {
					config.Precision = defaultConfig.Precision
				}
				return config
			}
		}
	}
	return defaultConfig
}

// GetBarOpacityConfig returns the bar opacity configuration
func (c *Chart) GetBarOpacityConfig() BarOpacityConfig {
	defaultConfig := BarOpacityConfig{
		Opacity: 1.0, // Default full opacity
	}

	for _, entry := range c.Settings {
		if entry.Key == "bar-opacity" {
			if config, ok := entry.Value.(BarOpacityConfig); ok {
				// Apply defaults for missing values
				if config.Opacity == 0 {
					config.Opacity = defaultConfig.Opacity
				}
				return config
			}
		}
	}
	return defaultConfig
}

// MetaEntry represents a metadata entry
type MetaEntry struct {
	Key   string
	Value interface{}
}

type SettingsEntry struct {
	Key   string
	Value interface{}
}

// GridConfig represents grid configuration
type GridConfig struct {
	Enabled   bool
	LineWidth float64
	Color     string
	Opacity   float64
}

// YAxisConfig represents Y-axis configuration
type YAxisConfig struct {
	Precision int
}

// BarOpacityConfig represents bar opacity configuration
type BarOpacityConfig struct {
	Opacity float64
}

// Bar represents OHLC price data
type Bar struct {
	DateTime time.Time
	Open     float64
	High     float64
	Low      float64
	Close    float64
}

// Drawing represents any drawing element
type Drawing interface {
	GetType() string
}

// Rectangle represents a rectangle drawing
type Rectangle struct {
	StartTime  time.Time
	StartPrice float64
	EndTime    time.Time
	EndPrice   float64
	Styles     map[string]interface{}
}

func (r Rectangle) GetType() string { return "rectangle" }

// Line represents a line drawing
type Line struct {
	StartTime  time.Time
	StartPrice float64
	EndTime    time.Time
	EndPrice   float64
	Arrow      string
	LineStyle  string
	Styles     map[string]interface{}
}

func (l Line) GetType() string { return "line" }

// ContinuousLine represents a continuous line drawing
type ContinuousLine struct {
	StartTime  time.Time
	StartPrice float64
	EndTime    time.Time
	EndPrice   float64
	LineStyle  string
	Styles     map[string]interface{}
}

func (cl ContinuousLine) GetType() string { return "continuous-line" }

// Triangle represents a triangle marker
type Triangle struct {
	DateTime  time.Time
	Direction string // "uptick" or "downtick"
	Styles    map[string]interface{}
}

func (t Triangle) GetType() string { return "triangle" }

// Circle represents a circle marker
type Circle struct {
	DateTime time.Time
	Position string // "under" or "over"
	Styles   map[string]interface{}
}

func (c Circle) GetType() string { return "circle" }

// Note represents a text note
type Note struct {
	DateTime time.Time
	Text     string
	Position string // "under" or "over"
	Styles   map[string]interface{}
}

func (n Note) GetType() string { return "note" }

// Indicator represents a technical indicator
type Indicator struct {
	Name       string
	Parameters map[string]interface{}
}

// CMLParser handles parsing of CML content
type CMLParser struct {
	datetimeRegex *regexp.Regexp
	colorRegex    *regexp.Regexp
}

// NewCMLParser creates a new CML parser
func NewCMLParser() *CMLParser {
	return &CMLParser{
		datetimeRegex: regexp.MustCompile(`(\d{4})/(\d{2})/(\d{2})\s+(\d{2}):(\d{2})(?::(\d{2}))?`),
		colorRegex:    regexp.MustCompile(`#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})`),
	}
}

// Parse parses CML content and returns a Chart
func (p *CMLParser) Parse(content string) (*Chart, error) {
	lines := strings.Split(content, "\n")
	chart := &Chart{
		Meta:       []MetaEntry{},
		Settings:   []SettingsEntry{},
		Bars:       []Bar{},
		Drawings:   []Drawing{},
		Indicators: []Indicator{},
	}

	var currentSection string
	var i int

	for i < len(lines) {
		originalLine := lines[i]
		line := strings.TrimSpace(originalLine)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			i++
			continue
		}

		// Check for section headers (only if not indented)
		if strings.HasSuffix(line, ":") && !strings.HasPrefix(originalLine, " ") && !strings.HasPrefix(originalLine, "\t") {
			currentSection = strings.TrimSuffix(line, ":")
			i++
			continue
		}

		// Parse based on current section
		switch currentSection {
		case "meta":
			meta, err := p.parseMetaEntry(line)
			if err != nil {
				return nil, fmt.Errorf("error parsing meta entry: %v", err)
			}
			chart.Meta = append(chart.Meta, meta)
		case "settings":
			settings, err := p.parseSettingsEntry(line)
			if err != nil {
				return nil, fmt.Errorf("error parsing settings entry: %v", err)
			}
			chart.Settings = append(chart.Settings, settings)

			// Check if this is a grid configuration with indented properties
			if settings.Key == "grid" {
				gridConfig := settings.Value.(GridConfig)
				// Check if it's an empty config (new indented format)
				if !gridConfig.Enabled && gridConfig.LineWidth == 0 && gridConfig.Color == "" && gridConfig.Opacity == 0 {
					// Parse indented grid properties
					gridConfig, err := p.parseIndentedGridProperties(lines, &i)
					if err != nil {
						return nil, fmt.Errorf("error parsing grid properties: %v", err)
					}
					// Update the last settings entry with the parsed grid config
					chart.Settings[len(chart.Settings)-1].Value = gridConfig
				}
			}
		case "bars":
			bar, err := p.parseBar(line)
			if err != nil {
				return nil, fmt.Errorf("error parsing bar: %v", err)
			}
			chart.Bars = append(chart.Bars, bar)
		case "drawings":
			drawing, err := p.parseDrawing(lines, &i)
			if err != nil {
				return nil, fmt.Errorf("error parsing drawing: %v", err)
			}
			chart.Drawings = append(chart.Drawings, drawing)
		case "indicators":
			indicator, err := p.parseIndicator(line)
			if err != nil {
				return nil, fmt.Errorf("error parsing indicator: %v", err)
			}
			chart.Indicators = append(chart.Indicators, indicator)
		}
		i++
	}

	return chart, nil
}

// parseMetaEntry parses a metadata entry
func (p *CMLParser) parseMetaEntry(line string) (MetaEntry, error) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return MetaEntry{}, fmt.Errorf("invalid meta entry format: %s", line)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// Check if it's a grid configuration
	if key == "grid" && strings.HasPrefix(value, "grid(") && strings.HasSuffix(value, ")") {
		config, err := p.parseGridConfig(value)
		if err != nil {
			return MetaEntry{}, err
		}
		return MetaEntry{Key: key, Value: config}, nil
	}

	// Remove quotes if present
	if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		value = value[1 : len(value)-1]
	} else {
		// Try to parse as number
		if num, err := strconv.ParseFloat(value, 64); err == nil {
			return MetaEntry{Key: key, Value: num}, nil
		}
	}

	return MetaEntry{Key: key, Value: value}, nil
}

// parseSettingsEntry parses a settings entry
func (p *CMLParser) parseSettingsEntry(line string) (SettingsEntry, error) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return SettingsEntry{}, fmt.Errorf("invalid settings entry format: %s", line)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// Check if it's a bar type
	if key == "bar-type" && (value == "candlestick" || value == "heikin-ashi" || value == "ohlc") {
		return SettingsEntry{Key: key, Value: value}, nil
	}

	// Check if it's a y-axis precision (just a number)
	if key == "y-axis-precision" {
		if precision, err := strconv.Atoi(value); err == nil {
			return SettingsEntry{Key: key, Value: YAxisConfig{Precision: precision}}, nil
		}
	}

	// Check if it's a bar opacity (just a number)
	if key == "bar-opacity" {
		if opacity, err := strconv.ParseFloat(value, 64); err == nil {
			return SettingsEntry{Key: key, Value: BarOpacityConfig{Opacity: opacity}}, nil
		}
	}

	// Check if it's a grid configuration
	if key == "grid" {
		// Handle both old format: grid: (enabled=true, ...) and new format: grid: (no value, properties on next lines)
		if value == "" {
			// New indented format - return empty config, will be populated by subsequent lines
			return SettingsEntry{Key: key, Value: GridConfig{}}, nil
		} else if strings.HasPrefix(value, "(") && strings.HasSuffix(value, ")") {
			// Old inline format
			config, err := p.parseGridConfig("grid" + value)
			if err != nil {
				return SettingsEntry{}, err
			}
			return SettingsEntry{Key: key, Value: config}, nil
		}
	}

	return SettingsEntry{}, fmt.Errorf("unknown settings key: %s", key)
}

// parseIndentedGridProperties parses indented grid properties
func (p *CMLParser) parseIndentedGridProperties(lines []string, i *int) (GridConfig, error) {
	config := GridConfig{}

	// Look ahead for indented lines
	for *i+1 < len(lines) {
		nextLine := strings.TrimSpace(lines[*i+1])

		// Check if line is indented (starts with spaces/tabs)
		if nextLine == "" || !strings.HasPrefix(lines[*i+1], " ") && !strings.HasPrefix(lines[*i+1], "\t") {
			break
		}

		*i++ // Move to next line

		// Parse grid property
		parts := strings.SplitN(nextLine, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "enabled":
			if value == "true" {
				config.Enabled = true
			} else if value == "false" {
				config.Enabled = false
			}
		case "line-width":
			if width, err := strconv.ParseFloat(value, 64); err == nil {
				config.LineWidth = width
			}
		case "color":
			config.Color = value
		case "opacity":
			if opacity, err := strconv.ParseFloat(value, 64); err == nil {
				config.Opacity = opacity
			}
		}
	}

	return config, nil
}

// parseGridConfig parses a grid configuration
func (p *CMLParser) parseGridConfig(value string) (GridConfig, error) {
	// Remove "grid(" and ")"
	content := strings.TrimPrefix(value, "grid(")
	content = strings.TrimSuffix(content, ")")

	config := GridConfig{
		Enabled:   true,      // Default enabled
		LineWidth: 0.5,       // Default line width
		Color:     "#000000", // Default color (black)
		Opacity:   1.0,       // Default opacity (fully opaque)
	}

	if content == "" {
		return config, nil
	}

	// Parse properties
	properties := strings.Split(content, ",")
	for _, prop := range properties {
		prop = strings.TrimSpace(prop)
		parts := strings.SplitN(prop, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "enabled":
			config.Enabled = (val == "true")
		case "line-width":
			if width, err := strconv.ParseFloat(val, 64); err == nil {
				config.LineWidth = width
			}
		case "color":
			config.Color = val
		case "opacity":
			if opacity, err := strconv.ParseFloat(val, 64); err == nil {
				config.Opacity = opacity
			}
		}
	}

	return config, nil
}

// parseYAxisConfig parses a Y-axis configuration
func (p *CMLParser) parseYAxisConfig(value string) (YAxisConfig, error) {
	// Remove "y-axis-precision(" and ")"
	content := strings.TrimPrefix(value, "y-axis-precision(")
	content = strings.TrimSuffix(content, ")")

	config := YAxisConfig{
		Precision: 2, // Default 2 decimal places
	}

	if content == "" {
		return config, nil
	}

	// Parse properties
	properties := strings.Split(content, ",")
	for _, prop := range properties {
		prop = strings.TrimSpace(prop)
		parts := strings.SplitN(prop, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "precision":
			if precision, err := strconv.Atoi(val); err == nil {
				config.Precision = precision
			}
		}
	}

	return config, nil
}

// parseBar parses a price bar
func (p *CMLParser) parseBar(line string) (Bar, error) {
	parts := strings.Split(line, ",")
	if len(parts) != 5 {
		return Bar{}, fmt.Errorf("invalid bar format: %s", line)
	}

	// Parse datetime
	dt, err := p.parseDateTime(strings.TrimSpace(parts[0]))
	if err != nil {
		return Bar{}, fmt.Errorf("error parsing datetime: %v", err)
	}

	// Parse OHLC values
	open, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return Bar{}, fmt.Errorf("error parsing open price: %v", err)
	}

	high, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	if err != nil {
		return Bar{}, fmt.Errorf("error parsing high price: %v", err)
	}

	low, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
	if err != nil {
		return Bar{}, fmt.Errorf("error parsing low price: %v", err)
	}

	close, err := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)
	if err != nil {
		return Bar{}, fmt.Errorf("error parsing close price: %v", err)
	}

	return Bar{
		DateTime: dt,
		Open:     open,
		High:     high,
		Low:      low,
		Close:    close,
	}, nil
}

// parseDrawing parses a drawing element
func (p *CMLParser) parseDrawing(lines []string, i *int) (Drawing, error) {
	line := strings.TrimSpace(lines[*i])

	// Parse styles from subsequent lines
	styles := make(map[string]interface{})
	*i++
	for *i < len(lines) {
		styleLine := strings.TrimSpace(lines[*i])
		if styleLine == "" || strings.HasPrefix(styleLine, "#") {
			break
		}

		// Check if this is a new drawing (no indentation and contains parentheses)
		if !strings.HasPrefix(styleLine, " ") && !strings.HasPrefix(styleLine, "\t") && strings.Contains(styleLine, "(") {
			*i-- // Back up one line
			break
		}

		// Parse style property
		parts := strings.SplitN(styleLine, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Try to parse as number
			if num, err := strconv.ParseFloat(value, 64); err == nil {
				styles[key] = num
			} else {
				styles[key] = value
			}
		}
		*i++
	}

	// Parse the drawing type and parameters
	if strings.HasPrefix(line, "rectangle(") {
		return p.parseRectangle(line, styles)
	} else if strings.HasPrefix(line, "line(") {
		return p.parseLine(line, styles)
	} else if strings.HasPrefix(line, "continuous-line(") {
		return p.parseContinuousLine(line, styles)
	} else if strings.HasPrefix(line, "uptick-triangle(") {
		return p.parseTriangle(line, "uptick", styles)
	} else if strings.HasPrefix(line, "downtick-triangle(") {
		return p.parseTriangle(line, "downtick", styles)
	} else if strings.HasPrefix(line, "undercircle(") {
		return p.parseCircle(line, "under", styles)
	} else if strings.HasPrefix(line, "overcircle(") {
		return p.parseCircle(line, "over", styles)
	} else if strings.HasPrefix(line, "undernote(") {
		return p.parseNote(line, "under", styles)
	} else if strings.HasPrefix(line, "overnote(") {
		return p.parseNote(line, "over", styles)
	}

	return nil, fmt.Errorf("unknown drawing type: %s", line)
}

// parseRectangle parses a rectangle drawing
func (p *CMLParser) parseRectangle(line string, styles map[string]interface{}) (Drawing, error) {
	// Extract parameters from rectangle(datetime1,price1;datetime2,price2)
	content := strings.TrimPrefix(line, "rectangle(")
	content = strings.TrimSuffix(content, ")")

	parts := strings.Split(content, ";")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid rectangle format")
	}

	// Parse start point
	startParts := strings.Split(parts[0], ",")
	if len(startParts) != 2 {
		return nil, fmt.Errorf("invalid rectangle start point")
	}

	startTime, err := p.parseDateTime(strings.TrimSpace(startParts[0]))
	if err != nil {
		return nil, err
	}

	startPrice, err := strconv.ParseFloat(strings.TrimSpace(startParts[1]), 64)
	if err != nil {
		return nil, err
	}

	// Parse end point
	endParts := strings.Split(parts[1], ",")
	if len(endParts) != 2 {
		return nil, fmt.Errorf("invalid rectangle end point")
	}

	endTime, err := p.parseDateTime(strings.TrimSpace(endParts[0]))
	if err != nil {
		return nil, err
	}

	endPrice, err := strconv.ParseFloat(strings.TrimSpace(endParts[1]), 64)
	if err != nil {
		return nil, err
	}

	return Rectangle{
		StartTime:  startTime,
		StartPrice: startPrice,
		EndTime:    endTime,
		EndPrice:   endPrice,
		Styles:     styles,
	}, nil
}

// parseLine parses a line drawing
func (p *CMLParser) parseLine(line string, styles map[string]interface{}) (Drawing, error) {
	// Similar to rectangle but with arrow and line style support
	content := strings.TrimPrefix(line, "line(")
	content = strings.TrimSuffix(content, ")")

	parts := strings.Split(content, ";")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid line format")
	}

	// Parse start and end points (similar to rectangle)
	startParts := strings.Split(parts[0], ",")
	if len(startParts) != 2 {
		return nil, fmt.Errorf("invalid line start point")
	}

	startTime, err := p.parseDateTime(strings.TrimSpace(startParts[0]))
	if err != nil {
		return nil, err
	}

	startPrice, err := strconv.ParseFloat(strings.TrimSpace(startParts[1]), 64)
	if err != nil {
		return nil, err
	}

	endParts := strings.Split(parts[1], ",")
	if len(endParts) != 2 {
		return nil, fmt.Errorf("invalid line end point")
	}

	endTime, err := p.parseDateTime(strings.TrimSpace(endParts[0]))
	if err != nil {
		return nil, err
	}

	endPrice, err := strconv.ParseFloat(strings.TrimSpace(endParts[1]), 64)
	if err != nil {
		return nil, err
	}

	// Extract arrow properties and line style from styles
	leftArrow := false
	rightArrow := false
	if val, ok := styles["left-arrow"]; ok {
		if str, ok := val.(string); ok && str == "true" {
			leftArrow = true
		}
	}
	if val, ok := styles["right-arrow"]; ok {
		if str, ok := val.(string); ok && str == "true" {
			rightArrow = true
		}
	}

	lineStyle := ""
	if val, ok := styles["style"]; ok {
		lineStyle = val.(string)
	}

	// Determine arrow type based on properties
	arrow := ""
	if leftArrow && rightArrow {
		arrow = "both-arrows"
	} else if leftArrow {
		arrow = "left-arrow"
	} else if rightArrow {
		arrow = "right-arrow"
	}

	return Line{
		StartTime:  startTime,
		StartPrice: startPrice,
		EndTime:    endTime,
		EndPrice:   endPrice,
		Arrow:      arrow,
		LineStyle:  lineStyle,
		Styles:     styles,
	}, nil
}

// parseContinuousLine parses a continuous line drawing
func (p *CMLParser) parseContinuousLine(line string, styles map[string]interface{}) (Drawing, error) {
	// Similar to line but without arrow support
	content := strings.TrimPrefix(line, "continuous-line(")
	content = strings.TrimSuffix(content, ")")

	parts := strings.Split(content, ";")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid continuous line format")
	}

	// Parse start and end points (same as line)
	startParts := strings.Split(parts[0], ",")
	if len(startParts) != 2 {
		return nil, fmt.Errorf("invalid continuous line start point")
	}

	startTime, err := p.parseDateTime(strings.TrimSpace(startParts[0]))
	if err != nil {
		return nil, err
	}

	startPrice, err := strconv.ParseFloat(strings.TrimSpace(startParts[1]), 64)
	if err != nil {
		return nil, err
	}

	endParts := strings.Split(parts[1], ",")
	if len(endParts) != 2 {
		return nil, fmt.Errorf("invalid continuous line end point")
	}

	endTime, err := p.parseDateTime(strings.TrimSpace(endParts[0]))
	if err != nil {
		return nil, err
	}

	endPrice, err := strconv.ParseFloat(strings.TrimSpace(endParts[1]), 64)
	if err != nil {
		return nil, err
	}

	lineStyle := ""
	if val, ok := styles["style"]; ok {
		lineStyle = val.(string)
	}

	return ContinuousLine{
		StartTime:  startTime,
		StartPrice: startPrice,
		EndTime:    endTime,
		EndPrice:   endPrice,
		LineStyle:  lineStyle,
		Styles:     styles,
	}, nil
}

// parseTriangle parses a triangle marker
func (p *CMLParser) parseTriangle(line string, direction string, styles map[string]interface{}) (Drawing, error) {
	content := strings.TrimPrefix(line, direction+"-triangle(")
	content = strings.TrimSuffix(content, ")")

	dt, err := p.parseDateTime(strings.TrimSpace(content))
	if err != nil {
		return nil, err
	}

	return Triangle{
		DateTime:  dt,
		Direction: direction,
		Styles:    styles,
	}, nil
}

// parseCircle parses a circle marker
func (p *CMLParser) parseCircle(line string, position string, styles map[string]interface{}) (Drawing, error) {
	content := strings.TrimPrefix(line, position+"circle(")
	content = strings.TrimSuffix(content, ")")

	dt, err := p.parseDateTime(strings.TrimSpace(content))
	if err != nil {
		return nil, err
	}

	return Circle{
		DateTime: dt,
		Position: position,
		Styles:   styles,
	}, nil
}

// parseNote parses a text note
func (p *CMLParser) parseNote(line string, position string, styles map[string]interface{}) (Drawing, error) {
	content := strings.TrimPrefix(line, position+"note(")
	content = strings.TrimSuffix(content, ")")

	parts := strings.SplitN(content, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid note format")
	}

	dt, err := p.parseDateTime(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, err
	}

	text := strings.TrimSpace(parts[1])
	// Remove quotes if present
	if strings.HasPrefix(text, `"`) && strings.HasSuffix(text, `"`) {
		text = text[1 : len(text)-1]
	}

	return Note{
		DateTime: dt,
		Text:     text,
		Position: position,
		Styles:   styles,
	}, nil
}

// parseIndicator parses a technical indicator
func (p *CMLParser) parseIndicator(line string) (Indicator, error) {
	// Extract indicator name and parameters
	openParen := strings.Index(line, "(")
	if openParen == -1 {
		return Indicator{}, fmt.Errorf("invalid indicator format: %s", line)
	}

	name := strings.TrimSpace(line[:openParen])
	paramsStr := strings.TrimSpace(line[openParen+1:])
	paramsStr = strings.TrimSuffix(paramsStr, ")")

	parameters := make(map[string]interface{})

	if paramsStr != "" {
		params := strings.Split(paramsStr, ",")
		for _, param := range params {
			parts := strings.SplitN(strings.TrimSpace(param), "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Try to parse as number
				if num, err := strconv.ParseFloat(value, 64); err == nil {
					parameters[key] = num
				} else {
					parameters[key] = value
				}
			}
		}
	}

	return Indicator{
		Name:       name,
		Parameters: parameters,
	}, nil
}

// parseDateTime parses a datetime string in format YYYY/DD/MM HH:MM[:SS]
func (p *CMLParser) parseDateTime(dtStr string) (time.Time, error) {
	matches := p.datetimeRegex.FindStringSubmatch(dtStr)
	if len(matches) < 6 {
		return time.Time{}, fmt.Errorf("invalid datetime format: %s", dtStr)
	}

	year, _ := strconv.Atoi(matches[1])
	month, _ := strconv.Atoi(matches[2])
	day, _ := strconv.Atoi(matches[3])
	hour, _ := strconv.Atoi(matches[4])
	minute, _ := strconv.Atoi(matches[5])

	second := 0
	if len(matches) > 6 && matches[6] != "" {
		second, _ = strconv.Atoi(matches[6])
	}

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC), nil
}

// parseBarOpacityConfig parses a bar opacity configuration
func (p *CMLParser) parseBarOpacityConfig(value string) (BarOpacityConfig, error) {
	// Remove "bar-opacity(" and ")"
	content := strings.TrimPrefix(value, "bar-opacity(")
	content = strings.TrimSuffix(content, ")")

	config := BarOpacityConfig{
		Opacity: 1.0, // Default full opacity
	}

	if content == "" {
		return config, nil
	}

	// Parse properties
	properties := strings.Split(content, ",")
	for _, prop := range properties {
		prop = strings.TrimSpace(prop)
		parts := strings.SplitN(prop, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "opacity":
			if opacity, err := strconv.ParseFloat(val, 64); err == nil {
				config.Opacity = opacity
			}
		}
	}

	return config, nil
}
