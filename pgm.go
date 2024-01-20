package Netpbm

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type PGM struct {
	data          [][]uint8
	width, height int
	magicNumber   string
	max           uint
}

func ReadPGM(filename string) (*PGM, error) {
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
	if magicNumber != "P2" && magicNumber != "P5" {
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

	// Lire l'image data
	data := make([][]uint8, height)
	expectedBytesPerPixel := 1

	if magicNumber == "P2" {
		// Lire le format P2 (ASCII)
		for y := 0; y < height; y++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("error reading data at row %d: %v", y, err)
			}
			fields := strings.Fields(line)
			rowData := make([]uint8, width)
			for x, field := range fields {
				if x >= width {
					return nil, fmt.Errorf("index out of range at row %d", y)
				}
				var pixelValue uint8
				_, err := fmt.Sscanf(field, "%d", &pixelValue)
				if err != nil {
					return nil, fmt.Errorf("error parsing pixel value at row %d, column %d: %v", y, x, err)
				}
				rowData[x] = pixelValue
			}
			data[y] = rowData
		}
	} else if magicNumber == "P5" {
		// Lire le format P5 (binary)
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

			rowData := make([]uint8, width)
			for x := 0; x < width; x++ {
				pixelValue := uint8(row[x*expectedBytesPerPixel])
				rowData[x] = pixelValue
			}
			data[y] = rowData
		}
	}

	// Return la struct PGM
	return &PGM{data, width, height, magicNumber, max}, nil
}

// Size retourne la largeur et la hauteur de l'image PGM.
func (pgm *PGM) Size() (int, int) {
	return pgm.width, pgm.height
}

// At retourne la valeur du pixel à la position (x, y) dans l'image PGM.
func (pgm *PGM) At(x, y int) uint8 {
	if x >= 0 && x < pgm.width && y >= 0 && y < pgm.height {
		return pgm.data[y][x]
	}
	return 0
}

// Set définit la valeur du pixel à la position (x, y).
func (pgm *PGM) Set(x, y int, value uint8) {
	if x >= 0 && x < pgm.width && y >= 0 && y < pgm.height {
		pgm.data[y][x] = value
	}
}

