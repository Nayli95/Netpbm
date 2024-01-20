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

// Structure représentant une image PPM
type PPM struct {
	data          [][]Pixel
	width, height int
	magicNumber   string
	max           uint
}

// Structure représentant un pixel avec des composantes rouge, verte et bleue
type Pixel struct {
	R, G, B uint8
}

// Fonction ReadPPM lit une image PPM depuis un fichier et retourne une structure représentant l'image.
func ReadPPM(filename string) (*PPM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// Lire magic number
	magicNumber, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading magic number: %v", err)
	}
	magicNumber = strings.TrimSpace(magicNumber)
	if magicNumber != "P3" && magicNumber != "P6" {
		return nil, fmt.Errorf("invalid magic number: %s", magicNumber)
	}

	// Lire les dimensions
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

	// Lire la valeur maximale
	maxValue, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading max value: %v", err)
	}
	maxValue = strings.TrimSpace(maxValue)
	var max uint
	_, err = fmt.Sscanf(maxValue, "%d", &max)
	if err != nil {
		return nil, fmt.Errorf("invalid max value: %v", err)
	}

	// Lire les données de l'image
	data := make([][]Pixel, height)
	expectedBytesPerPixel := 3

	if magicNumber == "P3" {
		// Lire le format P3 (ASCII)
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
		// Lire le format P6 (binaire)
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

	// Retourner la structure PPM
	return &PPM{data, width, height, magicNumber, max}, nil
}

// Fonction PrintPPM affiche les informations de base et les données des pixels de l'image PPM.
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

// Fonction Size retourne la largeur et la hauteur de l'image PPM.
func (ppm *PPM) Size() (int, int) {
	return ppm.width, ppm.height
}

// Fonction At retourne le pixel aux coordonnées spécifiées (x, y) dans l'image PPM.
func (ppm *PPM) At(x, y int) Pixel {
	// Vérification des limites pour éviter les erreurs d'index
	if x < 0 || x >= ppm.width || y < 0 || y >= ppm.height {
		panic("Index out of bounds")
	}

	return ppm.data[y][x]
}

// Fonction Set affecte la valeur d'un pixel aux coordonnées spécifiées (x, y) dans l'image PPM.
func (ppm *PPM) Set(x, y int, value Pixel) {
	// Vérification des limites pour éviter les erreurs d'index
	if x < 0 || x >= ppm.width || y < 0 || y >= ppm.height {
		panic("Index out of bounds")
	}

	ppm.data[y][x] = value
}

// Fonction Save enregistre l'image PPM dans un fichier spécifié par le nom de fichier.
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

// Fonction Invert inverse les composantes de couleur de tous les pixels de l'image PPM.
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

// Fonction Flip inverse l'ordre des pixels horizontalement dans l'image PPM.
func (ppm *PPM) Flip() {
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width/2; x++ {
			ppm.data[y][x], ppm.data[y][ppm.width-x-1] = ppm.data[y][ppm.width-x-1], ppm.data[y][x]
		}
	}
}

// Fonction Flop inverse l'ordre des lignes verticalement dans l'image PPM.
func (ppm *PPM) Flop() {
	for y := 0; y < ppm.height/2; y++ {
		ppm.data[y], ppm.data[ppm.height-y-1] = ppm.data[ppm.height-y-1], ppm.data[y]
	}
}

// Fonction SetMagicNumber affecte le magic number d'une image PPM.
func (ppm *PPM) SetMagicNumber(magicNumber string) {
	ppm.magicNumber = magicNumber
}

// SetMaxValue met à jour la valeur maximale des pixels dans la structure PPM et ajuste les valeurs des pixels dans les données en fonction de la nouvelle valeur maximale.
func (ppm *PPM) SetMaxValue(maxValue uint8) {
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			// Mettre à l'échelle les valeurs RGB en fonction de la nouvelle valeur maximale
			ppm.data[y][x].R = uint8(float64(ppm.data[y][x].R) * float64(maxValue) / float64(ppm.max))
			ppm.data[y][x].G = uint8(float64(ppm.data[y][x].G) * float64(maxValue) / float64(ppm.max))
			ppm.data[y][x].B = uint8(float64(ppm.data[y][x].B) * float64(maxValue) / float64(ppm.max))
		}
	}

	// Mettre à jour la valeur maximale
	ppm.max = uint(maxValue)
}

