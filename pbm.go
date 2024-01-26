package Netpbm

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type PBM struct {
	data        [][]bool
	width       int
	height      int
	magicNumber string
}

func ReadPBM(filename string) (*PBM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	pbm := &PBM{}
	line := 0

	for scanner.Scan() {
		text := scanner.Text()

		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}

		if pbm.magicNumber == "" {
			pbm.magicNumber = strings.TrimSpace(text)
		} else if pbm.width == 0 {
			fmt.Sscanf(text, "%d %d", &pbm.width, &pbm.height)
			pbm.data = make([][]bool, pbm.height)
			for i := range pbm.data {
				pbm.data[i] = make([]bool, pbm.width)
			}
		} else {
			if pbm.magicNumber == "P1" {
				test := strings.Fields(text)
				for i := 0; i < pbm.width; i++ {
					pbm.data[line][i] = (test[i] == "1")
				}
				line++
			} else if pbm.magicNumber == "P4" {
				expectedBytesPerRow := (pbm.width + 7) / 8
				totalExpectedBytes := expectedBytesPerRow * pbm.height
				allPixelData := make([]byte, totalExpectedBytes)

				fileContent, err := os.ReadFile(filename)
				if err != nil {
					return nil, fmt.Errorf("couldn't read file: %v", err)
				}

				copy(allPixelData, fileContent[len(fileContent)-totalExpectedBytes:])

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
				break
			}
		}
	}

	return pbm, nil
}

func (pbm *PBM) Size() (int, int) {
	return pbm.width, pbm.height
}

func (pbm *PBM) At(x, y int) bool {
	return pbm.data[y][x]
}

func (pbm *PBM) Set(x, y int, value bool) {
	pbm.data[y][x] = value
}

func (pbm *PBM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	fmt.Fprint(writer, pbm.magicNumber+"\n")
	fmt.Fprintf(writer, "%d %d\n", pbm.width, pbm.height)
	writer.Flush()

	if pbm.magicNumber == "P1" {
		for y, row := range pbm.data {
			for i, pixel := range row {
				xtra := " "
				if i == len(row)-1 {
					xtra = ""
				}
				if pixel {
					fmt.Fprint(writer, "1"+xtra)
				} else {
					fmt.Fprint(writer, "0"+xtra)
				}
			}
			if y != len(pbm.data)-1 {
				fmt.Fprintln(writer, "")
			}
		}
		writer.Flush()
	} else if pbm.magicNumber == "P4" {
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

func (pbm *PBM) Invert() {
	for y, row := range pbm.data {
		for x := range row {
			pbm.data[y][x] = !pbm.data[y][x]
		}
	}
}

func (pbm *PBM) Flip() {
	for y, row := range pbm.data {
		cursor := pbm.width - 1
		for x := 0; x < pbm.width/2; x++ {
			temp := row[x]
			pbm.data[y][x] = row[cursor]
			pbm.data[y][cursor] = temp
			cursor--
		}
	}
}

func (pbm *PBM) Flop() {
	cursor := pbm.height - 1
	for y := 0; y < pbm.height/2; y++ {
		temp := pbm.data[y]
		pbm.data[y] = pbm.data[cursor]
		pbm.data[cursor] = temp
		cursor--
	}
}

func (pbm *PBM) SetMagicNumber(magicNumber string) {
	pbm.magicNumber = magicNumber
}
