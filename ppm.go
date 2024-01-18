package Netpbm

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
)

type PPM struct {
	data          [][]Pixel
	width, height int
	magicNumber   string
	max           int
}

type Pixel struct {
	R, G, B uint8
}

// ReadPPM reads a PPM image from a file and returns a struct that represents the image.
func ReadPPM(filename string) (*PPM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// Read magic number
	magicNumber, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading magic number: %v", err)
	}
	magicNumber = strings.TrimSpace(magicNumber)
	if magicNumber != "P3" && magicNumber != "P6" {
		return nil, fmt.Errorf("invalid magic number: %s", magicNumber)
	}

	// Read dimensions
	dimensions, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading dimensions: %v", err)
	}
	var width, height int
	_, err = fmt.Sscanf(strings.TrimSpace(dimensions), "%d %d", &width, &height)
	if err != nil {
		return nil, fmt.Errorf("invalid dimensions: %v", err)
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: width and height must be positive")
	}

	// Read max value
	maxValue, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading max value: %v", err)
	}
	maxValue = strings.TrimSpace(maxValue)
	var max int
	_, err = fmt.Sscanf(maxValue, "%d", &max)
	if err != nil {
		return nil, fmt.Errorf("invalid max value: %v", err)
	}

	// Read image data
	data := make([][]Pixel, height)
	expectedBytesPerPixel := 3

	if magicNumber == "P3" {
		// Read P3 format (ASCII)
		for y := 0; y < height; y++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("error reading data at row %d: %v", y, err)
			}
			fields := strings.Fields(line)
			rowData := make([]Pixel, width)
			for x := 0; x < width; x++ {
				if x*3+2 >= len(fields) {
					return nil, fmt.Errorf("index out of range at row %d, column %d", y, x)
				}
				var pixel Pixel
				_, err := fmt.Sscanf(fields[x*3], "%d", &pixel.R)
				if err != nil {
					return nil, fmt.Errorf("error parsing Red value at row %d, column %d: %v", y, x, err)
				}
				_, err = fmt.Sscanf(fields[x*3+1], "%d", &pixel.G)
				if err != nil {
					return nil, fmt.Errorf("error parsing Green value at row %d, column %d: %v", y, x, err)
				}
				_, err = fmt.Sscanf(fields[x*3+2], "%d", &pixel.B)
				if err != nil {
					return nil, fmt.Errorf("error parsing Blue value at row %d, column %d: %v", y, x, err)
				}
				rowData[x] = pixel
			}
			data[y] = rowData
		}
	} else if magicNumber == "P6" {
		// Read P6 format (binary)
		for y := 0; y < height; y++ {
			row := make([]byte, width*expectedBytesPerPixel)
			n, err := reader.Read(row)
			if err != nil {
				if err == io.EOF {
					return nil, fmt.Errorf("unexpected end of file at row %d", y)
				}
				return nil, fmt.Errorf("error reading pixel data at row %d: %v", y, err)
			}
			if n < width*expectedBytesPerPixel {
				return nil, fmt.Errorf("unexpected end of file at row %d, expected %d bytes, got %d", y, width*expectedBytesPerPixel, n)
			}

			rowData := make([]Pixel, width)
			for x := 0; x < width; x++ {
				pixel := Pixel{R: row[x*expectedBytesPerPixel], G: row[x*expectedBytesPerPixel+1], B: row[x*expectedBytesPerPixel+2]}
				rowData[x] = pixel
			}
			data[y] = rowData
		}
	}

	// Return the PPM struct
	return &PPM{data, width, height, magicNumber, max}, nil
}

func (ppm *PPM) PrintPPM() {
	fmt.Printf("Magic Number: %s\n", ppm.magicNumber)
	fmt.Printf("Width: %d\n", ppm.width)
	fmt.Printf("Height: %d\n", ppm.height)
	fmt.Printf("Max Value: %d\n", ppm.max)

	fmt.Println("Pixel Data:")
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			pixel := ppm.data[y][x]
			fmt.Printf("(%d, %d, %d) ", pixel.R, pixel.G, pixel.B)
		}
		fmt.Println()
	}
}