// Fonction Rotate90CW fait pivoter l'image PPM actuelle de 90 degrés dans le sens des aiguilles d'une montre.
func (ppm *PPM) Rotate90CW() {
	// Créer une nouvelle structure PPM pour contenir l'image pivotée
	newPPM := PPM{
		data:        make([][]Pixel, ppm.width),
		width:       ppm.height,
		height:      ppm.width,
		magicNumber: ppm.magicNumber,
		max:         ppm.max,
	}

	// Initialiser le tableau bidimensionnel dans la nouvelle structure PPM
	for i := range newPPM.data {
		newPPM.data[i] = make([]Pixel, newPPM.width)
	}

	// Effectuer la rotation en copiant les pixels de l'image actuelle vers la nouvelle structure pivotée
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			newPPM.data[x][ppm.height-y-1] = ppm.data[y][x]
		}
	}

	// Mettre à jour la structure PPM actuelle avec la nouvelle image pivotée
	*ppm = newPPM
}

// Fonction ToPGM convertit l'image PPM en une image PGM (niveaux de gris).
func (ppm *PPM) ToPGM() *PGM {
	// Créer une nouvelle structure PGM pour contenir l'image en niveaux de gris
	pgm := &PGM{
		width:       ppm.width,
		height:      ppm.height,
		magicNumber: "P2",
		max:         ppm.max,
	}

	// Initialiser le tableau bidimensionnel dans la nouvelle structure PGM
	pgm.data = make([][]uint8, ppm.height)
	for i := range pgm.data {
		pgm.data[i] = make([]uint8, ppm.width)
	}

	// Convertir chaque pixel RGB en niveaux de gris et les assigner à la nouvelle structure PGM
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			// Convertir RGB en niveaux de gris
			gray := uint8((int(ppm.data[y][x].R) + int(ppm.data[y][x].G) + int(ppm.data[y][x].B)) / 3)
			pgm.data[y][x] = gray
		}
	}

	return pgm
}

// Structure représentant un point avec des coordonnées X et Y.
type Point struct {
	X, Y int
}

// Fonction ToPBM convertit l'image PPM en une image PBM (noir et blanc).
func (ppm *PPM) ToPBM() *PBM {
	pbm := &PBM{
		width:       ppm.width,
		height:      ppm.height,
		magicNumber: "P1",
	}

	// Initialiser le tableau bidimensionnel dans la nouvelle structure PBM
	pbm.data = make([][]bool, ppm.height)
	for i := range pbm.data {
		pbm.data[i] = make([]bool, ppm.width)
	}

	// Définir le seuil pour la conversion en noir et blanc
	threshold := uint8(ppm.max / 2)

	// Convertir chaque pixel en noir et blanc et les assigner à la nouvelle structure PBM
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			average := (uint16(ppm.data[y][x].R) + uint16(ppm.data[y][x].G) + uint16(ppm.data[y][x].B)) / 3
			pbm.data[y][x] = average > uint16(threshold)
		}
	}

	return pbm
}

// Fonction SetPixel définit la couleur d'un pixel à une position spécifiée dans l'image PPM.
func (ppm *PPM) SetPixel(p Point, color Pixel) {
	// Vérifier si le point est dans les dimensions de l'image PPM.
	if p.X >= 0 && p.X < ppm.width && p.Y >= 0 && p.Y < ppm.height {
		ppm.data[p.Y][p.X] = color
	}
}

// Fonction DrawLine utilise l'algorithme de Bresenham pour dessiner une ligne entre deux points dans l'image PPM.
func (ppm *PPM) DrawLine(p1, p2 Point, color Pixel) {
	// Algorithme de tracé de ligne de Bresenham
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
	p2 := Point{p1.X + width, p1.Y}
	p3 := Point{p1.X + width, p1.Y + height}
	p4 := Point{p1.X, p1.Y + height}

	// Dessiner les quatre côtés du rectangle
	ppm.DrawLine(p1, p2, color)
	ppm.DrawLine(p2, p3, color)
	ppm.DrawLine(p3, p4, color)
	ppm.DrawLine(p4, p1, color)
}

