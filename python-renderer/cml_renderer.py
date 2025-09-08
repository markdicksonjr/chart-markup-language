"""
Chart Markup Language Renderer

This module provides a renderer for the Chart Markup Language (CML) that can
convert parsed CML data into visual charts using matplotlib.
"""

import matplotlib.pyplot as plt
import matplotlib.patches as patches
import matplotlib.dates as mdates
from matplotlib.patches import FancyBboxPatch, Circle, Polygon
from matplotlib.lines import Line2D
import numpy as np
from typing import List, Dict, Any, Optional
from datetime import datetime
import re

from cml_parser import Chart, Bar, Drawing, Rectangle, Line, ContinuousLine, Triangle, Circle, Note


class CMLRenderer:
    """Renderer for Chart Markup Language charts."""
    
    def __init__(self, figsize=(12, 8), dpi=100):
        self.figsize = figsize
        self.dpi = dpi
        self.fig = None
        self.ax = None
        self.bars = []  # Store bars for triangle positioning
    
    def render(self, chart: Chart, output_file: Optional[str] = None) -> None:
        """Render a chart to a file or display it."""
        self.fig, self.ax = plt.subplots(figsize=self.figsize, dpi=self.dpi)
        
        # Store chart and bars for later access
        self.chart = chart
        self.bars = chart.bars
        
        # Set up the chart
        self._setup_chart(chart)
        
        # Render bars based on bar-type setting
        if chart.bars:
            bar_type = chart.get_bar_type()
            if bar_type == "candlestick":
                self._render_candlesticks(chart.bars)
            elif bar_type == "ohlc":
                self._render_ohlc_bars(chart.bars)
            elif bar_type == "heikin-ashi":
                self._render_heikin_ashi(chart.bars)
            else:
                # Default to candlestick
                self._render_candlesticks(chart.bars)
        
        # Render drawings
        for drawing in chart.drawings:
            self._render_drawing(drawing)
        
        # Render indicators (placeholder)
        if chart.indicators:
            self._render_indicators(chart.indicators)
        
        # Add title from meta
        title = self._get_meta_value(chart.meta, "title")
        if title:
            self.ax.set_title(title)
        
        # Format the chart
        self._format_chart()
        
        # Save or show
        if output_file:
            plt.savefig(output_file, dpi=self.dpi, bbox_inches='tight')
            print(f"Chart saved to {output_file}")
        else:
            plt.show()
    
    def _setup_chart(self, chart: Chart) -> None:
        """Setup the basic chart structure."""
        if not chart.bars:
            return
        
        # Set up time and price ranges
        times = [mdates.date2num(bar.datetime) for bar in chart.bars]
        prices = []
        for bar in chart.bars:
            prices.extend([bar.open, bar.high, bar.low, bar.close])
        
        # Add one extra interval on each side (same as Go renderer)
        min_time = min(times)
        max_time = max(times)
        
        if len(chart.bars) > 1:
            # Calculate interval from first two bars
            interval = mdates.date2num(chart.bars[1].datetime) - mdates.date2num(chart.bars[0].datetime)
            min_time -= interval
            max_time += interval
        
        self.ax.set_xlim(min_time, max_time)
        self.ax.set_ylim(min(prices) * 0.995, max(prices) * 1.005)
        
        # Set labels
        self.ax.set_xlabel("Time")
        self.ax.set_ylabel("Price")
        
        # Configure grid
        grid_config = chart.get_grid_config()
        if grid_config.enabled:
            # Convert hex color to matplotlib format
            color = grid_config.color
            if color.startswith('#'):
                color = color[1:]  # Remove #
            
            self.ax.grid(True, 
                        color=f'#{color}', 
                        linewidth=grid_config.line_width, 
                        alpha=grid_config.opacity)
        else:
            self.ax.grid(False)
        
        # Configure Y-axis precision
        yaxis_config = chart.get_y_axis_config()
        if yaxis_config.precision != 2:  # Default is 2
            # Format Y-axis labels with custom precision
            from matplotlib.ticker import FuncFormatter
            def format_price(x, pos):
                return f"{x:.{yaxis_config.precision}f}"
            self.ax.yaxis.set_major_formatter(FuncFormatter(format_price))
    
    def _render_candlesticks(self, bars: List[Bar]) -> None:
        """Render candlestick bars."""
        if not bars:
            return
            
        # Calculate proper bar width based on time intervals
        times = [mdates.date2num(bar.datetime) for bar in bars]
        if len(times) > 1:
            avg_interval = (max(times) - min(times)) / (len(times) - 1)
            bar_width = avg_interval * 0.6  # 60% of the interval
        else:
            bar_width = 0.01  # Default small width
            
        for bar in bars:
            dt_num = mdates.date2num(bar.datetime)
            
            # Calculate body boundaries
            body_top = max(bar.open, bar.close)
            body_bottom = min(bar.open, bar.close)
            
            # Draw the upper wick (from body top to high)
            if bar.high > body_top:
                self.ax.plot([dt_num, dt_num], [body_top, bar.high], 
                            color='black', linewidth=0.8)
            
            # Draw the lower wick (from body bottom to low)
            if bar.low < body_bottom:
                self.ax.plot([dt_num, dt_num], [bar.low, body_bottom], 
                            color='black', linewidth=0.8)
            
            # Draw the open tick (left side) - thin line
            self.ax.plot([dt_num - bar_width/2, dt_num], [bar.open, bar.open], 
                        color='black', linewidth=0.8)
            
            # Draw the close tick (right side) - thin line
            self.ax.plot([dt_num, dt_num + bar_width/2], [bar.close, bar.close], 
                        color='black', linewidth=0.8)
            
            # Draw the open-close rectangle (body)
            height = abs(bar.close - bar.open)
            bottom = min(bar.open, bar.close)
            color = 'green' if bar.close >= bar.open else 'red'
            
            # Create rectangle for the body with configurable opacity
            bar_opacity_config = self.chart.get_bar_opacity_config()
            rect = patches.Rectangle(
                (dt_num - bar_width/2, bottom), bar_width, height,
                linewidth=0.8, edgecolor='black', facecolor=color, alpha=bar_opacity_config.opacity
            )
            self.ax.add_patch(rect)
    
    def _render_ohlc_bars(self, bars: List[Bar]) -> None:
        """Render OHLC bars (not candlesticks)."""
        if not bars:
            return
            
        # Calculate proper bar width based on time intervals
        times = [mdates.date2num(bar.datetime) for bar in bars]
        if len(times) > 1:
            avg_interval = (max(times) - min(times)) / (len(times) - 1)
            bar_width = avg_interval * 0.6  # 60% of the interval
        else:
            bar_width = 0.01  # Default small width
            
        for bar in bars:
            dt_num = mdates.date2num(bar.datetime)
            
            # Draw the high-low line (wick)
            self.ax.plot([dt_num, dt_num], [bar.low, bar.high], 
                        color='black', linewidth=1)
            
            # Draw the open tick (left side)
            self.ax.plot([dt_num - bar_width/2, dt_num], [bar.open, bar.open], 
                        color='black', linewidth=2)
            
            # Draw the close tick (right side)  
            self.ax.plot([dt_num, dt_num + bar_width/2], [bar.close, bar.close], 
                        color='black', linewidth=2)
            
            # No body rectangle for OHLC bars
    
    def _render_heikin_ashi(self, bars: List[Bar]) -> None:
        """Render Heikin Ashi bars (placeholder - same as candlesticks for now)."""
        # For now, render as candlesticks
        # TODO: Implement proper Heikin Ashi calculation
        self._render_candlesticks(bars)
    
    def _render_drawing(self, drawing: Drawing) -> None:
        """Render a drawing element."""
        if isinstance(drawing, Rectangle):
            self._render_rectangle(drawing)
        elif isinstance(drawing, Line):
            self._render_line(drawing)
        elif isinstance(drawing, ContinuousLine):
            self._render_continuous_line(drawing)
        elif isinstance(drawing, Triangle):
            self._render_triangle(drawing)
        elif isinstance(drawing, Circle):
            self._render_circle(drawing)
        elif isinstance(drawing, Note):
            self._render_note(drawing)
    
    def _render_rectangle(self, rect: Rectangle) -> None:
        """Render a rectangle."""
        # Convert datetime objects to matplotlib date numbers
        start_x = mdates.date2num(rect.start_time)
        end_x = mdates.date2num(rect.end_time)
        
        # Calculate rectangle dimensions
        width = end_x - start_x
        height = abs(rect.end_price - rect.start_price)
        bottom = min(rect.start_price, rect.end_price)
        
        # Get styles
        border_color = self._get_style_value(rect.styles, "border-color", "#000000")
        fill_color = self._get_style_value(rect.styles, "fill-color", "#AAAAAA")
        line_width = self._get_style_value(rect.styles, "line-width", 1)
        fill_opacity = self._get_style_value(rect.styles, "fill-opacity", 0.3)
        line_opacity = self._get_style_value(rect.styles, "line-opacity", 1.0)
        
        rect_patch = patches.Rectangle(
            (start_x, bottom), width, height,
            linewidth=line_width, edgecolor=border_color, 
            facecolor=fill_color, alpha=fill_opacity
        )
        self.ax.add_patch(rect_patch)
    
    def _render_line(self, line: Line) -> None:
        """Render a line."""
        # Get styles
        border_color = self._get_style_value(line.styles, "border-color", "#0000FF")
        line_width = self._get_style_value(line.styles, "line-width", 2)
        line_opacity = self._get_style_value(line.styles, "line-opacity", 1.0)
        
        # Determine line style
        if line.line_style == "dashed":
            self.ax.plot([mdates.date2num(line.start_time), mdates.date2num(line.end_time)], 
                        [line.start_price, line.end_price],
                        color=border_color, linewidth=line_width, 
                        alpha=line_opacity, linestyle='--')
        elif line.line_style == "dotted":
            # Use same approach as continuous lines
            self.ax.plot([mdates.date2num(line.start_time), mdates.date2num(line.end_time)], 
                        [line.start_price, line.end_price],
                        color=border_color, linewidth=line_width, 
                        alpha=line_opacity, linestyle='-', dashes=(2, 2))
        else:  # solid
            self.ax.plot([mdates.date2num(line.start_time), mdates.date2num(line.end_time)], 
                        [line.start_price, line.end_price],
                        color=border_color, linewidth=line_width, 
                        alpha=line_opacity, linestyle='-')
        
        # Draw arrows if specified
        if line.arrow:
            if line.arrow in ["left-arrow", "both-arrows"]:
                self._add_arrow(line.start_time, line.start_price, line.end_time, line.end_price, "left", border_color, line_width, line.line_style)
            if line.arrow in ["right-arrow", "both-arrows"]:
                self._add_arrow(line.start_time, line.start_price, line.end_time, line.end_price, "right", border_color, line_width, line.line_style)
    
    def _add_arrow(self, start_time, start_price, end_time, end_price, direction, color, line_width, line_style="solid"):
        """Add an arrow to a line."""
        # Convert to matplotlib date numbers
        start_x = mdates.date2num(start_time)
        end_x = mdates.date2num(end_time)
        
        # Determine arrow direction
        if direction == "left":
            # Arrow points from end to start
            xy = (start_x, start_price)
            xytext = (end_x, end_price)
        else:  # right arrow
            # Arrow points from start to end
            xy = (end_x, end_price)
            xytext = (start_x, start_price)
        
        # Add arrow with more pronounced styling
        arrow_width = float(line_width) * 1.5 if line_width else 3  # Make arrow thicker
        
        # Create arrow properties based on line style
        arrow_props = dict(arrowstyle='->', color=color, 
                          lw=arrow_width,  # Make arrow thicker
                          shrinkA=0, shrinkB=0,  # Don't shrink the arrow
                          mutation_scale=20)  # Make arrow larger
        
        # Add line style to arrow
        if line_style == "dashed":
            arrow_props['linestyle'] = '--'
        elif line_style == "dotted":
            # For dotted lines, draw simple arrow with two tiny lines
            import numpy as np
            dx = xy[0] - xytext[0]
            dy = xy[1] - xytext[1]
            length = np.sqrt(dx**2 + dy**2)
            
            if length > 0:
                # Arrow size
                arrow_size = 0.0007  # Even smaller arrow (1/3 of 0.002)
                
                # Calculate arrow direction
                dx_norm = dx / length
                dy_norm = dy / length
                
                # Calculate arrow points (sharper angle)
                perp_x = -dy_norm * arrow_size * 0.5  # Make angle sharper
                perp_y = dx_norm * arrow_size * 0.5
                
                # Draw two arrow lines
                self.ax.plot([xy[0], xy[0] - dx_norm * arrow_size + perp_x], 
                            [xy[1], xy[1] - dy_norm * arrow_size + perp_y], 
                            color=color, linewidth=arrow_width, linestyle='-')
                self.ax.plot([xy[0], xy[0] - dx_norm * arrow_size - perp_x], 
                            [xy[1], xy[1] - dy_norm * arrow_size - perp_y], 
                            color=color, linewidth=arrow_width, linestyle='-')
            
            # No arrow annotation needed
            return
        
        self.ax.annotate('', xy=xy, xytext=xytext, arrowprops=arrow_props)
    
    def _render_continuous_line(self, line: ContinuousLine) -> None:
        """Render a continuous line."""
        # Similar to regular line but without arrow support
        border_color = self._get_style_value(line.styles, "border-color", "#008000")
        line_width = self._get_style_value(line.styles, "line-width", 1)
        
        # Get the full chart time range (including padding)
        if hasattr(self, 'chart') and self.chart.bars:
            # Use the chart's time range which includes padding
            min_time = min(bar.datetime for bar in self.chart.bars)
            max_time = max(bar.datetime for bar in self.chart.bars)
            
            # Add padding (one interval on each side)
            if len(self.chart.bars) > 1:
                interval = self.chart.bars[1].datetime - self.chart.bars[0].datetime
                min_time -= interval
                max_time += interval
            
            # Convert to matplotlib date numbers
            x_start = mdates.date2num(min_time)
            x_end = mdates.date2num(max_time)
        else:
            # Fallback to original behavior
            x_start = mdates.date2num(line.start_time)
            x_end = mdates.date2num(line.end_time)
        
        # Draw horizontal line across full chart width with proper line style
        if line.line_style == "dashed":
            self.ax.plot([x_start, x_end], 
                        [line.start_price, line.start_price],
                        color=border_color, linewidth=line_width, linestyle='--')
        elif line.line_style == "dotted":
            # Use custom dash pattern for dotted lines
            self.ax.plot([x_start, x_end], 
                        [line.start_price, line.start_price],
                        color=border_color, linewidth=line_width, 
                        linestyle='-', dashes=(2, 2))
        else:  # solid
            self.ax.plot([x_start, x_end], 
                        [line.start_price, line.start_price],
                        color=border_color, linewidth=line_width, linestyle='-')
    
    def _render_triangle(self, triangle: Triangle) -> None:
        """Render a triangle marker."""
        border_color = self._get_style_value(triangle.styles, "border-color", "#000000")
        fill_color = self._get_style_value(triangle.styles, "fill-color", "#AAAAAA")
        line_width = self._get_style_value(triangle.styles, "line-width", 1)
        fill_opacity = self._get_style_value(triangle.styles, "fill-opacity", 1.0)
        line_opacity = self._get_style_value(triangle.styles, "line-opacity", 1.0)
        
        # Find the price at this time by looking at the bars
        price = None
        for bar in self.bars:
            if mdates.date2num(bar.datetime) == mdates.date2num(triangle.datetime):
                if triangle.direction == "uptick":
                    # Place uptick triangle further below the low price
                    price = bar.low - (bar.high - bar.low) * 0.25  # 25% of bar range below low
                else:
                    # Place downtick triangle further above the high price
                    price = bar.high + (bar.high - bar.low) * 0.25  # 25% of bar range above high
                break
        
        # If not found, use a reasonable default
        if price is None:
            y_min, y_max = self.ax.get_ylim()
            if triangle.direction == "uptick":
                price = y_min + (y_max - y_min) * 0.05  # Near the bottom
            else:
                price = y_max - (y_max - y_min) * 0.05  # Near the top
        
        # Create triangle points - make it larger (3% of price range instead of 1%)
        size = (self.ax.get_ylim()[1] - self.ax.get_ylim()[0]) * 0.03  # 3% of price range
        
        # Convert datetime to matplotlib date number
        dt_num = mdates.date2num(triangle.datetime)
        
        # Calculate horizontal spread for more balanced triangle
        # Use a value between the original (size/2) and equilateral (size * 0.577)
        horizontal_spread = size * 0.7  # More balanced proportion
        
        if triangle.direction == "uptick":
            # Upward triangle - positioned below the price
            points = [
                (dt_num, price + size),  # Top point
                (dt_num - horizontal_spread, price),  # Bottom left
                (dt_num + horizontal_spread, price)   # Bottom right
            ]
        else:  # downtick
            # Downward triangle - positioned above the price
            points = [
                (dt_num, price - size),  # Bottom point
                (dt_num - horizontal_spread, price),  # Top left
                (dt_num + horizontal_spread, price)   # Top right
            ]
        
        # Draw triangle with proper style properties
        triangle_patch = Polygon(points, closed=True, 
                               facecolor=fill_color, edgecolor=border_color, 
                               linewidth=line_width, alpha=fill_opacity)
        self.ax.add_patch(triangle_patch)
    
    def _render_circle(self, circle: Circle) -> None:
        """Render a circle marker."""
        border_color = self._get_style_value(circle.styles, "border-color", "#000000")
        fill_color = self._get_style_value(circle.styles, "fill-color", "#FFFF00")
        line_width = self._get_style_value(circle.styles, "line-width", 1)
        fill_opacity = self._get_style_value(circle.styles, "fill-opacity", 0.5)
        
        # This is a placeholder - you'd need to determine the actual price position
        price = 0  # This should be determined from the chart context
        
        circle_patch = patches.Circle(
            (mdates.date2num(circle.datetime), price), 0.001,
            linewidth=line_width, edgecolor=border_color,
            facecolor=fill_color, alpha=fill_opacity
        )
        self.ax.add_patch(circle_patch)
    
    def _render_note(self, note: Note) -> None:
        """Render a text note."""
        font_size = self._get_style_value(note.styles, "font-size", 12)
        font_color = self._get_style_value(note.styles, "font-color", "#000000")
        
        # Find the price at the note's time by looking at the bars
        price = None
        for bar in self.bars:
            if bar.datetime == note.datetime:
                # Add a small offset to move notes away from wicks
                bar_range = bar.high - bar.low
                small_offset = bar_range * 0.05  # 5% of bar range
                
                if note.position == 'over':
                    price = bar.high + small_offset  # Move over notes slightly above high
                else:  # under
                    price = bar.low - small_offset  # Move under notes slightly below low
                break
        
        # If no exact match, use the chart's price range
        if price is None:
            all_prices = []
            for bar in self.bars:
                all_prices.extend([bar.open, bar.high, bar.low, bar.close])
            if note.position == 'over':
                price = max(all_prices)
            else:  # under
                price = min(all_prices)
        
        self.ax.text(mdates.date2num(note.datetime), price, note.text, 
                    fontsize=font_size, color=font_color,
                    ha='center', va='bottom' if note.position == 'over' else 'top')
    
    def _render_indicators(self, indicators: List) -> None:
        """Render technical indicators."""
        if not indicators or not hasattr(self, 'chart') or not self.chart.bars:
            return
        
        import numpy as np
        import pandas as pd
        
        # Convert bars to DataFrame for easier calculation
        data = []
        for bar in self.chart.bars:
            data.append({
                'datetime': bar.datetime,
                'open': bar.open,
                'high': bar.high,
                'low': bar.low,
                'close': bar.close
            })
        
        df = pd.DataFrame(data)
        df.set_index('datetime', inplace=True)
        
        # Check if we need subplots for RSI/MACD
        has_rsi = any(ind.name == 'rsi' for ind in indicators)
        has_macd = any(ind.name == 'macd' for ind in indicators)
        needs_subplot = has_rsi or has_macd
        
        if needs_subplot:
            # Create subplots: main chart + indicator subplot
            self.fig.clear()
            if has_rsi and has_macd:
                # Two subplots: main chart, RSI, MACD
                gs = self.fig.add_gridspec(3, 1, height_ratios=[2, 1, 1], hspace=0.3)
                self.ax = self.fig.add_subplot(gs[0])  # Main chart
                rsi_ax = self.fig.add_subplot(gs[1])   # RSI
                macd_ax = self.fig.add_subplot(gs[2])  # MACD
            elif has_rsi:
                # Two subplots: main chart + RSI
                gs = self.fig.add_gridspec(2, 1, height_ratios=[2, 1], hspace=0.3)
                self.ax = self.fig.add_subplot(gs[0])  # Main chart
                rsi_ax = self.fig.add_subplot(gs[1])   # RSI
                macd_ax = None
            else:  # has_macd only
                # Two subplots: main chart + MACD
                gs = self.fig.add_gridspec(2, 1, height_ratios=[2, 1], hspace=0.3)
                self.ax = self.fig.add_subplot(gs[0])  # Main chart
                rsi_ax = None
                macd_ax = self.fig.add_subplot(gs[1])  # MACD
            
            # Re-render the main chart elements
            self._render_candlesticks(self.chart.bars)
            for drawing in self.chart.drawings:
                self._render_drawing(drawing)
            
            # Format subplot X-axes
            self._format_subplot_xaxis(rsi_ax, macd_ax)
        
        # Calculate and render each indicator
        for indicator in indicators:
            if indicator.name == 'ema':
                period = indicator.parameters.get('period', 20)
                if len(df) >= period:
                    ema_values = df['close'].ewm(span=period).mean()
                    self.ax.plot(mdates.date2num(ema_values.index), ema_values.values, 
                               label=f'EMA({period})', linewidth=2, alpha=0.8)
            
            elif indicator.name == 'sma':
                period = indicator.parameters.get('period', 20)
                if len(df) >= period:
                    sma_values = df['close'].rolling(window=period).mean()
                    self.ax.plot(mdates.date2num(sma_values.index), sma_values.values, 
                               label=f'SMA({period})', linewidth=2, alpha=0.8)
            
            elif indicator.name == 'bollinger':
                period = indicator.parameters.get('period', 20)
                stddev = indicator.parameters.get('stddev', 2)
                if len(df) >= period:
                    sma = df['close'].rolling(window=period).mean()
                    std = df['close'].rolling(window=period).std()
                    upper = sma + (std * stddev)
                    lower = sma - (std * stddev)
                    
                    # Plot bands
                    self.ax.plot(mdates.date2num(upper.index), upper.values, 
                               label=f'BB Upper({period})', linewidth=1, alpha=0.6, linestyle='--')
                    self.ax.plot(mdates.date2num(sma.index), sma.values, 
                               label=f'BB Middle({period})', linewidth=1, alpha=0.6)
                    self.ax.plot(mdates.date2num(lower.index), lower.values, 
                               label=f'BB Lower({period})', linewidth=1, alpha=0.6, linestyle='--')
            
            elif indicator.name == 'rsi' and rsi_ax is not None:
                period = indicator.parameters.get('period', 14)
                if len(df) >= period + 1:
                    delta = df['close'].diff()
                    gain = (delta.where(delta > 0, 0)).rolling(window=period).mean()
                    loss = (-delta.where(delta < 0, 0)).rolling(window=period).mean()
                    rs = gain / loss
                    rsi = 100 - (100 / (1 + rs))
                    
                    # Plot RSI on its own subplot
                    rsi_ax.plot(mdates.date2num(rsi.index), rsi.values, 
                               label=f'RSI({period})', linewidth=2, color='orange')
                    rsi_ax.axhline(y=70, color='r', linestyle='--', alpha=0.5, label='Overbought')
                    rsi_ax.axhline(y=30, color='g', linestyle='--', alpha=0.5, label='Oversold')
                    rsi_ax.set_ylabel('RSI')
                    rsi_ax.set_ylim(0, 100)
                    rsi_ax.legend(loc='upper right', fontsize=8)
                    rsi_ax.grid(True, alpha=0.3)
            
            elif indicator.name == 'macd' and macd_ax is not None:
                fast = indicator.parameters.get('fast', 12)
                slow = indicator.parameters.get('slow', 26)
                signal = indicator.parameters.get('signal', 9)
                
                if len(df) >= slow:
                    ema_fast = df['close'].ewm(span=fast).mean()
                    ema_slow = df['close'].ewm(span=slow).mean()
                    macd_line = ema_fast - ema_slow
                    signal_line = macd_line.ewm(span=signal).mean()
                    histogram = macd_line - signal_line
                    
                    # Plot MACD on its own subplot
                    macd_ax.plot(mdates.date2num(macd_line.index), macd_line.values, 
                               label=f'MACD({fast},{slow})', linewidth=2, color='purple')
                    macd_ax.plot(mdates.date2num(signal_line.index), signal_line.values, 
                               label=f'Signal({signal})', linewidth=2, color='magenta', linestyle='--')
                    macd_ax.bar(mdates.date2num(histogram.index), histogram.values, 
                              label='Histogram', alpha=0.3, color='gray')
                    macd_ax.axhline(y=0, color='black', linestyle='-', alpha=0.3)
                    macd_ax.set_ylabel('MACD')
                    macd_ax.legend(loc='upper right', fontsize=8)
                    macd_ax.grid(True, alpha=0.3)
        
        # Add legend to main chart
        self.ax.legend(loc='upper left', fontsize=8)
    
    def _format_subplot_xaxis(self, rsi_ax, macd_ax) -> None:
        """Format X-axis for subplots."""
        import matplotlib.dates as mdates
        
        if hasattr(self, 'bars') and self.bars:
            num_bars = len(self.bars)
            times = [mdates.date2num(bar.datetime) for bar in self.bars]
            time_range = max(times) - min(times)
            
            # Use same logic as main chart but with fewer ticks for subplots
            target_ticks = 4  # Fewer ticks for subplots
            
            # Format RSI subplot X-axis
            if rsi_ax is not None:
                if time_range <= 1:  # Less than 1 day
                    rsi_ax.xaxis.set_major_locator(mdates.HourLocator(interval=1))
                    rsi_ax.xaxis.set_major_formatter(mdates.DateFormatter('%H:%M'))
                elif time_range <= 7:  # Less than 1 week
                    rsi_ax.xaxis.set_major_locator(mdates.DayLocator(interval=1))
                    rsi_ax.xaxis.set_major_formatter(mdates.DateFormatter('%m/%d'))
                else:  # More than 1 week
                    rsi_ax.xaxis.set_major_locator(mdates.DayLocator(interval=2))
                    rsi_ax.xaxis.set_major_formatter(mdates.DateFormatter('%m/%d'))
                
                plt.setp(rsi_ax.xaxis.get_majorticklabels(), rotation=45, ha='right')
                rsi_ax.grid(True, alpha=0.3)
            
            # Format MACD subplot X-axis
            if macd_ax is not None:
                if time_range <= 1:  # Less than 1 day
                    macd_ax.xaxis.set_major_locator(mdates.HourLocator(interval=1))
                    macd_ax.xaxis.set_major_formatter(mdates.DateFormatter('%H:%M'))
                elif time_range <= 7:  # Less than 1 week
                    macd_ax.xaxis.set_major_locator(mdates.DayLocator(interval=1))
                    macd_ax.xaxis.set_major_formatter(mdates.DateFormatter('%m/%d'))
                else:  # More than 1 week
                    macd_ax.xaxis.set_major_locator(mdates.DayLocator(interval=2))
                    macd_ax.xaxis.set_major_formatter(mdates.DateFormatter('%m/%d'))
                
                plt.setp(macd_ax.xaxis.get_majorticklabels(), rotation=45, ha='right')
                macd_ax.grid(True, alpha=0.3)
    
    
    def _get_style_value(self, styles: Dict[str, Any], key: str, default: Any) -> Any:
        """Get a style value with default."""
        if not styles:
            return default
        
        value = styles.get(key, default)
        
        # Convert numeric values from strings
        if isinstance(value, str):
            if key in ["line-width", "font-size"]:
                try:
                    return int(value)
                except ValueError:
                    return default
            elif key in ["fill-opacity", "line-opacity"]:
                try:
                    return float(value)
                except ValueError:
                    return default
        
        return value
    
    def _get_meta_value(self, meta: List, key: str) -> Optional[str]:
        """Get a meta value by key."""
        for entry in meta:
            if entry.key == key:
                return str(entry.value)
        return None
    
    def _format_chart(self) -> None:
        """Format the chart appearance."""
        # Format x-axis for datetime
        import matplotlib.dates as mdates
        
        # Set tick frequency based on number of data points - MAX 8 ticks
        if hasattr(self, 'bars') and self.bars:
            num_bars = len(self.bars)
            
            # Special case for single bar - use a simple time range
            if num_bars == 1:
                single_time = mdates.date2num(self.bars[0].datetime)
                # Set a small range around the single point
                self.ax.xaxis.set_major_locator(mdates.HourLocator(interval=1))
                self.ax.xaxis.set_major_formatter(mdates.DateFormatter('%H:%M'))
                return
            
            times = [mdates.date2num(bar.datetime) for bar in self.bars]
            time_range = max(times) - min(times)
            
            # Calculate tick interval to get approximately 6-8 ticks
            target_ticks = 6
            tick_interval = max(1, num_bars // target_ticks)
            
            # Determine time unit and format based on data range
            if time_range <= 1:  # Less than 1 day
                if num_bars <= 20:
                    self.ax.xaxis.set_major_locator(mdates.MinuteLocator(interval=15))
                    self.ax.xaxis.set_major_formatter(mdates.DateFormatter('%H:%M'))
                else:
                    self.ax.xaxis.set_major_locator(mdates.HourLocator(interval=1))
                    self.ax.xaxis.set_major_formatter(mdates.DateFormatter('%H:%M'))
            elif time_range <= 7:  # Less than 1 week
                self.ax.xaxis.set_major_locator(mdates.DayLocator(interval=1))
                self.ax.xaxis.set_major_formatter(mdates.DateFormatter('%m/%d'))
            elif time_range <= 30:  # Less than 1 month
                self.ax.xaxis.set_major_locator(mdates.DayLocator(interval=2))
                self.ax.xaxis.set_major_formatter(mdates.DateFormatter('%m/%d'))
            else:  # More than 1 month
                self.ax.xaxis.set_major_locator(mdates.WeekdayLocator(interval=1))
                self.ax.xaxis.set_major_formatter(mdates.DateFormatter('%m/%d'))
            
            # Set minor ticks for better grid (but not too many)
            if time_range <= 1:
                self.ax.xaxis.set_minor_locator(mdates.MinuteLocator(interval=30))
            else:
                self.ax.xaxis.set_minor_locator(mdates.DayLocator(interval=1))
        
        # Rotate x-axis labels for better readability
        plt.xticks(rotation=45, ha='right')
        
        # Add grid
        self.ax.grid(True, alpha=0.3)
        self.ax.grid(True, alpha=0.1, which='minor')


def render_cml_file(input_file: str, output_file: Optional[str] = None) -> None:
    """Render a CML file to a chart."""
    from cml_parser import parse_cml_file
    
    chart = parse_cml_file(input_file)
    renderer = CMLRenderer()
    renderer.render(chart, output_file)


if __name__ == "__main__":
    import sys
    from cml_parser import CMLParser
    
    if len(sys.argv) < 2:
        print("Usage: python3 cml_renderer.py <input.cml> [output.png]")
        print("Example: python3 cml_renderer.py example.cml chart.png")
        sys.exit(1)
    
    input_file = sys.argv[1]
    output_file = sys.argv[2] if len(sys.argv) > 2 else "output.png"
    
    try:
        # Read the CML file
        with open(input_file, 'r') as f:
            cml_content = f.read()
        
        # Parse and render
        parser = CMLParser()
        chart = parser.parse(cml_content)
        
        renderer = CMLRenderer()
        renderer.render(chart, output_file)
        print(f"Chart rendered successfully to {output_file}!")
    except Exception as e:
        print(f"Rendering failed: {e}")
        sys.exit(1)
