package Netpbm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type PGM struct {
	data        [][]uint8
	width       int
	height      int
	magicNumber string
	max         uint8
}

func ReadPGM(filename string) (*PGM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	pgm := &PGM{}
	line := 0

	for scanner.Scan() {
		text := scanner.Text()

		if strings.HasPrefix(text, "#") {
			continue
		}

		if pgm.magicNumber == "" {
			pgm.magicNumber = strings.TrimSpace(text)
		} else if pgm.width == 0 {
			fmt.Sscanf(text, "%d %d", &pgm.width, &pgm.height)
			pgm.data = make([][]uint8, pgm.height)
			for i := range pgm.data {
				pgm.data[i] = make([]uint8, pgm.width)
			}
		} else if pgm.max == 0 {
			fmt.Sscanf(text, "%d", &pgm.max)
		} else {
			if pgm.magicNumber == "P2" {
				val := strings.Fields(text)
				for i := 0; i < pgm.width; i++ {
					num, _ := strconv.ParseUint(val[i], 10, 8)
					pgm.data[line][i] = uint8(num)
				}
				line++
			} else if pgm.magicNumber == "P5" {
				pixelData := make([]uint8, pgm.width*pgm.height)
				fileContent, err := os.ReadFile(filename)
				if err != nil {
					return nil, fmt.Errorf("couldn't read file: %v", err)
				}
				copy(pixelData, fileContent[len(fileContent)-(pgm.width*pgm.height):])
				pixelIndex := 0
				for y := 0; y < pgm.height; y++ {
					for x := 0; x < pgm.width; x++ {
						pgm.data[y][x] = pixelData[pixelIndex]
						pixelIndex++
					}
				}
				break
			}
		}
	}

	return pgm, nil
}

func (pgm *PGM) Size() (int, int) {
	return pgm.width, pgm.height
}

func (pgm *PGM) At(x, y int) uint8 {
	return pgm.data[y][x]
}

func (pgm *PGM) Set(x, y int, value uint8) {
	pgm.data[y][x] = value
}

func (pgm *PGM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	fmt.Fprint(writer, pgm.magicNumber+"\n")
	fmt.Fprintf(writer, "%d %d\n", pgm.width, pgm.height)
	fmt.Fprintf(writer, "%d\n", pgm.max)
	writer.Flush()

	if pgm.magicNumber == "P2" {
		for y, row := range pgm.data {
			for i, pixel := range row {
				xtra := " "
				if i == len(row)-1 {
					xtra = ""
				}
				fmt.Fprint(writer, strconv.Itoa(int(pixel))+xtra)
			}
			if y != len(pgm.data)-1 {
				fmt.Fprintln(writer, "")
			}
		}
		writer.Flush()
	} else if pgm.magicNumber == "P5" {
		for _, row := range pgm.data {
			for _, pixel := range row {
				_, err = file.Write([]byte{pixel})
				if err != nil {
					return fmt.Errorf("error writing pixel data: %v", err)
				}
			}
		}
	}

	return nil
}

func (pgm *PGM) Invert() {
	for y, row := range pgm.data {
		for x := range row {
			pgm.data[y][x] = pgm.max - pgm.data[y][x]
		}
	}
}

func (pgm *PGM) Flip() {
	for y, row := range pgm.data {
		cursor := pgm.width - 1
		for x := 0; x < pgm.width/2; x++ {
			temp := row[x]
			pgm.data[y][x] = row[cursor]
			pgm.data[y][cursor] = temp
			cursor--
		}
	}
}

func (pgm *PGM) Flop() {
	cursor := pgm.height - 1
	for y := 0; y < pgm.height/2; y++ {
		temp := pgm.data[y]
		pgm.data[y] = pgm.data[cursor]
		pgm.data[cursor] = temp
		cursor--
	}
}

func (pgm *PGM) SetMagicNumber(magicNumber string) {
	pgm.magicNumber = magicNumber
}

func (pgm *PGM) SetMaxValue(maxValue uint8) {
	for y, _ := range pgm.data {
		for x, _ := range pgm.data[y] {
			prevvalue := pgm.data[y][x]
			newvalue := prevvalue * uint8(5) / pgm.max
			pgm.data[y][x] = newvalue
		}
	}
	pgm.max = maxValue
}

func (pgm *PGM) Rotate90CW() {
	rotatedData := make([][]uint8, pgm.width)
	for i := range rotatedData {
		rotatedData[i] = make([]uint8, pgm.height)
	}

	for i := 0; i < pgm.width; i++ {
		for j := 0; j < pgm.height; j++ {
			rotatedData[i][j] = pgm.data[pgm.height-1-j][i]
		}
	}

	pgm.width, pgm.height = pgm.height, pgm.width
	pgm.data = rotatedData
}

func (pgm *PGM) ToPBM() *PBM {
	pbm := &PBM{
		magicNumber: "P1",
		height:      pgm.height,
		width:       pgm.width,
		data:        make([][]bool, pgm.height),
	}

	for y, row := range pgm.data {
		pbm.data[y] = make([]bool, pgm.width)
		for x, grayValue := range row {
			isBlack := grayValue < pgm.max/2
			pbm.data[y][x] = isBlack
		}
	}

	return pbm
}
