package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

/*
 -f C:\Temp\GPF.MZ.BBA2.ICTPPEND.D2107* -find 1:2:14:PROPOSTA -find 1:38:40:"NOME CLIENTE" -find 3:276:2:CANAL
 -f C:\Temp\GPF.MZ.BBA2.ICTPPEND.D2107* -find 1:2:14:PROPOSTA -find 1:38:40:"NOME CLIENTE" -find 3:276:2:CANAL
 -f C:\Temp\GPF.MZ.BBA2.ICTPJUNC.D2107* -find 1:16:14:CPF -find 1:106:3:DD1 -find 1:109:9:TEL1 -find 1:118:3:DD2 -find 1:121:9:TEL2 -find 1:130:3:DD3 -find 1:133:9:TEL3 > c:\temp\find.txt
 -f c:\Temp\GPF.MZ.BBA2.ICTPRVFN.D* -find :2:14:PROPOSTA -find 3:95:6:"ADESAO ?" > c:\temp\find.txt

 -f ".\file_delimiter.txt" -d ";" -findd H:2:NAME -findd D:2:NAME -findd T:2:QUANTITY

*/

type FindPosition struct {
	lineBegin string
	pos       int
	size      int
	name      string
}

type FindDelimiter struct {
	lineBegin string
	nth       int
	name      string
}

type arrayFindP []FindPosition
type arrayFindD []FindDelimiter

var flagFile string
var flagFindP arrayFindP
var flagFindD arrayFindD
var flagDelimiter string

func init() {
	flag.StringVar(&flagFile, "f", "", `cardinality (1) - file ou dir to search, may contem special char *. ex: C:\temp\file.txt, C:\temp\*.txt`)

	flag.Var(&flagFindP, "findp", "extract from positional text file. cardinality (0*) - format = l:p:s:h = get data at line beginning by #l (l is opcional), position #p, size #s, data name #h. p=>1, s=>1")

	flag.Var(&flagFindD, "findd", "extract from delimited text file. cardinality (0*) - format = l:n:h = get data at line beginning by #l (l is opcional), n-th element #n, data name #h. n=>1")
	flag.StringVar(&flagDelimiter, "d", "", `cardinality (0-1) - delimiter when extracting from delimited files`)
}

func (i *arrayFindP) String() string {
	return fmt.Sprint(*i)
}

func (i *arrayFindP) Set(value string) error {

	tok := strings.Split(value, ":")
	if len(tok) != 4 {
		return errors.New("flag format is l:p:s:h")
	}

	var novaVal FindPosition
	for idx, val := range tok {
		if idx == 0 {
			novaVal.lineBegin = val
		} else if idx == 1 {
			novaVal.pos, _ = strconv.Atoi(val)
		} else if idx == 2 {
			novaVal.size, _ = strconv.Atoi(val)
		} else if idx == 3 {
			novaVal.name = val
		}
	}
	if novaVal.pos < 1 {
		return errors.New("position in flag must be equal ou greater than 1")
	}
	if novaVal.size < 1 {
		return errors.New("size in flag must be equal ou greater than 1")
	}
	//as posições na linha de comando começam com 1, porem os metodos Go para procurar começam por 0
	novaVal.pos--
	*i = append(*i, novaVal)

	//fmt.Println(novaVal)

	return nil
}

func (i *arrayFindD) String() string {
	return fmt.Sprint(*i)
}

func (i *arrayFindD) Set(value string) error {

	tok := strings.Split(value, ":")
	if len(tok) != 3 {
		return errors.New("flag format is l:n:h")
	}

	var novaVal FindDelimiter
	for idx, val := range tok {
		if idx == 0 {
			novaVal.lineBegin = val
		} else if idx == 1 {
			novaVal.nth, _ = strconv.Atoi(val)
		} else if idx == 2 {
			novaVal.name = val
		}
	}
	if novaVal.nth < 1 {
		return errors.New("element number in flag must be equal ou greater than 1")
	}

	//indexes in flag begin with 1, but begins with 0 in Go
	novaVal.nth--
	*i = append(*i, novaVal)

	//fmt.Println(novaVal)

	return nil
}

func main() {

	flag.Parse()

	if flagFile == "" {
		fmt.Println("flag -f is mandatory")

		flag.PrintDefaults()
		return
	}

	if len(flagFindP) == 0 && len(flagFindD) == 0 {
		fmt.Println("find flags not found. either use findp or findd")

		flag.PrintDefaults()
		return
	}

	if len(flagFindP) > 0 && len(flagFindD) > 0 {
		fmt.Println("both find flags found. either use findp or findd")

		flag.PrintDefaults()
		return
	}

	if len(flagFindD) > 0 && flagDelimiter == "" {
		fmt.Println("flag delimiter is mandatory when using findd")

		flag.PrintDefaults()
		return
	}

	Process()
}

