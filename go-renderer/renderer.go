package main

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"golang.org/x/image/font/basicfont"
)

// CMLRenderer handles rendering of CML charts
type CMLRenderer struct {
	Width  int
	Height int
	dc     *gg.Context

	// Chart bounds
	minTime  time.Time
	maxTime  time.Time
	minPrice float64
	maxPrice float64

	// Margins
	marginLeft   float64
	marginRight  float64
	marginTop    float64
	marginBottom float64

	// Chart data
	bars  []Bar
	chart *Chart
}

// NewCMLRenderer creates a new CML renderer
func NewCMLRenderer(width, height int) *CMLRenderer {
	dc := gg.NewContext(width, height)
	dc.SetColor(color.White)
	dc.Clear()

	return &CMLRenderer{
		Width:  width,
		Height: height,
		dc:     dc,

		// Set default margins
		marginLeft:   60.0,
		marginRight:  20.0,
		marginTop:    40.0,
		marginBottom: 60.0,
	}
}

// Render renders a chart to a file
func (r *CMLRenderer) Render(chart *Chart, outputFile string) error {
	// Set up the chart
	r.setupChart(chart)

	// Render bars
	if len(chart.Bars) > 0 {
		r.renderBars(chart.Bars)
	}

	// Render drawings
	for _, drawing := range chart.Drawings {
		r.renderDrawing(drawing)
	}

	// Render indicators (placeholder)
	if len(chart.Indicators) > 0 {
		r.renderIndicators(chart.Indicators)
	}

	// Add title from meta
	title := r.getMetaValue(chart.Meta, "title")
	if title != "" {
		r.dc.SetColor(color.Black)
		r.dc.SetFontFace(basicfont.Face7x13)
		r.dc.DrawStringAnchored(title, float64(r.Width)/2, 20, 0.5, 0.5)
	}

	// Save the image
	return r.dc.SavePNG(outputFile)
}

// setupChart sets up the basic chart structure
func (r *CMLRenderer) setupChart(chart *Chart) {
	fmt.Printf("DEBUG: setupChart called with %d bars\n", len(chart.Bars))
	if len(chart.Bars) == 0 {
		return
	}

	// Store chart and bars for later use
	r.chart = chart
	r.bars = chart.Bars

	// Calculate time and price ranges
	r.minTime = chart.Bars[0].DateTime
	r.maxTime = chart.Bars[0].DateTime
	r.minPrice = chart.Bars[0].Low
	r.maxPrice = chart.Bars[0].High

	for _, bar := range chart.Bars {
		if bar.DateTime.Before(r.minTime) {
			r.minTime = bar.DateTime
		}
		if bar.DateTime.After(r.maxTime) {
			r.maxTime = bar.DateTime
		}
		if bar.Low < r.minPrice {
			r.minPrice = bar.Low
		}
		if bar.High > r.maxPrice {
			r.maxPrice = bar.High
		}
	}

	// Add some padding
	priceRange := r.maxPrice - r.minPrice
	if priceRange > 0 {
		r.minPrice -= priceRange * 0.05
		r.maxPrice += priceRange * 0.05
	} else {
		r.minPrice -= 1.0
		r.maxPrice += 1.0
	}

	// Add one extra interval on each side
	if len(chart.Bars) > 1 {
		interval := chart.Bars[1].DateTime.Sub(chart.Bars[0].DateTime)
		fmt.Printf("Interval: %v\n", interval)
		fmt.Printf("Before: %v to %v\n", r.minTime, r.maxTime)
		r.minTime = r.minTime.Add(-interval)
		r.maxTime = r.maxTime.Add(interval)
		fmt.Printf("After: %v to %v\n", r.minTime, r.maxTime)
	}

	// Draw chart background and axes
	r.dc.SetColor(color.Black)
	r.dc.SetLineWidth(1)

	// Chart area
	chartLeft := r.marginLeft
	chartRight := float64(r.Width) - r.marginRight
	chartTop := r.marginTop
	chartBottom := float64(r.Height) - r.marginBottom

	// Draw border
	r.dc.DrawRectangle(chartLeft, chartTop, chartRight-chartLeft, chartBottom-chartTop)
	r.dc.Stroke()

	// Draw grid lines (configurable)
	gridConfig := r.chart.GetGridConfig()
	if gridConfig.Enabled {
		gridColor := r.parseColor(gridConfig.Color)
		// Apply opacity and convert to NRGBA (premultiplied alpha)
		if rgba, ok := gridColor.(color.RGBA); ok {
			alpha := float64(rgba.A) / 255.0 * gridConfig.Opacity
			gridColorNRGBA := color.NRGBA{
				R: uint8(float64(rgba.R) * alpha),
				G: uint8(float64(rgba.G) * alpha),
				B: uint8(float64(rgba.B) * alpha),
				A: uint8(255 * gridConfig.Opacity),
			}
			r.dc.SetColor(gridColorNRGBA)
		} else {
			r.dc.SetColor(gridColor)
		}
		r.dc.SetLineWidth(gridConfig.LineWidth)

		// Horizontal grid lines (price levels)
		for i := 0; i <= 5; i++ {
			y := chartTop + (chartBottom-chartTop)*float64(i)/5.0
			r.dc.DrawLine(chartLeft, y, chartRight, y)
		}

		// Vertical grid lines (time levels) - match X-axis ticks exactly
		timeRange := r.maxTime.Sub(r.minTime)
		numBars := len(r.bars)

		// Calculate target number of ticks (max 8)
		targetTicks := 6
		if numBars < 10 {
			targetTicks = numBars
		}

		// Calculate interval to get approximately targetTicks
		interval := timeRange / time.Duration(targetTicks)

		// Round to nice intervals based on data frequency (same as X-axis labels)
		if timeRange <= 24*time.Hour {
			// Intraday data
			if interval <= 5*time.Minute {
				interval = 5 * time.Minute
			} else if interval <= 15*time.Minute {
				interval = 15 * time.Minute
			} else if interval <= 30*time.Minute {
				interval = 30 * time.Minute
			} else if interval <= 1*time.Hour {
				interval = 1 * time.Hour
			} else if interval <= 2*time.Hour {
				interval = 2 * time.Hour
			} else if interval <= 6*time.Hour {
				interval = 6 * time.Hour
			} else {
				interval = 12 * time.Hour
			}
		} else if timeRange <= 7*24*time.Hour {
			// Weekly data
			interval = 24 * time.Hour // Daily
		} else if timeRange <= 30*24*time.Hour {
			// Monthly data
			interval = 7 * 24 * time.Hour // Weekly
		} else if timeRange <= 90*24*time.Hour {
			// Quarterly data
			interval = 14 * 24 * time.Hour // Bi-weekly
		} else {
			// Longer periods
			interval = 30 * 24 * time.Hour // Monthly
		}

		// Find the first nice time that's >= minTime
		startTime := r.minTime.Truncate(interval)
		if startTime.Before(r.minTime) {
			startTime = startTime.Add(interval)
		}

		// Draw grid lines only at labeled tick positions (max 8)
		tickCount := 0
		for t := startTime; !t.After(r.maxTime) && tickCount < 8; t = t.Add(interval) {
			// Calculate X position
			timeOffset := t.Sub(r.minTime).Seconds()
			x := chartLeft + (chartRight-chartLeft)*(timeOffset/timeRange.Seconds())

			// Draw vertical grid line
			r.dc.DrawLine(x, chartTop, x, chartBottom)
			tickCount++
		}

		r.dc.Stroke()
	}

	// Draw axis labels
	r.drawAxisLabels()
}

