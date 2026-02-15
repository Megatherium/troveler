// Package ui provides UI utilities for rendering.
package ui

// GradientColors is the predefined list of gradient colors.
var GradientColors = []string{
	"#90EE90", "#8BE88C", "#86E288", "#81DC88", "#7CD688",
	"#77D088", "#72CA88", "#6DC488", "#68BE88", "#63B888",
	"#5EB288", "#59AC88", "#54A688", "#4FA088", "#4A9A88",
	"#459488", "#408E88", "#3B8888", "#368288", "#317C88",
	"#2C7688", "#277088", "#226A88", "#1D6488", "#185E88",
	"#135888", "#0E5288", "#094C88", "#044688", "#004088",
}

// GetGradientColorSimple returns a gradient color by index (looping).
func GetGradientColorSimple(index int) string {
	return GradientColors[index%len(GradientColors)]
}

// GetGradientColor returns a gradient color based on position in a range.
func GetGradientColor(pos int, total int) string {
	if total == 0 {
		return GradientColors[0]
	}
	p := float64(pos) / float64(total)
	if p > 1 {
		p = 1
	}
	idx := int(p * float64(len(GradientColors)-1))

	return GradientColors[idx]
}