func (ppm *PPM) Size() (int, int) {
	return ppm.width, ppm.height
}

func (ppm *PPM) At(x, y int) Pixel {
	// Vérification des limites pour éviter les erreurs d'index
	if x < 0 || x >= ppm.width || y < 0 || y >= ppm.height {
		// Vous pouvez également gérer cela différemment, comme renvoyer une valeur par défaut ou une erreur.
		panic("Index out of bounds")
	}

	return ppm.data[y][x]
}

func (ppm *PPM) Set(x, y int, value Pixel) {
	// Vérification des limites pour éviter les erreurs d'index
	if x < 0 || x >= ppm.width || y < 0 || y >= ppm.height {
		// Vous pouvez également gérer cela différemment, comme renvoyer une valeur par défaut ou une erreur.
		panic("Index out of bounds")
	}

	ppm.data[y][x] = value
}

func (ppm *PPM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	if ppm.magicNumber == "P6" || ppm.magicNumber == "P3" {
		fmt.Fprintf(file, "%s\n%d %d\n%d\n", ppm.magicNumber, ppm.width, ppm.height, ppm.max)
	} else {
		err = fmt.Errorf("magic number error")
		return err
	}

	//bytesPerPixel := 3 // Nombre d'octets par pixel pour P6

	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			pixel := ppm.data[y][x]
			if ppm.magicNumber == "P6" {
				// Conversion inverse des pixels
				file.Write([]byte{pixel.R, pixel.G, pixel.B})
			} else if ppm.magicNumber == "P3" {
				// Conversion inverse des pixels
				fmt.Fprintf(file, "%d %d %d ", pixel.R, pixel.G, pixel.B)
			}
		}
		if ppm.magicNumber == "P3" {
			fmt.Fprint(file, "\n")
		}
	}

	return nil
}

func (ppm *PPM) Invert() {
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			pixel := &ppm.data[y][x]
			pixel.R = 255 - pixel.R
			pixel.G = 255 - pixel.G
			pixel.B = 255 - pixel.B
		}
	}
}

func (ppm *PPM) Flip() {
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width/2; x++ {
			ppm.data[y][x], ppm.data[y][ppm.width-x-1] = ppm.data[y][ppm.width-x-1], ppm.data[y][x]
		}
	}
}

func (ppm *PPM) Flop() {
	for y := 0; y < ppm.height/2; y++ {
		ppm.data[y], ppm.data[ppm.height-y-1] = ppm.data[ppm.height-y-1], ppm.data[y]
	}
}

func (ppm *PPM) SetMagicNumber(magicNumber string) {
	ppm.magicNumber = magicNumber
}

// SetMaxValue updates the maximum pixel value in the PPM structure
// and scales the pixel values in data based on the new max value.
func (ppm *PPM) SetMaxValue(maxValue uint8) {
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			// Scale the RGB values based on the new max value
			ppm.data[y][x].R = uint8(float64(ppm.data[y][x].R) * float64(maxValue) / float64(ppm.max))
			ppm.data[y][x].G = uint8(float64(ppm.data[y][x].G) * float64(maxValue) / float64(ppm.max))
			ppm.data[y][x].B = uint8(float64(ppm.data[y][x].B) * float64(maxValue) / float64(ppm.max))
		}
	}

	// Update the max value
	ppm.max = int(maxValue)
}

func (ppm *PPM) Rotate90CW() {
	newPPM := PPM{
		data:        make([][]Pixel, ppm.width),
		width:       ppm.height,
		height:      ppm.width,
		magicNumber: ppm.magicNumber,
		max:         ppm.max,
	}

	for i := range newPPM.data {
		newPPM.data[i] = make([]Pixel, newPPM.width)
	}

	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			newPPM.data[x][ppm.height-y-1] = ppm.data[y][x]
		}
	}

	*ppm = newPPM
}