// renderBars renders OHLC bars
func (r *CMLRenderer) renderBars(bars []Bar) {
	if len(bars) == 0 {
		return
	}

	// Calculate bar width
	chartLeft := r.marginLeft
	chartRight := float64(r.Width) - r.marginRight
	chartWidth := chartRight - chartLeft
	barWidth := chartWidth / float64(len(bars)) * 0.6

	for i, bar := range bars {
		// Calculate X position (center of bar) - not used directly since we use timePriceToScreen
		_ = chartLeft + (chartRight-chartLeft)*float64(i)/float64(len(bars)-1)

		// Convert prices to screen coordinates
		highX, highY := r.timePriceToScreen(bar.DateTime, bar.High)
		_, lowY := r.timePriceToScreen(bar.DateTime, bar.Low)
		openX, openY := r.timePriceToScreen(bar.DateTime, bar.Open)
		closeX, closeY := r.timePriceToScreen(bar.DateTime, bar.Close)

		// Draw upper wick (from high to body top)
		bodyTop := math.Min(openY, closeY)
		bodyBottom := math.Max(openY, closeY)

		r.dc.SetColor(color.Black)
		r.dc.SetLineWidth(1)

		// Draw upper wick (from high to body top)
		if highY < bodyTop {
			r.dc.DrawLine(highX, highY, highX, bodyTop)
			r.dc.Stroke()
		}

		// Draw lower wick (from low to body bottom)
		if lowY > bodyBottom {
			r.dc.DrawLine(highX, lowY, highX, bodyBottom)
			r.dc.Stroke()
		}

		// Draw open tick (left side)
		r.dc.DrawLine(openX-barWidth/4, openY, openX, openY)
		r.dc.Stroke()

		// Draw close tick (right side)
		r.dc.DrawLine(closeX, closeY, closeX+barWidth/4, closeY)
		r.dc.Stroke()

		// Draw open-close body
		bodyHeight := bodyBottom - bodyTop
		if bodyHeight < 1 {
			bodyHeight = 1 // Minimum height for visibility
		}

		// Choose color based on open vs close with configurable opacity
		barOpacityConfig := r.chart.GetBarOpacityConfig()
		opacity := uint8(255 * barOpacityConfig.Opacity)

		if bar.Close >= bar.Open {
			r.dc.SetColor(color.RGBA{0, 150, 0, opacity}) // Green
		} else {
			r.dc.SetColor(color.RGBA{200, 0, 0, opacity}) // Red
		}

		// Draw body rectangle
		r.dc.DrawRectangle(openX-barWidth/2, bodyTop, barWidth, bodyHeight)
		r.dc.Fill()

		// Draw body border
		r.dc.SetColor(color.Black)
		r.dc.SetLineWidth(1)
		r.dc.DrawRectangle(openX-barWidth/2, bodyTop, barWidth, bodyHeight)
		r.dc.Stroke()
	}
}