func (ppm *PPM) DrawFilledRectangle(p1 Point, width, height int, color Pixel) {
	// Si la hauteur du rectangle est dans la limite du fichier
	if p1.Y+height < ppm.height {
		for y := p1.Y; y < p1.Y+height; y++ {
			// Si la largeur du rectangle est dans les données du fichier
			if p1.X+width < ppm.width {
				for x := p1.X; x <= p1.X+width; x++ {
					ppm.data[y][x] = color
				}
				// Si la largeur du rectangle dépasse les données du fichier
			} else if p1.X+width > ppm.width {
				for x := p1.X; x < ppm.width; x++ {
					ppm.data[y][x] = color
				}
			}
		}
		// Si la hauteur du rectangle dépasse la limite du fichier
	} else if p1.Y+height > ppm.height {
		for y := p1.Y; y < ppm.height; y++ {
			// Si la largeur du rectangle est dans les données du fichier
			if p1.X+width < ppm.width {
				for x := p1.X; x <= p1.X+width; x++ {
					ppm.data[y][x] = color
				}
				// Si la largeur du rectangle dépasse les données du fichier
			} else if p1.X+width > ppm.width {
				for x := p1.X; x < ppm.width; x++ {
					ppm.data[y][x] = color
				}
			}
		}
	}
}

// Fonction DrawCircle dessine un cercle avec le centre, le rayon et la couleur spécifiés.
func (ppm *PPM) DrawCircle(center Point, radius int, color Pixel) {
	// Parcourir chaque pixel
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			// Calculer la distance du pixel actuel au centre du cercle
			dx := float64(x - center.X)
			dy := float64(y - center.Y)
			distance := math.Sqrt(dx*dx + dy*dy)
			// Vérifier si la distance est approximativement égale au rayon spécifié
			if math.Abs(distance-float64(radius)*0.85) < 0.5 {
				ppm.data[y][x] = color
			}
		}
	}
}

func (ppm *PPM) DrawFilledCircle(center Point, radius int, color Pixel) {
	// Dessiner un cercle avec le rayon de plus en plus petit jusqu'à ce qu'il atteigne 0
	for radius >= 0 {
		ppm.DrawCircle(center, radius, color)
		radius--
	}
}

// Fonction DrawTriangle dessine un triangle avec les points spécifiés et la couleur.
func (ppm *PPM) DrawTriangle(p1, p2, p3 Point, color Pixel) {
	// Dessine les trois côtés du triangle en utilisant la fonction DrawLine.
	ppm.DrawLine(p1, p2, color)
	ppm.DrawLine(p2, p3, color)
	ppm.DrawLine(p3, p1, color)
}

