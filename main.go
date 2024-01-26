package Netpbm

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
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

//////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////

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

//////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////

type PPM struct {
	data        [][]Pixel
	width       int
	height      int
	magicNumber string
	max         uint8
}

type Pixel struct {
	R, G, B uint8
}

func ReadPPM(filename string) (*PPM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	ppm := &PPM{}
	line := 0

	for scanner.Scan() {
		text := scanner.Text()

		if strings.HasPrefix(text, "#") {
			continue
		}

		if ppm.magicNumber == "" {
			ppm.magicNumber = strings.TrimSpace(text)
		} else if ppm.width == 0 {
			fmt.Sscanf(text, "%d %d", &ppm.width, &ppm.height)
			ppm.data = make([][]Pixel, ppm.height)
			for i := range ppm.data {
				ppm.data[i] = make([]Pixel, ppm.width)
			}
		} else if ppm.max == 0 {
			fmt.Sscanf(text, "%d", &ppm.max)
		} else {
			if ppm.magicNumber == "P3" {
				val := strings.Fields(text)
				for i := 0; i < ppm.width; i++ {
					r, _ := strconv.ParseUint(val[i*3], 10, 8)
					g, _ := strconv.ParseUint(val[i*3+1], 10, 8)
					b, _ := strconv.ParseUint(val[i*3+2], 10, 8)
					ppm.data[line][i] = Pixel{R: uint8(r), G: uint8(g), B: uint8(b)}
				}
				line++
			} else if ppm.magicNumber == "P6" {
				pixelData := make([]byte, ppm.width*ppm.height*3)
				fileContent, err := os.ReadFile(filename)
				if err != nil {
					return nil, fmt.Errorf("couldn't read file: %v", err)
				}
				copy(pixelData, fileContent[len(fileContent)-(ppm.width*ppm.height*3):])
				pixelIndex := 0
				for y := 0; y < ppm.height; y++ {
					for x := 0; x < ppm.width; x++ {
						ppm.data[y][x].R = pixelData[pixelIndex]
						ppm.data[y][x].G = pixelData[pixelIndex+1]
						ppm.data[y][x].B = pixelData[pixelIndex+2]
						pixelIndex += 3
					}
				}
				break
			}
		}
	}

	return ppm, nil
}

func (ppm *PPM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	fmt.Fprint(writer, ppm.magicNumber+"\n")
	fmt.Fprintf(writer, "%d %d\n", ppm.width, ppm.height)
	fmt.Fprintf(writer, "%d\n", ppm.max)
	writer.Flush()

	if ppm.magicNumber == "P3" {
		for y, row := range ppm.data {
			for i, pixel := range row {
				xtra := " "
				if i == len(row)-1 {
					xtra = ""
				}
				fmt.Fprintf(writer, "%d %d %d%s", pixel.R, pixel.G, pixel.B, xtra)
			}
			if y != len(ppm.data)-1 {
				fmt.Fprintln(writer, "")
			}
		}
		writer.Flush()
	} else if ppm.magicNumber == "P6" {
		for _, row := range ppm.data {
			for _, pixel := range row {
				_, err = file.Write([]byte{pixel.R, pixel.G, pixel.B})
				if err != nil {
					return fmt.Errorf("error writing pixel data: %v", err)
				}
			}
		}
	}

	return nil
}

func (ppm *PPM) Size() (int, int) {
	return ppm.width, ppm.height
}

func (ppm *PPM) At(x, y int) Pixel {
	return ppm.data[y][x]
}

func (ppm *PPM) Set(x, y int, value Pixel) {
	ppm.data[y][x] = value
}

func (ppm *PPM) Invert() {
	for y, row := range ppm.data {
		for x := range row {
			ppm.data[y][x].R = 255 - ppm.data[y][x].R
			ppm.data[y][x].G = 255 - ppm.data[y][x].G
			ppm.data[y][x].B = 255 - ppm.data[y][x].B
		}
	}
}

func (ppm *PPM) Flip() {
	for _, row := range ppm.data {
		cursor := ppm.width - 1

		for x := 0; x < ppm.width/2; x++ {
			row[x], row[cursor] = row[cursor], row[x]
			cursor--
		}
	}
}