// renderDrawing renders a drawing element
func (r *CMLRenderer) renderDrawing(drawing Drawing) {
	switch d := drawing.(type) {
	case Rectangle:
		r.renderRectangle(d)
	case Line:
		r.renderLine(d)
	case ContinuousLine:
		r.renderContinuousLine(d)
	case Triangle:
		r.renderTriangle(d)
	case Circle:
		r.renderCircle(d)
	case Note:
		r.renderNote(d)
	}
}

// renderRectangle renders a rectangle
func (r *CMLRenderer) renderRectangle(rect Rectangle) {
	// Convert coordinates to screen space
	x1, y1 := r.timePriceToScreen(rect.StartTime, rect.StartPrice)
	x2, y2 := r.timePriceToScreen(rect.EndTime, rect.EndPrice)

	// Get styles
	borderColor := r.getStyleColor(rect.Styles, "border-color", color.RGBA{0, 0, 0, 255})
	fillColor := r.getStyleColor(rect.Styles, "fill-color", color.RGBA{170, 170, 170, 128})
	lineWidth := r.getStyleFloat(rect.Styles, "line-width", 1.0)
	fillOpacity := r.getStyleFloat(rect.Styles, "fill-opacity", 0.3)
	lineOpacity := r.getStyleFloat(rect.Styles, "line-opacity", 1.0)

	// Don't apply opacity here - will be handled in NRGBA conversion

	// Ensure proper rectangle dimensions (handle inverted Y coordinates)
	rectX := math.Min(x1, x2)
	rectY := math.Min(y1, y2)
	rectWidth := math.Abs(x2 - x1)
	rectHeight := math.Abs(y2 - y1)

	// Draw rectangle - convert RGBA to NRGBA for proper alpha blending
	// Convert RGBA to NRGBA (premultiplied alpha) with fill opacity
	if fillColorRGBA, ok := fillColor.(color.RGBA); ok {
		alpha := fillOpacity
		fillColorNRGBA := color.NRGBA{
			R: uint8(float64(fillColorRGBA.R) * alpha),
			G: uint8(float64(fillColorRGBA.G) * alpha),
			B: uint8(float64(fillColorRGBA.B) * alpha),
			A: uint8(255 * alpha),
		}
		_ = fillColorNRGBA // Keep this to maintain working behavior
		r.dc.SetColor(fillColorNRGBA)
	} else {
		fmt.Printf("DEBUG: Rectangle fill - not RGBA, using: %v\n", fillColor)
		r.dc.SetColor(fillColor)
	}

	r.dc.DrawRectangle(rectX, rectY, rectWidth, rectHeight)
	r.dc.Fill()

	// Draw border - convert RGBA to NRGBA with line opacity
	if borderColorRGBA, ok := borderColor.(color.RGBA); ok {
		alpha := lineOpacity
		borderColorNRGBA := color.NRGBA{
			R: uint8(float64(borderColorRGBA.R) * alpha),
			G: uint8(float64(borderColorRGBA.G) * alpha),
			B: uint8(float64(borderColorRGBA.B) * alpha),
			A: uint8(255 * alpha),
		}
		r.dc.SetColor(borderColorNRGBA)
	} else {
		fmt.Printf("DEBUG: Rectangle border - not RGBA, using: %v\n", borderColor)
		r.dc.SetColor(borderColor)
	}

	r.dc.SetLineWidth(lineWidth)
	r.dc.DrawRectangle(rectX, rectY, rectWidth, rectHeight)
	r.dc.Stroke()
}

// renderLine renders a line
func (r *CMLRenderer) renderLine(line Line) {
	// Convert coordinates to screen space
	x1, y1 := r.timePriceToScreen(line.StartTime, line.StartPrice)
	x2, y2 := r.timePriceToScreen(line.EndTime, line.EndPrice)

	// Get styles
	borderColor := r.getStyleColor(line.Styles, "border-color", color.RGBA{0, 0, 255, 255})
	lineWidth := r.getStyleFloat(line.Styles, "line-width", 2.0)
	lineOpacity := r.getStyleFloat(line.Styles, "line-opacity", 1.0)
	lineStyle := r.getStyleString(line.Styles, "style", "solid")

	// Apply opacity to border color
	if borderColorRGBA, ok := borderColor.(color.RGBA); ok {
		alpha := lineOpacity
		borderColorNRGBA := color.NRGBA{
			R: uint8(float64(borderColorRGBA.R) * alpha),
			G: uint8(float64(borderColorRGBA.G) * alpha),
			B: uint8(float64(borderColorRGBA.B) * alpha),
			A: uint8(255 * alpha),
		}
		r.dc.SetColor(borderColorNRGBA)
	} else {
		r.dc.SetColor(borderColor)
	}

	// Set line style
	r.dc.SetLineWidth(lineWidth)

	// Apply line style (dashed/dotted)
	switch lineStyle {
	case "dashed":
		r.dc.SetDash(lineWidth*2, lineWidth*2)
	case "dotted":
		r.dc.SetDash(lineWidth*0.5, lineWidth*2.5) // Small dots with even larger gaps
	default: // solid
		r.dc.SetDash() // Reset to solid
	}

	// Draw line
	r.dc.DrawLine(x1, y1, x2, y2)
	r.dc.Stroke()

	// Add arrow if specified
	if line.Arrow == "left-arrow" {
		r.drawArrow(x1, y1, x2, y2, borderColor, "left")
	} else if line.Arrow == "right-arrow" {
		r.drawArrow(x1, y1, x2, y2, borderColor, "right")
	} else if line.Arrow == "both-arrows" {
		r.drawArrow(x1, y1, x2, y2, borderColor, "left")
		r.drawArrow(x1, y1, x2, y2, borderColor, "right")
	}
}