// ToPGM converts the PPM image to a PGM image (grayscale).
func (ppm *PPM) ToPGM() *PGM {
	pgm := &PGM{
		width:       ppm.width,
		height:      ppm.height,
		magicNumber: "P2",
		max:         ppm.max,
	}

	pgm.data = make([][]uint8, ppm.height)
	for i := range pgm.data {
		pgm.data[i] = make([]uint8, ppm.width)
	}

	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			// Convert RGB to grayscale
			gray := uint8((int(ppm.data[y][x].R) + int(ppm.data[y][x].G) + int(ppm.data[y][x].B)) / 3)
			pgm.data[y][x] = gray
		}
	}

	return pgm
}

type Point struct {
	X, Y int
}

func rgbToGray(color Pixel) uint8 {

	return uint8(0.299*float64(color.R) + 0.587*float64(color.G) + 0.114*float64(color.B))
}

func (ppm *PPM) ToPBM() *PBM {
	pbm := &PBM{
		width:       ppm.width,
		height:      ppm.height,
		magicNumber: "P1",
	}

	pbm.data = make([][]bool, ppm.height)
	for i := range pbm.data {
		pbm.data[i] = make([]bool, ppm.width)
	}

	threshold := uint8(ppm.max / 2)

	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {

			average := (uint16(ppm.data[y][x].R) + uint16(ppm.data[y][x].G) + uint16(ppm.data[y][x].B)) / 3

			pbm.data[y][x] = average < uint16(threshold)
		}
	}
	return pbm
}

func (ppm *PPM) SetPixel(p Point, color Pixel) {
	// Check if the point is within the PPM dimensions.
	if p.X >= 0 && p.X < ppm.width && p.Y >= 0 && p.Y < ppm.height {
		ppm.data[p.Y][p.X] = color
	}
}

