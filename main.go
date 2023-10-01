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
 -f ".\file_delimiter.txt" -d ";" -findd H:2:NAME -findd D:2:NAME -findd T:2:QUANTITY
 -f ".\file_position.txt" -findp H:2:8:DATE -findp D:2:20:NAME -findp D:22:2:AGE
 -f ".\file_position.txt" -findp H:2:8:DATE -findp D:2:20:NAME -findp D:24:20:COUNTRY -findp D:44:15:TELEPHONE
 -f ".\file_position.txt" -findp H:2:8:DATE -findp D:2:20:NAME -findp D:24:20:COUNTRY -findp D:44:15:TELEPHONE -o c:\temp\findp.txt
 -f ".\file_delimiter.txt" -d ";" -findd H:2:NAME -findd D:2:NAME -findd D:4:COUNTRY -findd D:5:REGION -o c:\temp\findd.txt

 -f "c:\temp\filepos\arq pos*.txt" -findp H:2:8:DATE -findp H:17:6:SEQUENTIAL
 -f "c:\temp\filepos\arq pos*.txt" -findp D:2:20:NAME -findp D:22:2:AGE -findp D:64:15:TELEPHONE -findp D:89:6:"ITEM NUMBER"
 -f "c:\temp\filepos\arq pos*.txt" -findp D:2:20:NAME -findp D:24:20:COUNTRY -findp D:64:15:TELEPHONE -findp D:89:6:"ITEM NUMBER" -o "c:\temp\arq pos.txt"
 -f "C:\temp\filedelim\arq delim*.txt" -d ";" -findd H:2:NAME -findd H:4:"FILE SEQ" -findd D:2:NAME -findd D:4:COUNTRY -findd D:5:REGION -findd D:6:TELEPHONE -findd D:7:"ITEM NUMBER" -o "c:\temp\arq delim.txt"
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
var flagOutputFile string

type fnExtract func(path string) ([]string, error)

func init() {
	flag.StringVar(&flagFile, "f", "", `cardinality (1) - file ou dir to search, may contem special char *. ex: C:\temp\file.txt, C:\temp\*.txt`)
	flag.StringVar(&flagOutputFile, "o", "", "cardinality (0-1) - output file")

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
	var fnExtract = getExtractFunction()
	if fnExtract == nil {
		panic("No extraction function defined for these flags")
	}

	FindInFiles(flagFile, fnExtract)

	elapsed := time.Since(start)
	fmt.Printf("duration\t%s\n", elapsed)
}

func getExtractFunction() fnExtract {
	//choose the apropriate Extract function
	if len(flagFindP) > 0 {
		return ExtactDataPosition
	}
	if len(flagFindD) > 0 {
		return ExtactDataDelimiter
	}
	return nil
}

func FindInFiles(path string, fnExtract fnExtract) {

	allFiles := getAllFilesInDir(path)

	fmt.Printf("searching %d file(s)\n", len(allFiles))

	allLines := make([]string, 0)

	for idx, arq := range allFiles {

		//fmt.Printf("file #%d\t%s\n", idx+1, arq)
		allLines = append(allLines, fmt.Sprintf("file #%d\t%s", idx+1, arq))

		lines, err := fnExtract(arq)

		if err != nil {
			panic(err)
		}

		allLines = append(allLines, lines...)

		//fmt.Printf("total lines\t%d\n", len(lines))
	}

	//write the lines in the output file if informed, or in the terminal
	if flagOutputFile != "" {
		writeOutputFile(allLines, flagOutputFile)
	} else {
		for _, l := range allLines {
			fmt.Println(l)
		}
	}
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
	var hdr = "Line"
	for _, lps := range flagFindP {
		hdr = fmt.Sprintf("%s\t%s:%s", hdr, lps.lineBegin, lps.name)
	}
	allLines = append(allLines, hdr)

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
			//dado = strUtf8(linha, lps)
			dado = substr(linha, lps)
			//fmt.Println(dado)
			dados = append(dados, dado)
		}
		if found {
			total++
			var out = fmt.Sprintf("%d\t%s", numLinha, strings.Join(dados, "\t"))
			allLines = append(allLines, out)
		}
	}

	if err := scanner.Err(); err != nil {
		return []string{}, err
	}

	return allLines, nil
}

// go reads in bytes, not in char. must 'convert' the string
// in Go the indices are byte-indices, not character or rune indices. Go stores the UTF-8 encoded byte sequence of texts in a string
func substr(s string, lps FindPosition) string {
	counter, startIdx := 0, 0
	for i := range s {
		if counter == lps.pos {
			startIdx = i
		}
		if counter == lps.pos+lps.size {
			return s[startIdx:i]
		}
		counter++
	}
	return s[startIdx:]
}

// func strUtf8(str string, lps FindPosition) string {
// 	return string([]rune(str)[lps.pos : lps.pos+lps.size])
// }

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
	var hdr = "Line"
	for _, lps := range flagFindD {
		hdr = fmt.Sprintf("%s\t%s:%s", hdr, lps.lineBegin, lps.name)
	}
	allLines = append(allLines, hdr)

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
			dados = append(dados, dado)
		}
		if found {
			total++
			var out = fmt.Sprintf("%d\t%s", numLinha, strings.Join(dados, "\t"))
			allLines = append(allLines, out)
		}
	}

	if err := scanner.Err(); err != nil {
		return allLines, err
	}

	return allLines, nil
}

func writeOutputFile(lines []string, outputFile string) {
	f, _ := os.Create(outputFile)
	defer f.Close()

	for _, l := range lines {
		fmt.Fprintln(f, l)
	}
}