// renderContinuousLine renders a continuous line
func (r *CMLRenderer) renderContinuousLine(line ContinuousLine) {
	// For continuous lines, extend to full chart width
	chartLeft := r.marginLeft
	chartRight := float64(r.Width) - r.marginRight

	// Convert Y coordinates (prices) to screen coordinates using dummy time
	_, y1 := r.timePriceToScreen(r.minTime, line.StartPrice)
	_, y2 := r.timePriceToScreen(r.minTime, line.EndPrice)

	// Use full chart width for X coordinates
	x1 := chartLeft
	x2 := chartRight

	// Get styles
	borderColor := r.getStyleColor(line.Styles, "border-color", color.RGBA{0, 128, 0, 255})
	lineWidth := r.getStyleFloat(line.Styles, "line-width", 1.0)
	lineOpacity := r.getStyleFloat(line.Styles, "line-opacity", 1.0)
	lineStyle := r.getStyleString(line.Styles, "style", "solid")

	// Apply opacity to border color
	if borderColorRGBA, ok := borderColor.(color.RGBA); ok {
		alpha := lineOpacity
		borderColorNRGBA := color.NRGBA{
			R: uint8(float64(borderColorRGBA.R) * alpha),
			G: uint8(float64(borderColorRGBA.G) * alpha),
			B: uint8(float64(borderColorRGBA.B) * alpha),
			A: uint8(255 * alpha),
		}
		r.dc.SetColor(borderColorNRGBA)
	} else {
		r.dc.SetColor(borderColor)
	}

	// Set line style
	r.dc.SetLineWidth(lineWidth)

	// Apply line style (dashed/dotted)
	switch lineStyle {
	case "dashed":
		r.dc.SetDash(lineWidth*2, lineWidth*2)
	case "dotted":
		r.dc.SetDash(lineWidth*0.5, lineWidth*2.5) // Small dots with even larger gaps
	default: // solid
		r.dc.SetDash() // Reset to solid
	}

	r.dc.DrawLine(x1, y1, x2, y2)
	r.dc.Stroke()
}

// renderTriangle renders a triangle marker
func (r *CMLRenderer) renderTriangle(triangle Triangle) {
	// Find the price at this time by looking at the bars
	var price float64
	found := false

	// Try to find the exact bar at this time
	for _, bar := range r.bars {
		if bar.DateTime.Equal(triangle.DateTime) {
			if triangle.Direction == "uptick" {
				price = bar.Low // Place uptick triangle below the price (at low)
			} else {
				price = bar.High // Place downtick triangle above the price (at high)
			}
			found = true
			break
		}
	}

	// If not found, use a reasonable default
	if !found {
		if triangle.Direction == "uptick" {
			price = r.minPrice + (r.maxPrice-r.minPrice)*0.1 // Near the bottom
		} else {
			price = r.maxPrice - (r.maxPrice-r.minPrice)*0.1 // Near the top
		}
	}

	x, y := r.timePriceToScreen(triangle.DateTime, price)

	borderColor := r.getStyleColor(triangle.Styles, "border-color", color.RGBA{0, 0, 0, 255})
	fillColor := r.getStyleColor(triangle.Styles, "fill-color", color.RGBA{170, 170, 170, 255})

	// Draw triangle
	size := 8.0
	if triangle.Direction == "uptick" {
		// Upward triangle - positioned below the price
		r.dc.SetColor(fillColor)
		r.dc.DrawRegularPolygon(3, x, y+size, size, 0)
		r.dc.Fill()
		r.dc.SetColor(borderColor)
		r.dc.DrawRegularPolygon(3, x, y+size, size, 0)
		r.dc.Stroke()
	} else {
		// Downward triangle - positioned above the price
		r.dc.SetColor(fillColor)
		r.dc.DrawRegularPolygon(3, x, y-size, size, math.Pi)
		r.dc.Fill()
		r.dc.SetColor(borderColor)
		r.dc.DrawRegularPolygon(3, x, y-size, size, math.Pi)
		r.dc.Stroke()
	}

	_ = found // Suppress unused variable warning
}

// renderCircle renders a circle marker
func (r *CMLRenderer) renderCircle(circle Circle) {
	// Find the price at this time by looking at the bars
	var price float64
	found := false

	// Try to find the exact bar at this time
	for _, bar := range r.bars {
		if bar.DateTime.Equal(circle.DateTime) {
			price = (bar.High + bar.Low) / 2 // Use middle of the bar
			found = true
			break
		}
	}

	// If not found, use a reasonable default
	if !found {
		price = r.minPrice + (r.maxPrice-r.minPrice)*0.5 // Middle of price range
	}

	x, y := r.timePriceToScreen(circle.DateTime, price)

	borderColor := r.getStyleColor(circle.Styles, "border-color", color.RGBA{0, 0, 0, 255})
	fillColor := r.getStyleColor(circle.Styles, "fill-color", color.RGBA{255, 255, 0, 255})
	lineWidth := r.getStyleFloat(circle.Styles, "line-width", 1.0)

	radius := 6.0

	// Draw circle
	r.dc.SetColor(fillColor)
	r.dc.DrawCircle(x, y, radius)
	r.dc.Fill()

	r.dc.SetColor(borderColor)
	r.dc.SetLineWidth(lineWidth)
	r.dc.DrawCircle(x, y, radius)
	r.dc.Stroke()
}