func (ppm *PPM) Flop() {
	cursor := ppm.height - 1
	for y := 0; y < ppm.height/2; y++ {
		ppm.data[y], ppm.data[cursor] = ppm.data[cursor], ppm.data[y]
		cursor--
	}
}

func (ppm *PPM) SetMagicNumber(magicNumber string) {
	ppm.magicNumber = magicNumber
}

func (ppm *PPM) SetMaxValue(maxValue uint8) {
	for y := range ppm.data {
		for x := range ppm.data[y] {
			pixel := ppm.data[y][x]
			pixel.R = uint8(float64(pixel.R) * float64(maxValue) / float64(ppm.max))
			pixel.G = uint8(float64(pixel.G) * float64(maxValue) / float64(ppm.max))
			pixel.B = uint8(float64(pixel.B) * float64(maxValue) / float64(ppm.max))
			ppm.data[y][x] = pixel
		}
	}
	ppm.max = maxValue
}

func (ppm *PPM) Rotate90CW() {
	rotatedData := make([][]Pixel, ppm.width)
	for i := range rotatedData {
		rotatedData[i] = make([]Pixel, ppm.height)
	}
	for i := 0; i < ppm.width; i++ {
		for j := 0; j < ppm.height; j++ {
			rotatedData[i][j] = ppm.data[ppm.height-1-j][i]
		}
	}
	ppm.width, ppm.height = ppm.height, ppm.width
	ppm.data = rotatedData
}

func (ppm *PPM) ToPBM() *PBM {
	pbm := &PBM{
		magicNumber: "P1",
		height:      ppm.height,
		width:       ppm.width,
	}
	for y := range ppm.data {
		pbm.data = append(pbm.data, []bool{})
		for x := range ppm.data[y] {
			r, g, b := ppm.data[y][x].R, ppm.data[y][x].G, ppm.data[y][x].B
			isBlack := (uint8((int(r)+int(g)+int(b))/3) < ppm.max/2)
			pbm.data[y] = append(pbm.data[y], isBlack)
		}
	}
	return pbm
}

func (ppm *PPM) ToPGM() *PGM {
	pgm := &PGM{
		magicNumber: "P2",
		height:      ppm.height,
		width:       ppm.width,
		max:         ppm.max,
	}
	for y := range ppm.data {
		pgm.data = append(pgm.data, []uint8{})
		for x := range ppm.data[y] {
			r, g, b := ppm.data[y][x].R, ppm.data[y][x].G, ppm.data[y][x].B
			grayValue := uint8((int(r) + int(g) + int(b)) / 3)
			pgm.data[y] = append(pgm.data[y], grayValue)
		}
	}
	return pgm
}

type Point struct {
	X, Y int
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func sign(x int) int {
	if x > 0 {
		return 1
	} else if x < 0 {
		return -1
	}
	return 0
}

func isWithinBounds(x, y, width, height int) bool {
	return x >= 0 && x < width && y >= 0 && y < height
}

func (ppm *PPM) DrawLine(p1, p2 Point, color Pixel) {
	deltaX := abs(p2.X - p1.X)
	deltaY := abs(p2.Y - p1.Y)
	sx, sy := sign(p2.X-p1.X), sign(p2.Y-p1.Y)
	err := deltaX - deltaY

	for {
		if isWithinBounds(p1.X, p1.Y, ppm.width, ppm.height) {
			ppm.data[p1.Y][p1.X] = color
		}

		if p1.X == p2.X && p1.Y == p2.Y {
			break
		}

		e2 := 2 * err

		if e2 > -deltaY {
			err -= deltaY
			p1.X += sx
		}

		if e2 < deltaX {
			err += deltaX
			p1.Y += sy
		}
	}
}

func (ppm *PPM) DrawRectangle(p1 Point, width, height int, color Pixel) {
	p2 := Point{p1.X + width, p1.Y}
	p3 := Point{p1.X, p1.Y + height}
	p4 := Point{p1.X + width, p1.Y + height}

	ppm.DrawLine(p1, p2, color)
	ppm.DrawLine(p2, p4, color)
	ppm.DrawLine(p4, p3, color)
	ppm.DrawLine(p3, p1, color)
}

func (ppm *PPM) DrawFilledRectangle(p1 Point, width, height int, color Pixel) {
	p2 := Point{p1.X + width, p1.Y}

	for i := 0; i <= height; i++ {
		ppm.DrawLine(p1, p2, color)
		p1.Y++
		p2.Y++
	}
}

func (ppm *PPM) DrawCircle(center Point, radius int, color Pixel) {
	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			dx := float64(x - center.X)
			dy := float64(y - center.Y)
			distance := math.Sqrt(dx*dx + dy*dy)

			if math.Abs(distance-float64(radius)*0.85) < 0.5 {
				ppm.data[y][x] = color
			}
		}
	}
}

