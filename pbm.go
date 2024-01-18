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

func ReadPBM(filename string) (*PBM, error) {
	PBMtwo := PBM{}

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Erreur lors de l'ouverture")
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	booleaone := false
	booleatwo := false
	line := 0

	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "#") {
			continue
		} else if !booleaone {
			PBMtwo.magicNumber = (scanner.Text())
			booleaone = true
		} else if !booleatwo {
			size := strings.Split(scanner.Text(), " ")
			PBMtwo.width, err = strconv.Atoi(size[0])
			if err != nil {
				return nil, err
			}
			PBMtwo.height, err = strconv.Atoi(size[1])
			if err != nil {
				return nil, err
			}
			booleatwo = true

			PBMtwo.data = make(([][]bool), PBMtwo.height)
			for i := range PBMtwo.data {
				PBMtwo.data[i] = make(([]bool), PBMtwo.width)
			}
		} else {

			if PBMtwo.magicNumber == "P1" {
				a := strings.Fields(scanner.Text())
				for i := 0; i < PBMtwo.width; i++ {
					if a[i] == "1" {
						PBMtwo.data[line][i] = true
					} else {
						PBMtwo.data[line][i] = false
					}
				}
				line++
			}
			if PBMtwo.magicNumber == "P4" {
				bytenumber := 0

				if PBMtwo.width%8 == 0 {
					bytenumber = (PBMtwo.width / 8)
				} else {
					bytenumber = (PBMtwo.width / 8) + 1
				}
				padding := (bytenumber * 8) - PBMtwo.width
				binary := ToBinary(scanner.Text(), bytenumber, padding)
				colonne := 0
				linev3 := 0
				for _, linev2 := range binary {
					if colonne == PBMtwo.width {
						linev3++
						colonne = 0
					}
					if linev2 == '1' {
						PBMtwo.data[linev3][colonne] = true
					} else {
						PBMtwo.data[linev3][colonne] = false
					}
					colonne++
				}
			}
		}
	}

	return &PBM{PBMtwo.data, PBMtwo.width, PBMtwo.height, PBMtwo.magicNumber}, nil

}

func ToBinary(test string, bytenumber, padding int) string {
	var result string = ""

	for i := 0; i < len(test); i++ {
		test := fmt.Sprintf("%08b", test[i])
		if i != 0 && (i+1)%bytenumber == 0 {
			test = test[:len(test)-padding]
		}
		result += test
	}
	return result
}

func (pbm *PBM) Size() (int, int) {
	width, height := pbm.height, pbm.width
	fmt.Printf("Largeur: %d, Hauteur: %d\n", width, height)
	return width, height
}

func (pbm *PBM) At(x, y int) bool {
	if x < 0 || x > pbm.width || y < 0 || y > pbm.height {
		return false
	}
	return pbm.data[y][x]
}

func (pbm *PBM) Set(x, y int, value bool) {
	pbm.data[y][x] = value
}

func (pbm *PBM) Save(filename string) error {
	fileSave, error := os.Create(filename)
	if error != nil {
		return error
	}

	fmt.Fprintf(fileSave, "%s\n%d %d\n", pbm.magicNumber, pbm.width, pbm.height)

	for _, i := range pbm.data {
		for _, j := range i {
			if j {
				fmt.Fprint(fileSave, "1 ")
			} else {
				fmt.Fprint(fileSave, "0 ")
			}
		}
		fmt.Fprintln(fileSave)
	}
	return nil
}

func (pbm *PBM) Invert() {
	for i := 0; i < pbm.height; i++ {
		for j := 0; j < pbm.width; j++ {
			pbm.data[i][j] = !pbm.data[i][j]
		}
	}
}

func (pbm *PBM) Flip() {
	var division int = (pbm.width / 2)
	var a bool
	for i := 0; i < pbm.height; i++ {
		for j := 0; j < division; j++ {
			a = pbm.data[i][j]
			pbm.data[i][j] = pbm.data[i][pbm.width-j-1]
			pbm.data[i][pbm.width-j-1] = a
		}
	}
}

func (pbm *PBM) Flop() {
	var division int = (pbm.height / 2)
	var a bool
	for i := 0; i < pbm.width; i++ {
		for j := 0; j < division; j++ {
			a = pbm.data[j][i]
			pbm.data[j][i] = pbm.data[pbm.height-j-1][i]
			pbm.data[pbm.height-j-1][i] = a
		}
	}
}

func (pbm *PBM) SetMagicNumber(magicNumber string) {
	pbm.magicNumber = magicNumber
}