// renderNote renders a text note
func (r *CMLRenderer) renderNote(note Note) {
	// Find the price at this time by looking at the bars
	var price float64
	found := false

	// Try to find the exact bar at this time
	for _, bar := range r.bars {
		if bar.DateTime.Equal(note.DateTime) {
			if note.Position == "over" {
				price = bar.High // Place over note at the high
			} else {
				price = bar.Low // Place under note at the low
			}
			found = true
			break
		}
	}

	// If not found, use a reasonable default
	if !found {
		if note.Position == "over" {
			price = r.maxPrice - (r.maxPrice-r.minPrice)*0.1 // Near the top
		} else {
			price = r.minPrice + (r.maxPrice-r.minPrice)*0.1 // Near the bottom
		}
	}

	x, y := r.timePriceToScreen(note.DateTime, price)

	fontSize := r.getStyleFloat(note.Styles, "font-size", 12.0)
	fontColor := r.getStyleColor(note.Styles, "font-color", color.RGBA{0, 0, 0, 255})

	// Set font
	r.dc.SetColor(fontColor)
	r.dc.SetFontFace(basicfont.Face7x13)

	// Draw text with proper positioning
	offset := 15.0
	if note.Position == "over" {
		r.dc.DrawStringAnchored(note.Text, x, y-offset, 0.5, 1.0)
	} else {
		r.dc.DrawStringAnchored(note.Text, x, y+offset, 0.5, 0.0)
	}

	_ = fontSize // Suppress unused variable warning
}

// drawAxisLabels draws price labels on Y-axis and datetime labels on X-axis
func (r *CMLRenderer) drawAxisLabels() {
	// Set font for labels
	r.dc.SetColor(color.Black)
	r.dc.SetFontFace(basicfont.Face7x13)

	// Chart area
	chartLeft := r.marginLeft
	chartRight := float64(r.Width) - r.marginRight
	chartTop := r.marginTop
	chartBottom := float64(r.Height) - r.marginBottom

	// Draw Y-axis price labels
	priceRange := r.maxPrice - r.minPrice
	yAxisConfig := r.chart.GetYAxisConfig()
	for i := 0; i <= 5; i++ {
		// Calculate price for this grid line
		price := r.minPrice + (priceRange * float64(i) / 5.0)

		// Calculate Y position
		y := chartBottom - (chartBottom-chartTop)*float64(i)/5.0

		// Format price with configurable precision
		formatStr := fmt.Sprintf("%%.%df", yAxisConfig.Precision)
		priceText := fmt.Sprintf(formatStr, price)

		// Draw price label to the left of the chart
		r.dc.DrawStringAnchored(priceText, chartLeft-10, y, 1.0, 0.5)
	}

	// Draw X-axis datetime labels with dynamic scaling
	timeRange := r.maxTime.Sub(r.minTime)
	numBars := len(r.bars)

	// Calculate target number of ticks (max 8)
	targetTicks := 6
	if numBars < 10 {
		targetTicks = numBars
	}

	// Calculate interval to get approximately targetTicks
	interval := timeRange / time.Duration(targetTicks)

	// Round to nice intervals based on data frequency
	if timeRange <= 24*time.Hour {
		// Intraday data
		if interval <= 5*time.Minute {
			interval = 5 * time.Minute
		} else if interval <= 15*time.Minute {
			interval = 15 * time.Minute
		} else if interval <= 30*time.Minute {
			interval = 30 * time.Minute
		} else if interval <= 1*time.Hour {
			interval = 1 * time.Hour
		} else if interval <= 2*time.Hour {
			interval = 2 * time.Hour
		} else if interval <= 6*time.Hour {
			interval = 6 * time.Hour
		} else {
			interval = 12 * time.Hour
		}
	} else if timeRange <= 7*24*time.Hour {
		// Weekly data
		interval = 24 * time.Hour // Daily
	} else if timeRange <= 30*24*time.Hour {
		// Monthly data
		interval = 7 * 24 * time.Hour // Weekly
	} else if timeRange <= 90*24*time.Hour {
		// Quarterly data
		interval = 14 * 24 * time.Hour // Bi-weekly
	} else {
		// Longer periods
		interval = 30 * 24 * time.Hour // Monthly
	}

	// Find the first nice time that's >= minTime
	startTime := r.minTime.Truncate(interval)
	if startTime.Before(r.minTime) {
		startTime = startTime.Add(interval)
	}

	// Draw labels at nice intervals
	tickCount := 0
	for t := startTime; !t.After(r.maxTime) && tickCount < 8; t = t.Add(interval) {
		// Calculate X position
		timeOffset := t.Sub(r.minTime).Seconds()
		x := chartLeft + (chartRight-chartLeft)*(timeOffset/timeRange.Seconds())

		// Format time based on range
		var timeText string
		if timeRange <= 24*time.Hour {
			timeText = t.Format("15:04")
		} else if timeRange <= 7*24*time.Hour {
			timeText = t.Format("01/02")
		} else {
			timeText = t.Format("01/02")
		}

		// Draw time label below the chart
		r.dc.DrawStringAnchored(timeText, x, chartBottom+20, 0.5, 0.0)
		tickCount++
	}
}