func (ppm *PPM) DrawFilledCircle(center Point, radius int, color Pixel) {
	for radius >= 0 {
		ppm.DrawCircle(center, radius, color)
		radius--
	}
}

func (ppm *PPM) DrawTriangle(p1, p2, p3 Point, color Pixel) {
	ppm.DrawLine(p1, p2, color)
	ppm.DrawLine(p2, p3, color)
	ppm.DrawLine(p3, p1, color)
}

func (ppm *PPM) DrawFilledTriangle(p1, p2, p3 Point, color Pixel) {
	for p1 != p2 {
		ppm.DrawLine(p3, p1, color)

		if p1.X != p2.X {
			p1.X += sign(p2.X - p1.X)
		}

		if p1.Y != p2.Y {
			p1.Y += sign(p2.Y - p1.Y)
		}
	}

	ppm.DrawLine(p3, p1, color)
}

func (ppm *PPM) DrawPolygon(points []Point, color Pixel) {
	for i := 0; i < len(points)-1; i++ {
		ppm.DrawLine(points[i], points[i+1], color)
	}

	ppm.DrawLine(points[len(points)-1], points[0], color)
}

func (ppm *PPM) DrawFilledPolygon(points []Point, color Pixel) {
	if len(points) < 3 {
		return
	}

	minY, maxY := boundingBoxY(points)

	for y := minY; y <= maxY; y++ {
		intersections := findIntersections(points, y)

		sort.Sort(intersectionSlice(intersections))

		for i := 0; i < len(intersections); i += 2 {
			x1 := intersections[i]
			x2 := intersections[i+1]

			for x := x1; x <= x2; x++ {
				ppm.data[y][x] = color
			}
		}
	}
}

