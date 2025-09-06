"""
Chart Markup Language Parser

This module provides a parser for the Chart Markup Language (CML) that can
parse CML files and convert them into structured data for rendering.
"""

import re
from typing import Dict, List, Optional, Union, Any
from dataclasses import dataclass
from datetime import datetime
import pyparsing as pp


@dataclass
class GridConfig:
    """Grid configuration."""
    enabled: bool = True
    line_width: float = 0.5
    color: str = "#000000"
    opacity: float = 1.0


@dataclass
class YAxisConfig:
    """Y-axis configuration."""
    precision: int = 2


@dataclass
class BarOpacityConfig:
    """Bar opacity configuration."""
    opacity: float = 1.0


@dataclass
class MetaEntry:
    """Represents a metadata entry."""
    key: str
    value: Union[str, float, GridConfig]

@dataclass
class SettingsEntry:
    """Represents a settings entry."""
    key: str
    value: Union[str, YAxisConfig]


@dataclass
class Bar:
    """Represents a price bar with OHLC data."""
    datetime: datetime
    open: float
    high: float
    low: float
    close: float


@dataclass
class Drawing:
    """Base class for all drawing types."""
    pass


@dataclass
class Rectangle(Drawing):
    """Rectangle drawing."""
    start_time: datetime
    start_price: float
    end_time: datetime
    end_price: float
    styles: Dict[str, Any]


@dataclass
class Line(Drawing):
    """Line drawing."""
    start_time: datetime
    start_price: float
    end_time: datetime
    end_price: float
    arrow: Optional[str] = None
    line_style: Optional[str] = None
    styles: Optional[Dict[str, Any]] = None


@dataclass
class ContinuousLine(Drawing):
    """Continuous line drawing."""
    start_time: datetime
    start_price: float
    end_time: datetime
    end_price: float
    line_style: Optional[str] = None
    styles: Optional[Dict[str, Any]] = None


@dataclass
class Triangle(Drawing):
    """Triangle marker."""
    datetime: datetime
    direction: str  # 'uptick' or 'downtick'
    styles: Optional[Dict[str, Any]] = None


@dataclass
class Circle(Drawing):
    """Circle marker."""
    datetime: datetime
    position: str  # 'under' or 'over'
    styles: Optional[Dict[str, Any]] = None


@dataclass
class Note(Drawing):
    """Text note."""
    datetime: datetime
    text: str
    position: str  # 'under' or 'over'
    styles: Optional[Dict[str, Any]] = None


@dataclass
class Indicator:
    """Technical indicator."""
    name: str
    parameters: Dict[str, Union[str, float]]


@dataclass
class Chart:
    """Complete chart representation."""
    meta: List[MetaEntry]
    settings: List[SettingsEntry]
    bars: List[Bar]
    drawings: List[Drawing]
    indicators: List[Indicator]
    
    def get_bar_type(self) -> str:
        """Get the bar type from settings, defaulting to 'candlestick'."""
        for entry in self.settings:
            if entry.key == "bar-type":
                return str(entry.value)
        return "candlestick"
    
    def get_grid_config(self) -> GridConfig:
        """Get the grid configuration from meta, with defaults."""
        default_config = GridConfig()
        for entry in self.meta:
            if entry.key == "grid" and isinstance(entry.value, GridConfig):
                return entry.value
        return default_config
    
    def get_y_axis_config(self) -> YAxisConfig:
        """Get the Y-axis configuration from settings, with defaults."""
        default_config = YAxisConfig()
        for entry in self.settings:
            if entry.key == "y-axis-precision":
                if isinstance(entry.value, int):
                    return YAxisConfig(precision=entry.value)
                elif isinstance(entry.value, YAxisConfig):
                    return entry.value
        return default_config

    def get_bar_opacity_config(self) -> BarOpacityConfig:
        """Get the bar opacity configuration from settings, with defaults."""
        default_config = BarOpacityConfig()
        for entry in self.settings:
            if entry.key == "bar-opacity":
                if isinstance(entry.value, (int, float)):
                    return BarOpacityConfig(opacity=float(entry.value))
                elif isinstance(entry.value, BarOpacityConfig):
                    return entry.value
        return default_config