// renderIndicators renders technical indicators
func (r *CMLRenderer) renderIndicators(indicators []Indicator) {
	if len(indicators) == 0 || len(r.bars) == 0 {
		return
	}

	// Calculate and render each indicator (only price-scale indicators for Go)
	for _, indicator := range indicators {
		switch indicator.Name {
		case "ema":
			if period, ok := indicator.Parameters["period"].(float64); ok {
				r.renderEMA(int(period))
			}
		case "sma":
			if period, ok := indicator.Parameters["period"].(float64); ok {
				r.renderSMA(int(period))
			}
		case "bollinger":
			if period, ok := indicator.Parameters["period"].(float64); ok {
				if stddev, ok := indicator.Parameters["stddev"].(float64); ok {
					r.renderBollingerBands(int(period), stddev)
				}
			}
		case "rsi":
			// Skip RSI - requires separate subplot for proper scaling
			continue
		case "macd":
			// Skip MACD - requires separate subplot for proper scaling
			continue
		}
	}
}

// renderEMA renders Exponential Moving Average
func (r *CMLRenderer) renderEMA(period int) {
	if len(r.bars) < period {
		return
	}

	// Calculate EMA
	alpha := 2.0 / float64(period+1)
	ema := make([]float64, len(r.bars))
	ema[0] = r.bars[0].Close

	for i := 1; i < len(r.bars); i++ {
		ema[i] = alpha*r.bars[i].Close + (1-alpha)*ema[i-1]
	}

	// Draw EMA line
	r.dc.SetColor(color.RGBA{255, 0, 0, 200}) // Red
	r.dc.SetLineWidth(2)

	for i := 1; i < len(ema); i++ {
		x1, y1 := r.timePriceToScreen(r.bars[i-1].DateTime, ema[i-1])
		x2, y2 := r.timePriceToScreen(r.bars[i].DateTime, ema[i])
		r.dc.DrawLine(x1, y1, x2, y2)
	}
	r.dc.Stroke()
}

// renderSMA renders Simple Moving Average
func (r *CMLRenderer) renderSMA(period int) {
	if len(r.bars) < period {
		return
	}

	// Calculate SMA
	sma := make([]float64, len(r.bars))
	for i := period - 1; i < len(r.bars); i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += r.bars[j].Close
		}
		sma[i] = sum / float64(period)
	}

	// Draw SMA line
	r.dc.SetColor(color.RGBA{0, 255, 0, 200}) // Green
	r.dc.SetLineWidth(2)

	for i := period; i < len(sma); i++ {
		x1, y1 := r.timePriceToScreen(r.bars[i-1].DateTime, sma[i-1])
		x2, y2 := r.timePriceToScreen(r.bars[i].DateTime, sma[i])
		r.dc.DrawLine(x1, y1, x2, y2)
	}
	r.dc.Stroke()
}

// renderBollingerBands renders Bollinger Bands
func (r *CMLRenderer) renderBollingerBands(period int, stddev float64) {
	if len(r.bars) < period {
		return
	}

	// Calculate SMA and standard deviation
	sma := make([]float64, len(r.bars))
	std := make([]float64, len(r.bars))

	for i := period - 1; i < len(r.bars); i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += r.bars[j].Close
		}
		sma[i] = sum / float64(period)

		// Calculate standard deviation
		variance := 0.0
		for j := i - period + 1; j <= i; j++ {
			variance += (r.bars[j].Close - sma[i]) * (r.bars[j].Close - sma[i])
		}
		std[i] = math.Sqrt(variance / float64(period))
	}

	// Draw bands
	r.dc.SetColor(color.RGBA{0, 0, 255, 150}) // Blue
	r.dc.SetLineWidth(1)

	// Upper band
	for i := period; i < len(sma); i++ {
		upper := sma[i] + std[i]*stddev
		x1, y1 := r.timePriceToScreen(r.bars[i-1].DateTime, sma[i-1]+std[i-1]*stddev)
		x2, y2 := r.timePriceToScreen(r.bars[i].DateTime, upper)
		r.dc.DrawLine(x1, y1, x2, y2)
	}
	r.dc.Stroke()

	// Middle band (SMA)
	for i := period; i < len(sma); i++ {
		x1, y1 := r.timePriceToScreen(r.bars[i-1].DateTime, sma[i-1])
		x2, y2 := r.timePriceToScreen(r.bars[i].DateTime, sma[i])
		r.dc.DrawLine(x1, y1, x2, y2)
	}
	r.dc.Stroke()

	// Lower band
	for i := period; i < len(sma); i++ {
		lower := sma[i] - std[i]*stddev
		x1, y1 := r.timePriceToScreen(r.bars[i-1].DateTime, sma[i-1]-std[i-1]*stddev)
		x2, y2 := r.timePriceToScreen(r.bars[i].DateTime, lower)
		r.dc.DrawLine(x1, y1, x2, y2)
	}
	r.dc.Stroke()
}