func boundingBoxY(points []Point) (minY, maxY int) {
	if len(points) == 0 {
		return 0, 0
	}

	minY, maxY = points[0].Y, points[0].Y
	for _, p := range points {
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	return minY, maxY
}

func boundingBox(points []Point) (int, int, int, int) {
	minX, minY, maxX, maxY := points[0].X, points[0].Y, points[0].X, points[0].Y

	for _, point := range points {
		if point.X < minX {
			minX = point.X
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}

	return minX, minY, maxX, maxY
}

func findIntersections(points []Point, y int) []int {
	intersections := make([]int, 0)

	for i := 0; i < len(points); i++ {
		j := (i + 1) % len(points)

		y1, y2 := points[i].Y, points[j].Y

		if (y1 <= y && y < y2) || (y2 <= y && y < y1) {
			x := int(float64(y-y1)*(float64(points[j].X-points[i].X)/float64(y2-y1))) + points[i].X
			intersections = append(intersections, x)
		}
	}

	return intersections
}

type intersectionSlice []int

func (s intersectionSlice) Len() int           { return len(s) }
func (s intersectionSlice) Less(i, j int) bool { return s[i] < s[j] }
func (s intersectionSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (ppm *PPM) DrawKochSnowflake(n int, start Point, width int, color Pixel) {
	var drawKoch func(level int, p1, p2 Point)

	p1 := start
	p2 := Point{X: start.X + width, Y: start.Y}

	drawKoch = func(level int, p1, p2 Point) {
		if level == 0 {
			ppm.DrawLine(p1, p2, color)
		} else {
			dx := p2.X - p1.X
			dy := p2.Y - p1.Y

			p1_3 := Point{X: p1.X + dx/3, Y: p1.Y + dy/3}
			p2_3 := Point{X: p1.X + 2*dx/3, Y: p1.Y + 2*dy/3}

			angle := math.Pi / 3.0
			cosAngle := math.Cos(angle)
			sinAngle := math.Sin(angle)
			pTriangle := Point{
				X: int(float64(p1_3.X-p2_3.X)*cosAngle-float64(p1_3.Y-p2_3.Y)*sinAngle) + p2_3.X,
				Y: int(float64(p1_3.X-p2_3.X)*sinAngle+float64(p1_3.Y-p2_3.Y)*cosAngle) + p2_3.Y,
			}

			drawKoch(level-1, p1, p1_3)
			drawKoch(level-1, p1_3, pTriangle)
			drawKoch(level-1, pTriangle, p2_3)
			drawKoch(level-1, p2_3, p2)
		}
	}

	drawKoch(n, p1, p2)
}

func (ppm *PPM) DrawSierpinskiTriangle(n int, start Point, width int, color Pixel) {
	var drawSierpinski func(level int, p1, p2, p3 Point)

	height := int(float64(width) * math.Sqrt(3) / 2)
	p1 := start
	p2 := Point{X: start.X + width, Y: start.Y}
	p3 := Point{X: start.X + width/2, Y: start.Y + height}

	drawSierpinski = func(level int, p1, p2, p3 Point) {
		if level == 0 {
			ppm.DrawFilledTriangle(p1, p2, p3, color)
		} else {
			pMiddle1 := Point{X: (p1.X + p2.X) / 2, Y: (p1.Y + p2.Y) / 2}
			pMiddle2 := Point{X: (p2.X + p3.X) / 2, Y: (p2.Y + p3.Y) / 2}
			pMiddle3 := Point{X: (p3.X + p1.X) / 2, Y: (p3.Y + p1.Y) / 2}

			drawSierpinski(level-1, p1, pMiddle1, pMiddle3)
			drawSierpinski(level-1, pMiddle1, p2, pMiddle2)
			drawSierpinski(level-1, pMiddle3, pMiddle2, p3)
		}
	}

	drawSierpinski(n, p1, p2, p3)
}

func (ppm *PPM) DrawPerlinNoise(color1 Pixel, color2 Pixel) {
	scale := 0.1
	octaves := 4
	persistence := 0.5

	for y := 0; y < ppm.height; y++ {
		for x := 0; x < ppm.width; x++ {
			perlinValue := perlinNoise(float64(x)*scale, float64(y)*scale, octaves, persistence)

			lerpColor := lerpColor(color1, color2, perlinValue)

			ppm.data[y][x] = lerpColor
		}
	}
}

func lerpColor(color1 Pixel, color2 Pixel, t float64) Pixel {
	lerpComponent := func(c1, c2 uint8, t float64) uint8 {
		return uint8(float64(c1)*(1.0-t) + float64(c2)*t)
	}

	return Pixel{
		R: lerpComponent(color1.R, color2.R, t),
		G: lerpComponent(color1.G, color2.G, t),
		B: lerpComponent(color1.B, color2.B, t),
	}
}

func perlinNoise(x, y float64, octaves int, persistence float64) float64 {
	total := 0.0
	frequency := 1.0
	amplitude := 1.0
	maxValue := 0.0

	for i := 0; i < octaves; i++ {
		total += noise(x*frequency, y*frequency) * amplitude

		maxValue += amplitude
		amplitude *= persistence
		frequency *= 2
	}

	return total / maxValue
}

func noise(x, y float64) float64 {
	n := int64(x) + int64(y)*57
	n = (n << 13) ^ n
	return (1.0 - float64((n*(n*n*15731+789221)+1376312589)&0x7fffffff)/1073741824.0)
}

func (ppm *PPM) KNearestNeighbors(newWidth, newHeight int) {
	scaleX := float64(ppm.width) / float64(newWidth)
	scaleY := float64(ppm.height) / float64(newHeight)

	resizedData := make([][]Pixel, newHeight)
	for i := range resizedData {
		resizedData[i] = make([]Pixel, newWidth)
	}

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			originalX := int(float64(x) * scaleX)
			originalY := int(float64(y) * scaleY)

			resizedData[y][x] = ppm.getNearestNeighbor(originalX, originalY)
		}
	}

	ppm.width = newWidth
	ppm.height = newHeight
	ppm.data = resizedData
}

func (ppm *PPM) getNearestNeighbor(x, y int) Pixel {
	x = clamp(x, 0, ppm.width-1)
	y = clamp(y, 0, ppm.height-1)

	return ppm.data[y][x]
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}