// Fonction DrawFilledTriangle dessine un triangle rempli avec les points spécifiés et la couleur.
func (ppm *PPM) DrawFilledTriangle(p1, p2, p3 Point, color Pixel) {
	// Utilise la fonction DrawLine pour dessiner chaque ligne horizontale en utilisant le point de départ, la taille et la couleur donnés.
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

// Fonction DrawPolygon dessine un polygone avec les points spécifiés et la couleur.
func (ppm *PPM) DrawPolygon(points []Point, color Pixel) {
	// Dessine les côtés du polygone en utilisant la fonction DrawLine.
	for i := 0; i < len(points)-1; i++ {
		ppm.DrawLine(points[i], points[i+1], color)
	}

	ppm.DrawLine(points[len(points)-1], points[0], color)
}

func (ppm *PPM) DrawFilledPolygon(points []Point, color Pixel) {
	// Dessine le contour du polygone
	ppm.DrawPolygon(points, color)

	// Parcourt chaque ligne de l'image PPM
	for i := 0; i < ppm.height; i++ {
		// positions : indices des pixels de la ligne ayant la couleur spécifiée
		var positions []int
		// number_points : nombre de pixels de la ligne avec la couleur spécifiée
		var number_points int

		// Parcourt chaque colonne de la ligne
		for j := 0; j < ppm.width; j++ {
			if ppm.data[i][j] == color {
				number_points++
				positions = append(positions, j)
			}
		}

		// Si plus d'un point, remplir la zone entre les deux premiers points
		if number_points > 1 {
			for k := positions[0] + 1; k < positions[len(positions)-1]; k++ {
				ppm.data[i][k] = color
			}
		}
	}
}

// La fonction DrawKochSnowflake dessine un flocon de Koch avec la récursivité spécifiée, en utilisant le point de départ, la taille et la couleur donnés.
func (ppm *PPM) DrawKochSnowflake(n int, start Point, size int, color Pixel) {
	// Calcule la hauteur du triangle équilatéral inscrit dans le flocon de Koch.
	height := int(math.Sqrt(3) * float64(size) / 2)
	// Définit les trois points du triangle de base du flocon.
	p1 := start
	p2 := Point{X: start.X + size, Y: start.Y}
	p3 := Point{X: start.X + size/2, Y: start.Y + height}

	// Appelle la fonction KochSnowflake pour chaque côté du triangle.
	ppm.KochSnowflake(n, p1, p2, color)
	ppm.KochSnowflake(n, p2, p3, color)
	ppm.KochSnowflake(n, p3, p1, color)
}

// La fonction KochSnowflake dessine un segment du flocon de Koch avec la récursivité spécifiée, en utilisant les points de départ et d'arrivée, ainsi que la couleur donnée.
func (ppm *PPM) KochSnowflake(n int, p1, p2 Point, color Pixel) {
	// Cas de base : si la récursivité atteint zéro, dessine le segment.
	if n == 0 {
		ppm.DrawLine(p1, p2, color)
	} else {
		// Calcule les deux tiers des côtés du segment.
		p1Third := Point{
			X: p1.X + (p2.X-p1.X)/3,
			Y: p1.Y + (p2.Y-p1.Y)/3,
		}
		p2Third := Point{
			X: p1.X + 2*(p2.X-p1.X)/3,
			Y: p1.Y + 2*(p2.Y-p1.Y)/3,
		}

		// Calcule le point médian du segment avec une rotation de 60 degrés.
		angle := math.Pi / 3
		cosTheta := math.Cos(angle)
		sinTheta := math.Sin(angle)

		p3 := Point{
			X: int(float64(p1Third.X-p2Third.X)*cosTheta-float64(p1Third.Y-p2Third.Y)*sinTheta) + p2Third.X,
			Y: int(float64(p1Third.X-p2Third.X)*sinTheta+float64(p1Third.Y-p2Third.Y)*cosTheta) + p2Third.Y,
		}

		// Appelle récursivement la fonction pour les quatre segments résultants.
		ppm.KochSnowflake(n-1, p1, p1Third, color)
		ppm.KochSnowflake(n-1, p1Third, p3, color)
		ppm.KochSnowflake(n-1, p3, p2Third, color)
		ppm.KochSnowflake(n-1, p2Third, p2, color)
	}
}

// Fonction DrawSierpinskiTriangle dessine un triangle de Sierpinski avec la récursivité spécifiée, le point de départ, la largeur et la couleur.
func (ppm *PPM) DrawSierpinskiTriangle(n int, start Point, width int, color Pixel) {
	// Dessine le triangle de Sierpinski en utilisant la récursivité.
	height := int(math.Sqrt(3) * float64(width) / 2)
	p1 := start
	p2 := Point{X: start.X + width, Y: start.Y}
	p3 := Point{X: start.X + width/2, Y: start.Y + height}

	ppm.sierpinskiTriangle(n, p1, p2, p3, color)
}

// La fonction sierpinskiTriangle dessine un triangle de Sierpinski avec la récursivité spécifiée, en utilisant les points de départ, la largeur, et la couleur donnés.
func (ppm *PPM) sierpinskiTriangle(n int, p1, p2, p3 Point, color Pixel) {
	// Cas de base : si la récursivité atteint zéro, dessine le triangle rempli.
	if n == 0 {
		ppm.DrawFilledTriangle(p1, p2, p3, color)
	} else {
		// Calcule les points médians des côtés du triangle.
		mid1 := Point{X: (p1.X + p2.X) / 2, Y: (p1.Y + p2.Y) / 2}
		mid2 := Point{X: (p2.X + p3.X) / 2, Y: (p2.Y + p3.Y) / 2}
		mid3 := Point{X: (p3.X + p1.X) / 2, Y: (p3.Y + p1.Y) / 2}

		// Appelle récursivement la fonction pour les trois sous-triangles.
		ppm.sierpinskiTriangle(n-1, p3, mid2, mid3, color)
		ppm.sierpinskiTriangle(n-1, mid2, mid1, p2, color)
		ppm.sierpinskiTriangle(n-1, mid1, p1, mid3, color)
	}
}

// Fonction DrawPerlinNoise dessine du bruit de Perlin avec les couleurs spécifiées.
func (ppm *PPM) DrawPerlinNoise(color1 Pixel, color2 Pixel) {
	// Utilise le bruit de Perlin pour générer des valeurs de hauteur et les convertit en couleurs pour créer un effet de texture.
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

// La fonction perlinNoise génère une valeur de bruit de Perlin en 2D à partir des coordonnées x et y fournies.
func perlinNoise(x, y float64) float64 {
	// Utilise un nombre entier unique basé sur x et y pour générer une séquence pseudo-aléatoire.
	n := int(x) + int(y)*57
	n = (n << 13) ^ n
	// Normalise la valeur du bruit dans la plage [0, 1].
	return (1.0 - ((float64((n*(n*n*15731+789221)+1376312589)&0x7fffffff)/1073741824.0)+1.0)/2.0)
}

// La fonction interpolateColors effectue une interpolation linéaire entre deux couleurs (color1 et color2) en utilisant le facteur t (dans la plage [0, 1]).
func interpolateColors(color1 Pixel, color2 Pixel, t float64) Pixel {
	// Interpole séparément chaque composante RGB des deux couleurs.
	r := uint8(float64(color1.R)*(1-t) + float64(color2.R)*t)
	g := uint8(float64(color1.G)*(1-t) + float64(color2.G)*t)
	b := uint8(float64(color1.B)*(1-t) + float64(color2.B)*t)

	// Retourne la nouvelle couleur interpolée.
	return Pixel{R: r, G: g, B: b}
}

// La fonction interpolate effectue une interpolation linéaire entre deux points (p1 et p2) selon la valeur y fournie.
func interpolate(p1, p2 Point, y int) float64 {
	// Utilise la formule de l'interpolation linéaire pour calculer la valeur interpolée.
	return float64(p1.X) + float64(y-p1.Y)*(float64(p2.X-p1.X)/float64(p2.Y-p1.Y))
}

// La fonction abs renvoie la valeur absolue d'un entier.
func abs(x int) int {
	// Vérifie si x est négatif et renvoie -x, sinon renvoie x.
	if x < 0 {
		return -x
	}
	return x
}

// KNearestNeighbors redimensionne l'image PPM en utilisant l'algorithme des k-plus proches voisins.
func (ppm *PPM) KNearestNeighbors(newWidth, newHeight int) {
	// Calculer les facteurs d'échelle pour la largeur et la hauteur
	scaleX := float64(ppm.width) / float64(newWidth)
	scaleY := float64(ppm.height) / float64(newHeight)

	// Créer une nouvelle image PPM avec les dimensions souhaitées
	resizedPPM := &PPM{
		data:        make([][]Pixel, newHeight),
		width:       newWidth,
		height:      newHeight,
		magicNumber: ppm.magicNumber,
		max:         ppm.max,
	}

	// Initialiser les données de pixels pour l'image redimensionnée
	for i := range resizedPPM.data {
		resizedPPM.data[i] = make([]Pixel, newWidth)
	}

	// Itérer sur chaque pixel de l'image redimensionnée
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			// Calculer les coordonnées de pixel correspondantes dans l'image originale
			origX := int(float64(x) * scaleX)
			origY := int(float64(y) * scaleY)

			// Trouver les k-plus proches voisins
			neighbors := make([]Pixel, 0)
			for i := -1; i <= 1; i++ {
				for j := -1; j <= 1; j++ {
					nx := origX + i
					ny := origY + j

					// S'assurer que les coordonnées sont dans les limites de l'image originale
					if nx >= 0 && nx < ppm.width && ny >= 0 && ny < ppm.height {
						neighbors = append(neighbors, ppm.data[ny][nx])
					}
				}
			}

			// Calculer la couleur moyenne des k-plus proches voisins
			var totalR, totalG, totalB uint64
			for _, color := range neighbors {
				totalR += uint64(color.R)
				totalG += uint64(color.G)
				totalB += uint64(color.B)
			}

			avgR := uint8(totalR / uint64(len(neighbors)))
			avgG := uint8(totalG / uint64(len(neighbors)))
			avgB := uint8(totalB / uint64(len(neighbors)))

			// Définir la couleur du pixel dans l'image redimensionnée
			resizedPPM.data[y][x] = Pixel{R: avgR, G: avgG, B: avgB}
		}
	}

	// Remplacer l'image originale par l'image redimensionnée
	*ppm = *resizedPPM
}