// renderRSI renders Relative Strength Index
func (r *CMLRenderer) renderRSI(period int) {
	if len(r.bars) < period+1 {
		return
	}

	// Calculate RSI
	gains := make([]float64, len(r.bars))
	losses := make([]float64, len(r.bars))

	for i := 1; i < len(r.bars); i++ {
		change := r.bars[i].Close - r.bars[i-1].Close
		if change > 0 {
			gains[i] = change
		} else {
			losses[i] = -change
		}
	}

	// Calculate average gains and losses
	avgGain := 0.0
	avgLoss := 0.0
	for i := 1; i <= period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	rsi := make([]float64, len(r.bars))
	for i := period; i < len(r.bars); i++ {
		if i > period {
			avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
			avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)
		}

		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}

	// Scale RSI to price range for visibility
	priceRange := r.maxPrice - r.minPrice
	r.dc.SetColor(color.RGBA{255, 165, 0, 200}) // Orange
	r.dc.SetLineWidth(2)

	for i := period + 1; i < len(rsi); i++ {
		// Scale RSI (0-100) to price range
		scaledRSI := r.minPrice + (rsi[i]/100)*priceRange
		x1, y1 := r.timePriceToScreen(r.bars[i-1].DateTime, r.minPrice+(rsi[i-1]/100)*priceRange)
		x2, y2 := r.timePriceToScreen(r.bars[i].DateTime, scaledRSI)
		r.dc.DrawLine(x1, y1, x2, y2)
	}
	r.dc.Stroke()
}

// renderMACD renders MACD indicator
func (r *CMLRenderer) renderMACD(fast, slow, signal int) {
	if len(r.bars) < slow {
		return
	}

	// Calculate EMAs
	fastAlpha := 2.0 / float64(fast+1)
	slowAlpha := 2.0 / float64(slow+1)

	emaFast := make([]float64, len(r.bars))
	emaSlow := make([]float64, len(r.bars))

	emaFast[0] = r.bars[0].Close
	emaSlow[0] = r.bars[0].Close

	for i := 1; i < len(r.bars); i++ {
		emaFast[i] = fastAlpha*r.bars[i].Close + (1-fastAlpha)*emaFast[i-1]
		emaSlow[i] = slowAlpha*r.bars[i].Close + (1-slowAlpha)*emaSlow[i-1]
	}

	// Calculate MACD line
	macd := make([]float64, len(r.bars))
	for i := 0; i < len(r.bars); i++ {
		macd[i] = emaFast[i] - emaSlow[i]
	}

	// Calculate signal line
	signalAlpha := 2.0 / float64(signal+1)
	signalLine := make([]float64, len(r.bars))
	signalLine[0] = macd[0]

	for i := 1; i < len(r.bars); i++ {
		signalLine[i] = signalAlpha*macd[i] + (1-signalAlpha)*signalLine[i-1]
	}

	// Scale MACD to price range for visibility
	priceRange := r.maxPrice - r.minPrice
	macdRange := 0.0
	for i := slow; i < len(macd); i++ {
		if math.Abs(macd[i]) > macdRange {
			macdRange = math.Abs(macd[i])
		}
	}

	// Draw MACD line
	r.dc.SetColor(color.RGBA{128, 0, 128, 200}) // Purple
	r.dc.SetLineWidth(2)

	for i := slow + 1; i < len(macd); i++ {
		scaledMACD1 := r.minPrice + (macd[i-1]/macdRange)*priceRange*0.1
		scaledMACD2 := r.minPrice + (macd[i]/macdRange)*priceRange*0.1
		x1, y1 := r.timePriceToScreen(r.bars[i-1].DateTime, scaledMACD1)
		x2, y2 := r.timePriceToScreen(r.bars[i].DateTime, scaledMACD2)
		r.dc.DrawLine(x1, y1, x2, y2)
	}
	r.dc.Stroke()

	// Draw signal line
	r.dc.SetColor(color.RGBA{255, 0, 255, 200}) // Magenta
	r.dc.SetLineWidth(2)

	for i := slow + 1; i < len(signalLine); i++ {
		scaledSignal1 := r.minPrice + (signalLine[i-1]/macdRange)*priceRange*0.1
		scaledSignal2 := r.minPrice + (signalLine[i]/macdRange)*priceRange*0.1
		x1, y1 := r.timePriceToScreen(r.bars[i-1].DateTime, scaledSignal1)
		x2, y2 := r.timePriceToScreen(r.bars[i].DateTime, scaledSignal2)
		r.dc.DrawLine(x1, y1, x2, y2)
	}
	r.dc.Stroke()
}

// Helper methods

// timePriceToScreen converts time and price to screen coordinates
func (r *CMLRenderer) timePriceToScreen(t time.Time, price float64) (float64, float64) {
	// Calculate chart area
	chartLeft := r.marginLeft
	chartRight := float64(r.Width) - r.marginRight
	chartTop := r.marginTop
	chartBottom := float64(r.Height) - r.marginBottom

	// Convert time to X coordinate
	timeRange := r.maxTime.Sub(r.minTime).Seconds()
	var x float64
	if timeRange > 0 {
		timeOffset := t.Sub(r.minTime).Seconds()
		x = chartLeft + (chartRight-chartLeft)*(timeOffset/timeRange)
	} else {
		x = chartLeft + (chartRight-chartLeft)/2
	}

	// Convert price to Y coordinate (inverted - higher prices at top)
	priceRange := r.maxPrice - r.minPrice
	var y float64
	if priceRange > 0 {
		priceOffset := price - r.minPrice
		y = chartBottom - (chartBottom-chartTop)*(priceOffset/priceRange)
	} else {
		y = chartTop + (chartBottom-chartTop)/2
	}

	return x, y
}