class CMLParser:
    """Parser for Chart Markup Language."""
    
    def __init__(self):
        self._setup_grammar()
    
    def _setup_grammar(self):
        """Setup the parsing grammar."""
        # Basic tokens
        identifier = pp.Word(pp.alphas, pp.alphanums + "_")
        number = pp.Combine(pp.Word(pp.nums) + pp.Optional("." + pp.Word(pp.nums)))
        quoted_string = pp.QuotedString('"', escChar='\\')
        color = pp.Literal("#") + pp.Word(pp.hexnums, exact=3) | pp.Literal("#") + pp.Word(pp.hexnums, exact=6)
        boolean = pp.Literal("true") | pp.Literal("false")
        line_style = pp.Literal("solid") | pp.Literal("dashed") | pp.Literal("dotted")
        
        # DateTime format: YYYY/MM/DD HH:MM[:SS]
        datetime_pattern = pp.Combine(
            pp.Word(pp.nums, exact=4) + "/" +
            pp.Word(pp.nums, exact=2) + "/" +
            pp.Word(pp.nums, exact=2) + " " +
            pp.Word(pp.nums, exact=2) + ":" +
            pp.Word(pp.nums, exact=2) +
            pp.Optional(":" + pp.Word(pp.nums, exact=2))
        )
        
        # Value can be quoted string, number, bar type, or complex config
        bar_type = pp.Literal("candlestick") | pp.Literal("heikin-ashi") | pp.Literal("ohlc")
        
        # Complex configurations (grid, y-axis) - parse as strings and handle in parse_value
        complex_config = pp.Combine(
            pp.Literal("grid(") + pp.SkipTo(")") + ")" |
            pp.Literal("y-axis-precision(") + pp.SkipTo(")") + ")"
        )
        
        meta_value = quoted_string | number | complex_config
        settings_value = bar_type | complex_config
        
        # Meta entry
        meta_entry = identifier + ":" + meta_value
        meta_section = pp.Literal("meta:") + pp.OneOrMore(meta_entry)
        
        # Settings entry
        settings_entry = identifier + ":" + settings_value
        settings_section = pp.Literal("settings:") + pp.OneOrMore(settings_entry)
        
        # Bar: datetime, open, high, low, close
        bar = datetime_pattern + "," + number + "," + number + "," + number + "," + number
        bars_section = pp.Literal("bars:") + pp.OneOrMore(bar)
        
        # Style properties
        style_property = (
            pp.Literal("border-color=") + color |
            pp.Literal("fill-color=") + color |
            pp.Literal("line-width=") + number |
            pp.Literal("line-opacity=") + number |
            pp.Literal("fill-opacity=") + number |
            pp.Literal("font-size=") + number |
            pp.Literal("font-color=") + color |
            pp.Literal("left-arrow=") + boolean |
            pp.Literal("right-arrow=") + boolean |
            pp.Literal("style=") + line_style
        )
        
        # Drawing types
        rectangle = pp.Literal("rectangle") + "(" + datetime_pattern + "," + number + ";" + datetime_pattern + "," + number + ")"
        line = pp.Literal("line") + "(" + datetime_pattern + "," + number + ";" + datetime_pattern + "," + number + ")"
        continuous_line = pp.Literal("continuous-line") + "(" + datetime_pattern + "," + number + ";" + datetime_pattern + "," + number + ")"
        uptick_triangle = pp.Literal("uptick-triangle") + "(" + datetime_pattern + ")"
        downtick_triangle = pp.Literal("downtick-triangle") + "(" + datetime_pattern + ")"
        undercircle = pp.Literal("undercircle") + "(" + datetime_pattern + ")"
        overcircle = pp.Literal("overcircle") + "(" + datetime_pattern + ")"
        undernote = pp.Literal("undernote") + "(" + datetime_pattern + "," + quoted_string + ")"
        overnote = pp.Literal("overnote") + "(" + datetime_pattern + "," + quoted_string + ")"
        
        drawing = rectangle | line | continuous_line | uptick_triangle | downtick_triangle | undercircle | overcircle | undernote | overnote
        drawings_section = pp.Literal("drawings:") + pp.OneOrMore(drawing + pp.ZeroOrMore(style_property))
        
        # Indicators
        param = identifier + "=" + (quoted_string | number)
        indicator = identifier + "(" + pp.Optional(param + pp.ZeroOrMore("," + param)) + ")"
        indicators_section = pp.Literal("indicators:") + pp.OneOrMore(indicator)
        
        # Complete chart
        self.grammar = pp.Optional(meta_section) + pp.Optional(settings_section) + pp.Optional(bars_section) + pp.Optional(drawings_section) + pp.Optional(indicators_section)
    
    def parse_datetime(self, dt_str: str) -> datetime:
        """Parse datetime string in format YYYY/MM/DD HH:MM[:SS]."""
        # Handle optional seconds
        if ":" in dt_str.split(" ")[1] and len(dt_str.split(" ")[1].split(":")) == 3:
            return datetime.strptime(dt_str, "%Y/%m/%d %H:%M:%S")
        else:
            return datetime.strptime(dt_str, "%Y/%m/%d %H:%M")
    
    def parse_color(self, color_str: str) -> str:
        """Parse color string."""
        return color_str
    
    def parse_value(self, value) -> Union[str, float, GridConfig]:
        """Parse a value (string, number, or grid config)."""
        if isinstance(value, str):
            # Check if it's a grid configuration
            if value.startswith("grid(") and value.endswith(")"):
                return self._parse_grid_config(value)
            # Try to parse as number
            try:
                return float(value)
            except ValueError:
                return value.strip('"')
        elif isinstance(value, (list, tuple)):
            # Handle parsed configurations
            if len(value) > 0 and str(value[0]).startswith("grid"):
                return self._parse_grid_config_from_tokens(value)
        return str(value)
    
    def parse_settings_value(self, value) -> Union[str, YAxisConfig, BarOpacityConfig, GridConfig]:
        """Parse a settings value (bar type, y-axis-precision, bar-opacity, or grid config)."""
        if isinstance(value, str):
            # Check if it's a bar type
            if value in ["candlestick", "heikin-ashi", "ohlc"]:
                return value
            # Check if it's a grid configuration
            if value.startswith("grid(") and value.endswith(")"):
                return self._parse_grid_config(value)
            # Check if it's a number (for y-axis-precision or bar-opacity)
            try:
                # Try integer first (for y-axis-precision)
                return int(value)
            except ValueError:
                try:
                    # Try float (for bar-opacity)
                    return float(value)
                except ValueError:
                    pass
        elif isinstance(value, (int, float)):
            return value
        return str(value)
    
    def _parse_indented_grid_properties(self, lines: list, start_idx: int) -> GridConfig:
        """Parse indented grid properties."""
        config = GridConfig()
        
        # Look ahead for indented lines
        for i in range(start_idx + 1, len(lines)):
            line = lines[i]
            
            # Check if line is indented (starts with spaces/tabs)
            if not line.startswith(' ') and not line.startswith('\t'):
                break
                
            line = line.strip()
            if not line or line.startswith('#'):
                continue
                
            if ':' in line:
                key, value = line.split(':', 1)
                key = key.strip()
                value = value.strip()
                
                if key == 'enabled':
                    config.enabled = value.lower() == 'true'
                elif key == 'line-width':
                    try:
                        config.line_width = float(value)
                    except ValueError:
                        pass
                elif key == 'color':
                    config.color = value
                elif key == 'opacity':
                    try:
                        config.opacity = float(value)
                    except ValueError:
                        pass
        
        return config
    
    def _parse_grid_config(self, value: str) -> GridConfig:
        """Parse grid configuration from string."""
        # Remove "grid(" and ")"
        content = value[5:-1].strip()
        config = GridConfig()
        
        if content:
            properties = [prop.strip() for prop in content.split(",")]
            for prop in properties:
                if "=" in prop:
                    key, val = prop.split("=", 1)
                    key = key.strip()
                    val = val.strip()
                    
                    if key == "enabled":
                        config.enabled = val.lower() == "true"
                    elif key == "line-width":
                        config.line_width = float(val)
                    elif key == "color":
                        config.color = val
                    elif key == "opacity":
                        config.opacity = float(val)
        
        return config
    
    def _parse_yaxis_config(self, value: str) -> YAxisConfig:
        """Parse Y-axis configuration from string."""
        # Remove "y-axis-precision(" and ")"
        content = value[17:-1].strip()
        config = YAxisConfig()
        
        if content:
            properties = [prop.strip() for prop in content.split(",")]
            for prop in properties:
                if "=" in prop:
                    key, val = prop.split("=", 1)
                    key = key.strip()
                    val = val.strip()
                    
                    if key == "precision":
                        config.precision = int(val)
        
        return config
    
    def _parse_bar_opacity_config(self, value: str) -> BarOpacityConfig:
        """Parse bar opacity configuration from string."""
        # Remove "bar-opacity(" and ")"
        content = value[12:-1].strip()
        config = BarOpacityConfig()
        
        if content:
            properties = [prop.strip() for prop in content.split(",")]
            for prop in properties:
                if "=" in prop:
                    key, val = prop.split("=", 1)
                    key = key.strip()
                    val = val.strip()
                    
                    if key == "opacity":
                        config.opacity = float(val)
        
        return config
    
    def _parse_coordinate(self, coord_str: str) -> tuple:
        """Parse a coordinate string like '2025/01/15 10:00,1.2500' into (datetime, float)."""
        if ',' in coord_str:
            time_part, price_part = coord_str.split(',', 1)
            dt = self.parse_datetime(time_part.strip())
            price = float(price_part.strip())
            return dt, price
        raise ValueError(f"Invalid coordinate format: {coord_str}")
    
    def _parse_grid_config_from_tokens(self, tokens) -> GridConfig:
        """Parse grid configuration from parsed tokens."""
        config = GridConfig()
        # This is a simplified implementation - in practice you'd parse the tokens
        return config
    
    def _parse_yaxis_config_from_tokens(self, tokens) -> YAxisConfig:
        """Parse Y-axis configuration from parsed tokens."""
        config = YAxisConfig()
        # This is a simplified implementation - in practice you'd parse the tokens
        return config
    
    def _parse_simple(self, cml_content: str) -> Chart:
        """Simple line-by-line parser."""
        lines = cml_content.strip().split('\n')
        
        meta = []
        settings = []
        bars = []
        drawings = []
        indicators = []
        
        current_section = None
        
        for line_idx, line in enumerate(lines):
            line = line.strip()
            if not line or line.startswith('#'):
                continue
                
            if line.endswith(':'):
                current_section = line[:-1]
                continue
            
            if current_section == 'meta':
                if ':' in line:
                    key, value = line.split(':', 1)
                    key = key.strip()
                    value = value.strip()
                    parsed_value = self.parse_value(value)
                    meta.append(MetaEntry(key, parsed_value))
            
            elif current_section == 'settings':
                if ':' in line:
                    key, value = line.split(':', 1)
                    key = key.strip()
                    value = value.strip()
                    parsed_value = self.parse_settings_value(value)
                    settings.append(SettingsEntry(key, parsed_value))
                    
                    # Check if this is a grid configuration with indented properties
                    if key == 'grid' and value == '':
                        # Parse indented grid properties
                        grid_config = self._parse_indented_grid_properties(lines, line_idx)
                        # Update the last settings entry
                        settings[-1] = SettingsEntry(key, grid_config)
            
            elif current_section == 'bars':
                if ',' in line:
                    parts = [p.strip() for p in line.split(',')]
                    if len(parts) == 5:
                        dt = self.parse_datetime(parts[0])
                        open_price = float(parts[1])
                        high_price = float(parts[2])
                        low_price = float(parts[3])
                        close_price = float(parts[4])
                        bars.append(Bar(dt, open_price, high_price, low_price, close_price))
            
            elif current_section == 'drawings':
                # Parse drawing elements
                if '(' in line and ')' in line:
                    # Extract drawing type and parameters
                    drawing_type = line.split('(')[0].strip()
                    params_start = line.find('(')
                    params_end = line.rfind(')')
                    params_str = line[params_start+1:params_end]
                    
                    # Parse style properties that follow this drawing definition
                    styles = {}
                    # Look ahead for style properties on subsequent lines
                    for next_line_idx in range(line_idx + 1, len(lines)):
                        next_line = lines[next_line_idx].strip()
                        if not next_line or next_line.startswith('#'):
                            continue
                        if next_line.endswith(':') or '(' in next_line:
                            # Hit next section or next drawing, stop
                            break
                        if '=' in next_line:
                            # This is a style property
                            key, value = next_line.split('=', 1)
                            styles[key.strip()] = value.strip()
                    
                    # Parse the drawing based on type
                    if drawing_type == 'rectangle':
                        # Parse rectangle(start_time,start_price ; end_time,end_price)
                        if ';' in params_str:
                            start_part, end_part = params_str.split(';', 1)
                            start_time, start_price = self._parse_coordinate(start_part.strip())
                            end_time, end_price = self._parse_coordinate(end_part.strip())
                            drawings.append(Rectangle(start_time, start_price, end_time, end_price, styles))
                    
                    elif drawing_type == 'continuous-line':
                        # Parse continuous-line(start_time,start_price ; end_time,end_price)
                        if ';' in params_str:
                            start_part, end_part = params_str.split(';', 1)
                            start_time, start_price = self._parse_coordinate(start_part.strip())
                            end_time, end_price = self._parse_coordinate(end_part.strip())
                            # Parse line style
                            line_style = styles.get("style", "solid")
                            drawings.append(ContinuousLine(start_time, start_price, end_time, end_price, line_style, styles))
                    
                    elif drawing_type == 'line':
                        # Parse line(start_time,start_price ; end_time,end_price)
                        if ';' in params_str:
                            start_part, end_part = params_str.split(';', 1)
                            start_time, start_price = self._parse_coordinate(start_part.strip())
                            end_time, end_price = self._parse_coordinate(end_part.strip())
                            
                            # Parse arrow properties
                            left_arrow = styles.get("left-arrow", "false").lower() == "true"
                            right_arrow = styles.get("right-arrow", "false").lower() == "true"
                            
                            # Determine arrow type
                            arrow_type = ""
                            if left_arrow and right_arrow:
                                arrow_type = "both-arrows"
                            elif left_arrow:
                                arrow_type = "left-arrow"
                            elif right_arrow:
                                arrow_type = "right-arrow"
                            
                            # Parse line style
                            line_style = styles.get("style", "solid")
                            
                            drawings.append(Line(start_time, start_price, end_time, end_price, arrow_type, line_style, styles))
                    
                    elif drawing_type == 'uptick-triangle':
                        # Parse uptick-triangle(time)
                        time_str = params_str.strip()
                        dt = self.parse_datetime(time_str)
                        drawings.append(Triangle(dt, "uptick", styles))
                    
                    elif drawing_type == 'downtick-triangle':
                        # Parse downtick-triangle(time)
                        time_str = params_str.strip()
                        dt = self.parse_datetime(time_str)
                        drawings.append(Triangle(dt, "downtick", styles))
                    
                    elif drawing_type == 'undernote':
                        # Parse undernote(time, "text")
                        if ',' in params_str:
                            time_part, text_part = params_str.split(',', 1)
                            time_str = time_part.strip()
                            text = text_part.strip().strip('"')
                            dt = self.parse_datetime(time_str)
                            drawings.append(Note(dt, text, "under", styles))
                    
                    elif drawing_type == 'overnote':
                        # Parse overnote(time, "text")
                        if ',' in params_str:
                            time_part, text_part = params_str.split(',', 1)
                            time_str = time_part.strip()
                            text = text_part.strip().strip('"')
                            dt = self.parse_datetime(time_str)
                            drawings.append(Note(dt, text, "over", styles))
            
            elif current_section == 'indicators':
                # Parse indicator line like "ema(period=20)"
                if '(' in line and ')' in line:
                    # Extract indicator name and parameters
                    name_part = line.split('(')[0].strip()
                    params_part = line.split('(')[1].split(')')[0].strip()
                    
                    # Parse parameters
                    parameters = {}
                    if params_part:
                        for param in params_part.split(','):
                            param = param.strip()
                            if '=' in param:
                                key, value = param.split('=', 1)
                                key = key.strip()
                                value = value.strip()
                                
                                # Try to parse as number, otherwise keep as string
                                try:
                                    if '.' in value:
                                        parameters[key] = float(value)
                                    else:
                                        parameters[key] = int(value)
                                except ValueError:
                                    parameters[key] = value
                    
                    indicators.append(Indicator(name_part, parameters))
        
        return Chart(meta, settings, bars, drawings, indicators)
    
    def parse(self, cml_content: str) -> Chart:
        """Parse CML content and return a Chart object."""
        # Use a simpler line-by-line parsing approach
        return self._parse_simple(cml_content)
    
    def _build_chart(self, parsed_result) -> Chart:
        """Build Chart object from parsed result."""
        # This is a simplified implementation
        # In a full implementation, you'd properly extract and structure the parsed data
        meta = []
        bars = []
        drawings = []
        indicators = []
        
        # Process parsed tokens and build objects
        # This would be more complex in a real implementation
        
        return Chart(meta=meta, settings=settings, bars=bars, drawings=drawings, indicators=indicators)


def parse_cml_file(filename: str) -> Chart:
    """Parse a CML file and return a Chart object."""
    try:
        with open(filename, 'r', encoding='utf-8') as f:
            content = f.read()
    except FileNotFoundError:
        raise FileNotFoundError(f"CML file not found: {filename}")
    except Exception as e:
        raise Exception(f"Error reading CML file {filename}: {e}")
    
    parser = CMLParser()
    return parser.parse(content)


if __name__ == "__main__":
    # Example usage
    parser = CMLParser()
    
    # Test with a simple example
    test_cml = """
meta:
    title: "Test Chart"
    author: "Test"

bars:
    2025/01/15 09:00, 100.50, 101.20, 100.30, 101.00
"""
    
    try:
        chart = parser.parse(test_cml)
        print("Parsing successful!")
        print(f"Meta entries: {len(chart.meta)}")
        print(f"Bars: {len(chart.bars)}")
        print(f"Drawings: {len(chart.drawings)}")
        print(f"Indicators: {len(chart.indicators)}")
    except Exception as e:
        print(f"Parsing failed: {e}")
