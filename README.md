# eff
command line tool to extract info from text files

arg -f : cardinality (1) - file ou dir to search, may contem special char * . ex: C:\temp\file.txt, C:\temp\*.txt
arg -o : cardinality (0-1) - output file"
arg -findp : cardinality (0*) - extract from positional text file. format = l:p:s:h = get data at line beginning by #l (l is opcional), position #p, size #s, data name #h. p=>1, s=>1
arg -findd : cardinality (0*) - extract from delimited text file. format = l:n:h = get data at line beginning by #l (l is opcional), n-th element #n, data name #h. n=>1
arg -d : cardinality (0-1) - delimiter when extracting from delimited files (use with arg -findd)
arg -p : cardinality (0-1) - activate parallel extractions (files quantity must be greater than 3)

examples : 
eff.exe -f "C:\temp\filedelim\arq delim*.txt" -d ";" -findd H:2:NAME -findd H:4:"FILE SEQ" -findd D:2:NAME -findd D:4:COUNTRY -findd D:5:REGION -findd D:6:TELEPHONE -findd D:7:"ITEM NUMBER" -o "c:\temp\arq delim.txt"

eff.exe -f "c:\temp\filepos\arq pos*.txt" -findp D:2:20:NAME -findp D:24:20:COUNTRY -findp D:64:15:TELEPHONE -findp D:89:6:"ITEM NUMBER" -o "c:\temp\arq pos.txt" -p
