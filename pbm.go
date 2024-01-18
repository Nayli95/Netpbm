package Netpbm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type PBM struct {
	data          [][]bool
	width, height int
	magicNumber   string
}

// ReadPBM lie l'image PBM du fichier et retourne dans la struct avec les infos de l'image.
func ReadPBM(filename string) (*PBM, error) {
	pbm := PBM{}

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Scanner for text-based information
	scanner := bufio.NewScanner(file)
	// reader := bufio.NewReader(file)
	confirmOne := false
	confirmTwo := false
	Line := 0

	// Ignore comments and empty lines
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}

		if strings.HasPrefix(scanner.Text(), "#") {
			continue
		} else if !confirmOne {
			pbm.magicNumber = scanner.Text()
			confirmOne = true
		} else if !confirmTwo {
			sepa := strings.Fields(scanner.Text())
			if len(sepa) > 0 {
				pbm.width, _ = strconv.Atoi(sepa[0])
				pbm.height, _ = strconv.Atoi(sepa[1])
			}
			confirmTwo = true

			pbm.data = make([][]bool, pbm.height)
			for i := range pbm.data {
				pbm.data[i] = make([]bool, pbm.width)
			}

		} else {
			if pbm.magicNumber == "P1" {
				// Process P1 format
				test := strings.Fields(scanner.Text())
				for i := 0; i < pbm.width; i++ {
					if test[i] == "1" {
						pbm.data[Line][i] = true
					} else {
						pbm.data[Line][i] = false
					}
				}
				Line++
			} else if pbm.magicNumber == "P4" {
				expectedBytesPerRow := (pbm.width + 7) / 8
				totalExpectedBytes := expectedBytesPerRow * pbm.height
				fmt.Printf("Expected total bytes for pixel data: %d\n", totalExpectedBytes)

				// Create a buffer to hold all pixel data
				allPixelData := make([]byte, totalExpectedBytes)

				// Read the file content directly into the buffer
				fileContent, err := os.ReadFile(filename)
				if err != nil {
					return nil, fmt.Errorf("error reading file: %v", err)
				}

				// Copy the relevant part of the file content into the pixel data buffer
				copy(allPixelData, fileContent[len(fileContent)-totalExpectedBytes:])

				// Process the buffer to update pbm.data
				byteIndex := 0
				for y := 0; y < pbm.height; y++ {
					for x := 0; x < pbm.width; x++ {
						if x%8 == 0 && x != 0 {
							byteIndex++
						}
						pbm.data[y][x] = (allPixelData[byteIndex]>>(7-(x%8)))&1 != 0
					}
					byteIndex++
				}

			}
		}
	}

	fmt.Printf("%v\n", pbm)
	return &pbm, nil
}

// Size retourne la hauteur et la largeur de l'image.
func (pbm *PBM) Size() (int, int) {
	return pbm.height, pbm.width
}

// At retourne la valeur de chaque pixel en (x, y).
func (pbm *PBM) At(x, y int) bool {
	return pbm.data[x][y]
}

// Set défini la valeur de chaque pixel à (x, y).
func (pbm *PBM) Set(x, y int, value bool) {
	pbm.data[x][y] = value
}

// Save saves the PBM image to a file and returns an error if there was a problem.
func (pbm *PBM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// Ecrire le magique number et la taille de l'image dans le nouveau fichier
	_, err = fmt.Fprintf(file, "%s\n%d %d\n", pbm.magicNumber, pbm.width, pbm.height)
	if err != nil {
		return fmt.Errorf("error writing magic number and dimensions: %v", err)
	}

	// Entrer les donnees de l'image
	if pbm.magicNumber == "P1" { //Pour le P1
		for _, row := range pbm.data {
			for _, pixel := range row {
				if pixel {
					_, err = file.WriteString("1 ")
				} else {
					_, err = file.WriteString("0 ")
				}
				if err != nil {
					return fmt.Errorf("error writing pixel data: %v", err)
				}
			}
			_, err = file.WriteString("\n")
			if err != nil {
				return fmt.Errorf("error writing pixel data: %v", err)
			}
		}
	} else if pbm.magicNumber == "P4" { // Pour le P4
		for _, row := range pbm.data {
			for x := 0; x < pbm.width; x += 8 {
				var byteValue byte
				for i := 0; i < 8 && x+i < pbm.width; i++ {
					bitIndex := 7 - i
					if row[x+i] {
						byteValue |= 1 << bitIndex
					}
				}
				_, err = file.Write([]byte{byteValue})
				if err != nil {
					return fmt.Errorf("error writing pixel data: %v", err)
				}
			}
		}
	}

	return nil
}

// Invert inverse les couleurs de l'image PBM.
func (pbm *PBM) Invert() {
	for y := 0; y < pbm.height; y++ {
		for x := 0; x < pbm.width; x++ {
			pbm.data[y][x] = !pbm.data[y][x]
		}
	}
}

// Flip flips the PBM image horizontally.
func (pbm *PBM) Flip() {
	for y := 0; y < pbm.height; y++ {
		start := make([]bool, pbm.width)
		end := make([]bool, pbm.width)
		copy(start, pbm.data[y])
		for i := 0; i < len(start); i++ {
			end[i] = start[len(start)-1-i]
		}
		copy(pbm.data[y], end[:])

	}
}

// Flop flops the PBM image vertically.
func (pbm *PBM) Flop() {
	cursor := pbm.height - 1
	for y := range pbm.data {
		temp := pbm.data[y]
		pbm.data[y] = pbm.data[cursor]
		pbm.data[cursor] = temp
		cursor--
		if cursor < y || cursor == y {
			break
		}
	}
}

// SetMagicNumber sets the magic number of the PBM image.
func (pbm *PBM) SetMagicNumber(magicNumber string) {
	pbm.magicNumber = magicNumber

}
