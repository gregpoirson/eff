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
.\eff.exe -f C:\Temp\CAIXA\GPF.MZ.BBA2.ICTPJUNC.D2107* -lb 1 -ps 38:40 -ps 99:2 > c:\temp\find.txt
.\eff.exe -f C:\Temp\CAIXA\GPF.MZ.BBA2.ICTPJUNC.D2107* -lb 1 -ps 2:14 -hdr PROPOSTA -ps 38:40 -hdr "NOME CLIENTE" -ps 99:2 -hdr UF > c:\temp\find.txt
.\eff.exe -f C:\Temp\CAIXA\GPF.MZ.BBA2.ICTPJUNC.D2107* -lb 1 -ps 16:14 -hdr CPF -ps 106:3 -hdr DD1 -ps 109:9 -hdr TEL1 -ps 118:3 -hdr DD2 -ps 121:9 -hdr TEL2 -ps 130:3 -hdr DD3 -ps 133:9 -hdr TEL3 > c:\temp\find.txt

.\eff.exe -f C:\Temp\CAIXA\GPF.MZ.BBA2.ICTPPEND.D2107* -find 1:2:14:PROPOSTA -find 1:38:40:"NOME CLIENTE" -find 3:276:2:CANAL
go run . -f C:\Temp\CAIXA\GPF.MZ.BBA2.ICTPPEND.D2107* -find 1:2:14:PROPOSTA -find 1:38:40:"NOME CLIENTE" -find 3:276:2:CANAL
go run . -f C:\Temp\CAIXA\GPF.MZ.BBA2.ICTPJUNC.D2107* -find 1:16:14:CPF -find 1:106:3:DD1 -find 1:109:9:TEL1 -find 1:118:3:DD2 -find 1:121:9:TEL2 -find 1:130:3:DD3 -find 1:133:9:TEL3 > c:\temp\find.txt

go run . -f c:\Temp\Caixa\GPF.MZ.BBA2.ICTPRVFN.D* -find :2:14:PROPOSTA -find 3:95:6:"ADESAO ?" > c:\temp\find.txt
*/

type Find struct {
	lineBegin string
	pos       int
	size      int
	nome      string
}

type arrayFind []Find

var flagFile string
var flagFind arrayFind

func init() {
	flag.StringVar(&flagFile, "f", "", `(1) - file ou dir to search, may contem special char *. ex: C:\temp\file.txt, C:\temp\*.txt`)
	flag.Var(&flagFind, "find", "(1*) - format = l:p:s:h = get data at line beginning by #l (l pode estar vazio), position #p, size #s, nome dado #h. p=>1, s=>1")
}

func (i *arrayFind) String() string {
	return fmt.Sprint(*i)
}

func (i *arrayFind) Set(value string) error {

	// if len(*i) > 0 {
	//           return errors.New("Find flag already set")
	// }
	tok := strings.Split(value, ":")
	if len(tok) != 4 {
		return errors.New("o formato do flag é l:p:s:h")
	}

	var novaVal Find
	for idx, val := range tok {
		if idx == 0 {
			novaVal.lineBegin = val
		} else if idx == 1 {
			novaVal.pos, _ = strconv.Atoi(val)
		} else if idx == 2 {
			novaVal.size, _ = strconv.Atoi(val)
		} else if idx == 3 {
			novaVal.nome = val
		}
	}
	if novaVal.pos < 1 {
		return errors.New("as positions são superiores a 1")
	}
	if novaVal.size < 1 {
		return errors.New("as sizes são superiores a 1")
	}
	//as posições na linha de comando começam com 1, porem os metodos Go para procurar começam por 0
	novaVal.pos--
	*i = append(*i, novaVal)

	//fmt.Println(novaVal)

	return nil
}

func main() {

	flag.Parse()

	if flagFile == "" {
		fmt.Println("Informa o flag -f para procurar em um arquivo ou caminho.")

		flag.PrintDefaults()
		return
	}

	if len(flagFind) == 0 {
		fmt.Println("Informa o flag -find para procurar algo nos arquivos.")

		flag.PrintDefaults()
		return
	}

	start := time.Now()

	FindInFiles(flagFile)

	elapsed := time.Since(start)
	fmt.Printf("Tempo Execução\t%s\n", elapsed)
}

func FindInFiles(path string) {

	allFiles := getAllFilesInDir(path)

	// for _, f := range allFiles {
	//           fmt.Println(f)
	// }
	fmt.Printf("Procurando em %d arquivos.\r\n", len(allFiles))

	qtd := 0
	for _, arq := range allFiles {

		if arq == "" {
			continue
		}

		qtd++
		fmt.Printf("Arquivo #%d\t%s\r\n", qtd, arq)

		lines, err := ExtactData(arq)

		if err == nil && len(lines) > 0 {
			for _, ln := range lines {
				fmt.Println(ln)
			}
		}
	}

	fmt.Printf("Arquivos processados com sucesso\t%d\r\n", qtd)
}

func getAllFilesInDir(path string) []string {
	allFiles, _ := filepath.Glob(path)

	return allFiles
}

func ExtactData(path string) ([]string, error) {
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
	for _, lps := range flagFind {
		fmt.Printf("\t%s:%s", lps.lineBegin, lps.nome)
	}
	fmt.Println("")

	for scanner.Scan() {
		numLinha++

		linha := scanner.Text()

		var dados []string
		found = false

		for _, lps := range flagFind {
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
			fmt.Printf("%d\t%s\r\n", numLinha, strings.Join(dados, "\t"))
		}
	}
	fmt.Printf("Total Lines\t%d\r\n", total)

	if err := scanner.Err(); err != nil {
		return allLines, err
	}

	return allLines, nil
}