// drawArrow draws an arrow at the specified end of a line
func (r *CMLRenderer) drawArrow(x1, y1, x2, y2 float64, color color.Color, direction string) {
	// Calculate arrow direction
	dx := x2 - x1
	dy := y2 - y1
	length := math.Sqrt(dx*dx + dy*dy)

	if length == 0 {
		return
	}

	// Normalize direction
	dx /= length
	dy /= length

	// Arrow size
	arrowSize := 10.0
	arrowAngle := math.Pi / 6 // 30 degrees

	var arrowX1, arrowY1, arrowX2, arrowY2 float64
	var arrowX, arrowY float64

	// Determine which end of the line to draw the arrow
	if direction == "left" {
		arrowX, arrowY = x1, y1
		// Reverse direction for left arrow
		dx = -dx
		dy = -dy
	} else { // right arrow
		arrowX, arrowY = x2, y2
	}

	// Calculate arrow points
	arrowX1 = arrowX - arrowSize*math.Cos(math.Atan2(dy, dx)-arrowAngle)
	arrowY1 = arrowY - arrowSize*math.Sin(math.Atan2(dy, dx)-arrowAngle)
	arrowX2 = arrowX - arrowSize*math.Cos(math.Atan2(dy, dx)+arrowAngle)
	arrowY2 = arrowY - arrowSize*math.Sin(math.Atan2(dy, dx)+arrowAngle)

	// Draw arrow
	r.dc.SetColor(color)
	r.dc.SetLineWidth(2)
	r.dc.DrawLine(arrowX, arrowY, arrowX1, arrowY1)
	r.dc.DrawLine(arrowX, arrowY, arrowX2, arrowY2)
	r.dc.Stroke()
}

// getStyleColor gets a color from styles with default
func (r *CMLRenderer) getStyleColor(styles map[string]interface{}, key string, defaultColor color.Color) color.Color {
	if styles == nil {
		return defaultColor
	}

	if val, ok := styles[key]; ok {
		if colorStr, ok := val.(string); ok {
			return r.parseColor(colorStr)
		}
	}

	return defaultColor
}

// getStyleFloat gets a float from styles with default
func (r *CMLRenderer) getStyleFloat(styles map[string]interface{}, key string, defaultValue float64) float64 {
	if styles == nil {
		return defaultValue
	}

	if val, ok := styles[key]; ok {
		if floatVal, ok := val.(float64); ok {
			return floatVal
		}
	}

	return defaultValue
}

// getStyleString gets a string from styles with default
func (r *CMLRenderer) getStyleString(styles map[string]interface{}, key string, defaultValue string) string {
	if styles == nil {
		return defaultValue
	}

	if val, ok := styles[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// getMetaValue gets a meta value by key
func (r *CMLRenderer) getMetaValue(meta []MetaEntry, key string) string {
	for _, entry := range meta {
		if entry.Key == key {
			if str, ok := entry.Value.(string); ok {
				return str
			}
		}
	}
	return ""
}

// parseColor parses a hex color string
func (r *CMLRenderer) parseColor(colorStr string) color.Color {
	// Remove # if present
	colorStr = strings.TrimPrefix(colorStr, "#")

	// Parse hex color
	var red, green, blue uint8

	if len(colorStr) == 3 {
		// Short format (RGB)
		redVal, err := strconv.ParseUint(colorStr[0:1]+colorStr[0:1], 16, 8)
		if err != nil {
			return color.RGBA{0, 0, 0, 255}
		}
		greenVal, err := strconv.ParseUint(colorStr[1:2]+colorStr[1:2], 16, 8)
		if err != nil {
			return color.RGBA{0, 0, 0, 255}
		}
		blueVal, err := strconv.ParseUint(colorStr[2:3]+colorStr[2:3], 16, 8)
		if err != nil {
			return color.RGBA{0, 0, 0, 255}
		}
		red, green, blue = uint8(redVal), uint8(greenVal), uint8(blueVal)
	} else if len(colorStr) == 6 {
		// Long format (RRGGBB)
		redVal, err := strconv.ParseUint(colorStr[0:2], 16, 8)
		if err != nil {
			return color.RGBA{0, 0, 0, 255}
		}
		greenVal, err := strconv.ParseUint(colorStr[2:4], 16, 8)
		if err != nil {
			return color.RGBA{0, 0, 0, 255}
		}
		blueVal, err := strconv.ParseUint(colorStr[4:6], 16, 8)
		if err != nil {
			return color.RGBA{0, 0, 0, 255}
		}
		red, green, blue = uint8(redVal), uint8(greenVal), uint8(blueVal)
	} else {
		return color.RGBA{0, 0, 0, 255}
	}

	return color.RGBA{red, green, blue, 255}
}