// Save enregistre l'image PGM dans un fichier au format opposé (P2 ou P5) et retourne une erreur en cas de problème.
func (pgm *PGM) Save(filename string) error {
	// Ouvrir le fichier pour écriture.
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Créer un écrivain pour le fichier.
	writer := bufio.NewWriter(file)
	_, err = fmt.Fprintln(writer, pgm.magicNumber)
	if err != nil {
		return fmt.Errorf("error writing magic number: %v", err)
	}

	// Écrire les dimensions.
	_, err = fmt.Fprintf(writer, "%d %d\n", pgm.width, pgm.height)
	if err != nil {
		return fmt.Errorf("error writing dimensions: %v", err)
	}

	// Écrire la valeur maximale.
	_, err = fmt.Fprintln(writer, pgm.max)
	if err != nil {
		return fmt.Errorf("error writing max value: %v", err)
	}
	for _, row := range pgm.data {
		if len(row) != pgm.width {
			return fmt.Errorf("inconsistent row length in data")
		}
	}

	// Écrire les data de l'image.
	if pgm.magicNumber == "P2" {
		err = saveP2PGM(writer, pgm)
		if err != nil {
			return err
		}
	} else if pgm.magicNumber == "P5" {
		err = saveP5PGM(writer, pgm)
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

// saveP2PGM enregistre l'image PGM dans le format P2 (ASCII).
func saveP2PGM(file *bufio.Writer, pgm *PGM) error {
	for y := 0; y < pgm.height; y++ {
		for x := 0; x < pgm.width; x++ {
			// Écrire la valeur du pixel.
			_, err := fmt.Fprint(file, pgm.data[y][x])
			if err != nil {
				return fmt.Errorf("error writing pixel data at row %d, column %d: %v", y, x, err)
			}

			// Ajouter un espace après chaque pixel, sauf le dernier de la ligne.
			if x < pgm.width-1 {
				_, err = fmt.Fprint(file, " ")
				if err != nil {
					return fmt.Errorf("error writing space after pixel at row %d, column %d: %v", y, x, err)
				}
			}
		}

		// Ajouter une newline après chaque ligne.
		_, err := fmt.Fprintln(file)
		if err != nil {
			return fmt.Errorf("error writing newline after row %d: %v", y, err)
		}
	}
	return nil
}

// saveP5PGM enregistre l'image PGM dans le format P5 (binaire).
func saveP5PGM(file *bufio.Writer, pgm *PGM) error {
	for y := 0; y < pgm.height; y++ {
		row := make([]byte, pgm.width)
		for x := 0; x < pgm.width; x++ {
			row[x] = byte(pgm.data[y][x])
		}
		_, err := file.Write(row)
		if err != nil {
			return fmt.Errorf("error writing pixel data at row %d: %v", y, err)
		}
	}
	return nil
}

// Invert inverse les couleurs de l'image PGM.
func (pgm *PGM) Invert() {
	for i := range pgm.data {
		for j := range pgm.data[i] {
			pgm.data[i][j] = uint8(pgm.max) - pgm.data[i][j]
		}
	}
}

// Flip retourne l'image PGM horizontalement.
func (pgm *PGM) Flip() {
	for i := range pgm.data {
		for j, k := 0, len(pgm.data[i])-1; j < k; j, k = j+1, k-1 {
			pgm.data[i][j], pgm.data[i][k] = pgm.data[i][k], pgm.data[i][j]
		}
	}
}

// Flop retourne l'image PGM verticalement.
func (pgm *PGM) Flop() {
	for i := 0; i < pgm.height/2; i++ {
		pgm.data[i], pgm.data[pgm.height-i-1] = pgm.data[pgm.height-i-1], pgm.data[i]
	}
}

// SetMagicNumber définit le numéro magique de l'image PGM.
func (pgm *PGM) SetMagicNumber(magicNumber string) {
	pgm.magicNumber = magicNumber
}

// SetMaxValue définit la valeur maximale des pixels dans l'image PGM.
func (pgm *PGM) SetMaxValue(maxValue uint8) {
	for y := 0; y < pgm.height; y++ {
		for x := 0; x < pgm.width; x++ {
			// Calculer la nouvelle valeur du pixel en ajustant l'échelle selon la nouvelle valeur maximale.
			scaledValue := float64(pgm.data[y][x]) * float64(maxValue) / float64(pgm.max)

			// Convertir la valeur à virgule flottante en entier non signé et mettre à jour la valeur du pixel.
			newValue := uint8(scaledValue)
			pgm.data[y][x] = newValue
		}
	}

	pgm.max = uint(maxValue)
}

// Rotate90CW fait pivoter l'image PGM de 90 degrés dans le sens des aiguilles d'une montre.
func (pgm *PGM) Rotate90CW() {
	// Vérifier que les dimensions de l'image sont valides.
	if pgm.width <= 0 || pgm.height <= 0 {
		return
	}

	// Créer un nouveau tableau pour stocker les données pivotées.
	newData := make([][]uint8, pgm.width)
	for i := 0; i < pgm.width; i++ {
		newData[i] = make([]uint8, pgm.height)
		for j := 0; j < pgm.height; j++ {
			// Effectuer la rotation en échangeant les indices i et j.
			newData[i][j] = pgm.data[pgm.height-j-1][i]
		}
	}

	// Mettre à jour les données de l'image et échanger les dimensions.
	pgm.data = newData
	pgm.width, pgm.height = pgm.height, pgm.width
}

// ToPBM convertit l'image PGM en une image PBM (Portable Bitmap).
func (pgm *PGM) ToPBM() *PBM {
	// Créer une nouvelle image PBM avec les mêmes dimensions.
	pbm := &PBM{
		data:        make([][]bool, pgm.height),
		width:       pgm.width,
		height:      pgm.height,
		magicNumber: "P1",
	}

	// Remplir les données de l'image PBM en convertissant les valeurs de pixels en valeurs booléennes.
	for y := 0; y < pgm.height; y++ {
		pbm.data[y] = make([]bool, pgm.width)
		for x := 0; x < pgm.width; x++ {
			pbm.data[y][x] = pgm.data[y][x] < uint8(pgm.max/2)
		}
	}

	// Retourner l'image PBM résultante.
	return pbm
}

// PrintData affiche les données de l'image PGM.
func (pgm *PGM) PrintData() {
	// Parcourir chaque ligne et colonne de l'image et afficher la valeur du pixel.
	for i := 0; i < pgm.height; i++ {
		for j := 0; j < pgm.width; j++ {
			fmt.Printf("%d ", pgm.data[i][j])
		}
		fmt.Println()
	}
}