func (ppm *PPM) DrawLine(p1, p2 Point, color Pixel) {
	// Bresenham's line algorithm
	x1, y1 := p1.X, p1.Y
	x2, y2 := p2.X, p2.Y

	dx := abs(x2 - x1)
	dy := abs(y2 - y1)

	var sx, sy int

	if x1 < x2 {
		sx = 1
	} else {
		sx = -1
	}

	if y1 < y2 {
		sy = 1
	} else {
		sy = -1
	}

	err := dx - dy

	for {
		ppm.SetPixel(Point{x1, y1}, color)

		if x1 == x2 && y1 == y2 {
			break
		}

		e2 := 2 * err

		if e2 > -dy {
			err -= dy
			x1 += sx
		}

		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func (ppm *PPM) DrawRectangle(p1 Point, width, height int, color Pixel) {
	ppm.DrawLine(p1, Point{X: p1.X + width - 1, Y: p1.Y}, color)
	ppm.DrawLine(Point{X: p1.X + width - 1, Y: p1.Y}, Point{X: p1.X + width - 1, Y: p1.Y + height - 1}, color)
	ppm.DrawLine(Point{X: p1.X + width - 1, Y: p1.Y + height - 1}, Point{X: p1.X, Y: p1.Y + height - 1}, color)
	ppm.DrawLine(Point{X: p1.X, Y: p1.Y + height - 1}, p1, color)
}

func (ppm *PPM) DrawFilledRectangle(p1 Point, width, height int, color Pixel) {
	ppm.DrawLine(p1, Point{X: p1.X + width - 1, Y: p1.Y}, color)
	ppm.DrawLine(Point{X: p1.X + width - 1, Y: p1.Y}, Point{X: p1.X + width - 1, Y: p1.Y + height - 1}, color)
	ppm.DrawLine(Point{X: p1.X + width - 1, Y: p1.Y + height - 1}, Point{X: p1.X, Y: p1.Y + height - 1}, color)
	ppm.DrawLine(Point{X: p1.X, Y: p1.Y + height - 1}, p1, color)

	for y := p1.Y + 1; y < p1.Y+height-1; y++ {
		ppm.DrawLine(Point{X: p1.X + 1, Y: y}, Point{X: p1.X + width - 2, Y: y}, color)
	}
}

func (ppm *PPM) DrawCircle(center Point, radius int, color Pixel) {
	x := radius
	y := 0
	err := 0

	for x >= y {
		ppm.Set(center.X+x, center.Y-y, color)
		ppm.Set(center.X+y, center.Y-x, color)
		ppm.Set(center.X-y, center.Y-x, color)
		ppm.Set(center.X-x, center.Y-y, color)
		ppm.Set(center.X-x, center.Y+y, color)
		ppm.Set(center.X-y, center.Y+x, color)
		ppm.Set(center.X+y, center.Y+x, color)
		ppm.Set(center.X+x, center.Y+y, color)

		y++
		if err <= 0 {
			err += 2*y + 1
		}

		if err > 0 {
			x--
			err -= 2*x + 1
		}
	}
}

func (ppm *PPM) DrawFilledCircle(center Point, radius int, color Pixel) {
	x := radius
	y := 0
	err := 0

	for x >= y {
		for i := center.X - x; i <= center.X+x; i++ {
			ppm.Set(i, center.Y+y, color)
			ppm.Set(i, center.Y-y, color)
		}

		for i := center.X - y; i <= center.X+y; i++ {
			ppm.Set(i, center.Y+x, color)
			ppm.Set(i, center.Y-x, color)
		}

		y++
		if err <= 0 {
			err += 2*y + 1
		}

		if err > 0 {
			x--
			err -= 2*x + 1
		}
	}
}

func (ppm *PPM) DrawTriangle(p1, p2, p3 Point, color Pixel) {
	ppm.DrawLine(p1, p2, color)
	ppm.DrawLine(p2, p3, color)
	ppm.DrawLine(p3, p1, color)
}

func (ppm *PPM) DrawFilledTriangle(p1, p2, p3 Point, color Pixel) {
	vertices := []Point{p1, p2, p3}
	sort.Slice(vertices, func(i, j int) bool {
		return vertices[i].Y < vertices[j].Y
	})

	for y := vertices[0].Y; y <= vertices[2].Y; y++ {
		x1 := interpolate(vertices[0], vertices[2], y)
		x2 := interpolate(vertices[1], vertices[2], y)

		ppm.DrawLine(Point{X: int(x1), Y: y}, Point{X: int(x2), Y: y}, color)
	}
}

func (ppm *PPM) DrawPolygon(points []Point, color Pixel) {
	for i := 0; i < len(points)-1; i++ {
		ppm.DrawLine(points[i], points[i+1], color)
	}

	ppm.DrawLine(points[len(points)-1], points[0], color)
}

func (ppm *PPM) DrawFilledPolygon(points []Point, color Pixel) {
	sort.Slice(points, func(i, j int) bool {
		return points[i].Y < points[j].Y
	})

	leftX := float64(points[0].X)
	rightX := float64(points[0].X)

	for y := points[0].Y; y <= points[len(points)-1].Y; y++ {
		for i := 0; i < len(points)-1; i++ {
			if y >= points[i].Y && y < points[i+1].Y || y >= points[i+1].Y && y < points[i].Y {
				x := interpolate(points[i], points[i+1], y)

				if x < leftX {
					leftX = x
				}
				if x > rightX {
					rightX = x
				}
			}
		}

		ppm.DrawLine(Point{X: int(leftX), Y: y}, Point{X: int(rightX), Y: y}, color)

		leftX = float64(points[0].X)
		rightX = float64(points[0].X)
	}
}

func (ppm *PPM) DrawKochSnowflake(n int, start Point, size int, color Pixel) {
	height := int(math.Sqrt(3) * float64(size) / 2)
	p1 := start
	p2 := Point{X: start.X + size, Y: start.Y}
	p3 := Point{X: start.X + size/2, Y: start.Y + height}

	ppm.KochSnowflake(n, p1, p2, color)
	ppm.KochSnowflake(n, p2, p3, color)
	ppm.KochSnowflake(n, p3, p1, color)
}

func (ppm *PPM) KochSnowflake(n int, p1, p2 Point, color Pixel) {
	if n == 0 {
		ppm.DrawLine(p1, p2, color)
	} else {
		p1Third := Point{
			X: p1.X + (p2.X-p1.X)/3,
			Y: p1.Y + (p2.Y-p1.Y)/3,
		}
		p2Third := Point{
			X: p1.X + 2*(p2.X-p1.X)/3,
			Y: p1.Y + 2*(p2.Y-p1.Y)/3,
		}

		angle := math.Pi / 3
		cosTheta := math.Cos(angle)
		sinTheta := math.Sin(angle)

		p3 := Point{
			X: int(float64(p1Third.X-p2Third.X)*cosTheta-float64(p1Third.Y-p2Third.Y)*sinTheta) + p2Third.X,
			Y: int(float64(p1Third.X-p2Third.X)*sinTheta+float64(p1Third.Y-p2Third.Y)*cosTheta) + p2Third.Y,
		}

		ppm.KochSnowflake(n-1, p1, p1Third, color)
		ppm.KochSnowflake(n-1, p1Third, p3, color)
		ppm.KochSnowflake(n-1, p3, p2Third, color)
		ppm.KochSnowflake(n-1, p2Third, p2, color)
	}
}

func (ppm *PPM) DrawSierpinskiTriangle(n int, start Point, width int, color Pixel) {

	height := int(math.Sqrt(3) * float64(width) / 2)
	p1 := start
	p2 := Point{X: start.X + width, Y: start.Y}
	p3 := Point{X: start.X + width/2, Y: start.Y + height}

	ppm.sierpinskiTriangle(n, p1, p2, p3, color)
}

func (ppm *PPM) sierpinskiTriangle(n int, p1, p2, p3 Point, color Pixel) {
	if n == 0 {
		ppm.DrawFilledTriangle(p1, p2, p3, color)
	} else {
		mid1 := Point{X: (p1.X + p2.X) / 2, Y: (p1.Y + p2.Y) / 2}
		mid2 := Point{X: (p2.X + p3.X) / 2, Y: (p2.Y + p3.Y) / 2}
		mid3 := Point{X: (p3.X + p1.X) / 2, Y: (p3.Y + p1.Y) / 2}

		ppm.sierpinskiTriangle(n-1, p3, mid2, mid3, color)
		ppm.sierpinskiTriangle(n-1, mid2, mid1, p2, color)
		ppm.sierpinskiTriangle(n-1, mid1, p1, mid3, color)
	}
}

func (ppm *PPM) DrawPerlinNoise(color1 Pixel, color2 Pixel) {
	frequency := 0.02
	amplitude := 50.0

	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			noiseValue := perlinNoise(float64(x)*frequency, float64(y)*frequency) * amplitude
			normalizedValue := (noiseValue + amplitude) / (2 * amplitude)
			interpolatedColor := interpolateColors(color1, color2, normalizedValue)
			ppm.Set(x, y, interpolatedColor)
		}
	}
}

func perlinNoise(x, y float64) float64 {
	n := int(x) + int(y)*57
	n = (n << 13) ^ n
	return (1.0 - ((float64((n*(n*n*15731+789221)+1376312589)&0x7fffffff)/1073741824.0)+1.0)/2.0)
}

func interpolateColors(color1 Pixel, color2 Pixel, t float64) Pixel {
	r := uint8(float64(color1.R)*(1-t) + float64(color2.R)*t)
	g := uint8(float64(color1.G)*(1-t) + float64(color2.G)*t)
	b := uint8(float64(color1.B)*(1-t) + float64(color2.B)*t)

	return Pixel{R: r, G: g, B: b}
}

func interpolate(p1, p2 Point, y int) float64 {
	return float64(p1.X) + float64(y-p1.Y)*(float64(p2.X-p1.X)/float64(p2.Y-p1.Y))
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