func Process() {
	start := time.Now()

	//choose the apropriate Extract function
	var fnExtract func(path string) ([]string, error) = nil
	if len(flagFindP) > 0 {
		fnExtract = ExtactDataPosition
	} else if len(flagFindD) > 0 {
		fnExtract = ExtactDataDelimiter
	}

	FindInFiles(flagFile, fnExtract)

	elapsed := time.Since(start)
	fmt.Printf("duration\t%s\n", elapsed)
}

func FindInFiles(path string, fnExtract func(path string) ([]string, error)) {

	allFiles := getAllFilesInDir(path)

	fmt.Printf("searching %d file(s)\n", len(allFiles))

	//fileswithData := 0
	for idx, arq := range allFiles {

		fmt.Printf("file #%d\t%s\n", idx+1, arq)

		fnExtract(arq)
		//lines, err := fnExtract(arq)

		// if err == nil && len(lines) > 0 {
		// 	fileswithData++
		// 	for _, ln := range lines {
		// 		fmt.Println(ln)
		// 	}
		// }
	}

	//fmt.Printf("extraction over. files with data\t%d\n", fileswithData)
}

func getAllFilesInDir(path string) []string {
	allFiles, _ := filepath.Glob(path)

	return allFiles
}

func ExtactDataPosition(path string) ([]string, error) {
	allLines := make([]string, 0)

	f, err := os.Open(path)
	if err != nil {
		return allLines, err
	}
	defer f.Close()

	//abre um scan
	scanner := bufio.NewScanner(f)

	numLinha := 0
	tipo := ""
	dado := ""
	found := false
	total := 0

	//header
	fmt.Print("Line")
	for _, lps := range flagFindP {
		fmt.Printf("\t%s:%s", lps.lineBegin, lps.name)
	}
	fmt.Println("")

	for scanner.Scan() {
		numLinha++

		linha := scanner.Text()

		var dados []string
		found = false

		for _, lps := range flagFindP {
			//search for the lines beginning with l
			if len(lps.lineBegin) > 0 {
				tipo = linha[:len(lps.lineBegin)]

				if tipo != lps.lineBegin {
					continue
				}
			}

			found = true
			dado = linha[lps.pos : lps.pos+lps.size]
			//dados = append(dados, fmt.Sprintf("%d\t%d\t%s", ps.pos, ps.size, dado))
			dados = append(dados, dado)
			//fmt.Printf("Line %d, pos %d size %d: %s", numLinha, ps.pos, ps.size, dado)
		}
		if found {
			total++
			fmt.Printf("%d\t%s\n", numLinha, strings.Join(dados, "\t"))
		}
	}
	fmt.Printf("total lines\t%d\n", total)

	if err := scanner.Err(); err != nil {
		return []string{}, err
	}

	return allLines, nil
}

func ExtactDataDelimiter(path string) ([]string, error) {
	allLines := make([]string, 0)

	f, err := os.Open(path)
	if err != nil {
		return allLines, err
	}
	defer f.Close()

	//abre um scan
	scanner := bufio.NewScanner(f)

	numLinha := 0
	tipo := ""
	dado := ""
	found := false
	total := 0

	//header
	fmt.Print("Line")
	for _, lps := range flagFindD {
		fmt.Printf("\t%s:%s", lps.lineBegin, lps.name)
	}
	fmt.Println("")

	for scanner.Scan() {
		numLinha++

		linha := scanner.Text()

		var dados []string
		found = false

		for _, ln := range flagFindD {
			//search for the lines beginning with l
			if len(ln.lineBegin) > 0 {
				tipo = linha[:len(ln.lineBegin)]

				if tipo != ln.lineBegin {
					continue
				}
			}

			//check if there is at least n elements in the line
			tok := strings.Split(linha, flagDelimiter)
			if ln.nth > len(tok) {
				//"number of elements in the line is inferior a n-th"
				fmt.Printf("number of elements in the line is inferior a n-th. %v\n", ln)
				found = false
				continue
			}

			found = true
			dado = tok[ln.nth]
			//dados = append(dados, fmt.Sprintf("%d\t%d\t%s", ps.pos, ps.size, dado))
			dados = append(dados, dado)
			//fmt.Printf("Line %d, pos %d size %d: %s", numLinha, ps.pos, ps.size, dado)
		}
		if found {
			total++
			fmt.Printf("%d\t%s\n", numLinha, strings.Join(dados, "\t"))
		}
	}
	fmt.Printf("total lines\t%d\n", total)

	if err := scanner.Err(); err != nil {
		return allLines, err
	}

	return allLines, nil
}
