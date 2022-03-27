package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Local struct {
	active bool
	data   int64
}

type ForthVM struct {
	mem    []int64
	stack  []int64
	n      int
	lstack []Local
	ln     int
	l_len  int
}

func NewForthVM() *ForthVM {
	fvm := new(ForthVM)

	fvm.mem = make([]int64, 1000)
	fvm.stack = make([]int64, 100)
	fvm.n = -1

	return fvm
}

func (fvm *ForthVM) push(i int64) {
	fvm.n += 1
	fvm.stack[fvm.n] = i
}

func (fvm *ForthVM) pop() int64 {
	v := fvm.stack[fvm.n]
	fvm.n -= 1
	return v
}

func (fvm *ForthVM) lctx() {
	fvm.ln += 1
	for i := 0; i < fvm.l_len; i++ {
		fvm.local_get(i, fvm.ln).active = false
	}
}

func (fvm *ForthVM) local_get(name, ctx int) *Local {
	return &fvm.lstack[fvm.l_len*ctx+name]
}

func (fvm *ForthVM) ldef(name int) {
	v := fvm.pop()
	local := fvm.local_get(name, fvm.ln)
	local.data = v
	local.active = true
}

func (fvm *ForthVM) lset(name int) {
	v := fvm.pop()
	for i := fvm.ln; i > -1; i-- {
		if local := fvm.local_get(name, i); local.active {
			local.data = v
			return
		}
	}
}

func (fvm *ForthVM) lcl(name int) {
	for i := fvm.ln; i > -1; i-- {
		if local := fvm.local_get(name, i); local.active {
			fvm.push(local.data)
			return
		}
	}
}

func (fvm *ForthVM) lclr() {
	fvm.ln -= 1
}

func (fvm *ForthVM) lv() {
	fvm.push(fvm.mem[fvm.pop()])
}

func (fvm *ForthVM) lsi() {
	var v int64 = 0

	a := fvm.pop()
	b := fvm.pop()

	if a > b {
		v = 1
	}

	fvm.push(v)
}

func (fvm *ForthVM) gri() {
	var v int64 = 0

	a := fvm.pop()
	b := fvm.pop()

	if a < b {
		v = 1
	}

	fvm.push(v)
}

func (fvm *ForthVM) jin() bool {
	return fvm.pop() == 0
}

func (fvm *ForthVM) adi() {
	fvm.push(fvm.pop() + fvm.pop())
}

func (fvm *ForthVM) sbi() {
	a := fvm.pop()
	b := fvm.pop()
	fvm.push(b - a)
}

func (fvm *ForthVM) dvi() {
	a := fvm.pop()
	b := fvm.pop()
	fvm.push(b / a)
}

func (fvm *ForthVM) mli() {
	fvm.push(fvm.pop() * fvm.pop())
}

func (fvm *ForthVM) pri() {
	// fmt.Fprintf(w, "%d", pop())
	fmt.Printf("%d", fvm.pop())
}

func (fvm *ForthVM) pra() {
	// fmt.Fprintf(w, "%c", pop())
	fmt.Printf("%c", fvm.pop())
}

func (fvm *ForthVM) rdi() {
	var i int64
	fmt.Scanf("%d", &i)
	fvm.push(i)
}

func (fvm *ForthVM) eqi() {
	var v int64 = 0

	a := fvm.pop()
	b := fvm.pop()

	if a == b {
		v = 1
	}

	fvm.push(v)
}

func (fvm *ForthVM) and() {
	var v int64 = 0

	a := fvm.pop()
	b := fvm.pop()

	if a != 0 && b != 0 {
		v = 1
	}

	fvm.push(v)
}

func (fvm *ForthVM) or() {
	var v int64 = 0

	a := fvm.pop()
	b := fvm.pop()

	if a != 0 || b != 0 {
		v = 1
	}

	fvm.push(v)
}

func (fvm *ForthVM) not() {
	var v int64 = 0

	a := fvm.pop()

	if a == 0 {
		v = 1
	}

	fvm.push(v)
}

func (fvm *ForthVM) str() {
	a := fvm.pop()
	b := fvm.pop()

	fvm.mem[a] = b
}

// Forth functions

func (fvm *ForthVM) dup() {
	fvm.push(fvm.stack[fvm.n])
}

func (fvm *ForthVM) ovr() {
	fvm.push(fvm.stack[fvm.n-1])
}

func (fvm *ForthVM) tvr() {
	c := fvm.stack[fvm.n-2]
	d := fvm.stack[fvm.n-3]
	fvm.push(d)
	fvm.push(c)
}

func (fvm *ForthVM) twp() {
	a := fvm.pop()
	b := fvm.pop()
	c := fvm.pop()
	d := fvm.pop()
	fvm.push(b)
	fvm.push(a)
	fvm.push(d)
	fvm.push(c)
}

func (fvm *ForthVM) qdp() {
	a := fvm.stack[fvm.n]

	if a != 0 {
		fvm.push(a)
	}
}

func (fvm *ForthVM) rot() {
	a := fvm.pop()
	b := fvm.pop()
	c := fvm.pop()
	fvm.push(b)
	fvm.push(a)
	fvm.push(c)
}

func (fvm *ForthVM) tdp() {
	a := fvm.stack[fvm.n]
	b := fvm.stack[fvm.n-1]
	fvm.push(b)
	fvm.push(a)
}

func (fvm *ForthVM) drp() {
	fvm.pop()
}

func (fvm *ForthVM) swp() {
	a := fvm.stack[fvm.n]
	fvm.stack[fvm.n] = fvm.stack[fvm.n-1]
	fvm.stack[fvm.n-1] = a
}

func (fvm *ForthVM) sys() {
	value := fvm.pop()

	switch value {
	case 1:
		mod := fvm.pop()
		n := fvm.pop()
		fvm.push(n % mod)
	case 2:
		// dest-addr readfile
		content, err := os.ReadFile("file.txt")

		if err != nil {
			log.Fatal(err)
		}

		dest := fvm.pop()

		content_len := int64(len(content))
		fvm.mem[dest] = content_len

		for i := int64(0); i < content_len; i++ {
			fvm.mem[dest+1+i] = int64(content[i])
		}
	case 3:
		// read memory from image
		// name-addr readimage
		content, err := os.ReadFile("file.image")

		if err != nil {
			log.Fatal(err)
		}

		buf := bytes.NewReader(content)
		err = binary.Read(buf, binary.LittleEndian, &fvm.mem)

		if err != nil {
			log.Fatal(err)
		}
	case 4:
		// write memory to image
		// name-addr writeimage
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, fvm.mem)

		if err != nil {
			log.Fatal(err)
		}

		err = os.WriteFile("file.image", buf.Bytes(), 0666)

		if err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Println("ERROR")
	}
}

const (
	RDI = iota
	PRI
	PRA
	DUP
	OVR
	TVR
	TWP
	QDP
	ROT
	TDP
	DRP
	SWP
	ADI
	NOP
	JMP
	JIN
	SBI
	DVI
	LSI
	GRI
	MLI
	OR
	AND
	NOT
	EQI
	LV
	L
	STR
	SYS
	STP
	SUB
	END
	MAIN
	LCTX
	LDEF
	LSET
	LCL
	LCLR
	CALL
)

type Cell struct {
	cmd    int
	arg    int64
	argStr string
}

func (c Cell) String() string {
	return fmt.Sprintf("{cmd:%d, arg:%d, argStr:%s}", c.cmd, c.arg, c.argStr)
}

// (SUB xx ... END)* MAIN ... STP delimited by semicolon
func parseCode(code string) []Cell {
	cmds := strings.Split(code, ";")
	cells := make([]Cell, 0, len(cmds)+1)

	for _, cmd := range cmds {
		if cmd == "" {
			//fmt.Println("EMPTY")
			continue
		}
		//fmt.Println(cmd)
		scmd := strings.Split(cmd, " ")

		if len(scmd) == 2 && scmd[0][0] == '#' {
			// NOP
			cells = append(cells, Cell{cmd: NOP, argStr: scmd[0]})
		} else {
			switch scmd[0] {
			case "RDI":
				cells = append(cells, Cell{cmd: RDI})
			case "PRI":
				cells = append(cells, Cell{cmd: PRI})
			case "PRA":
				cells = append(cells, Cell{cmd: PRA})
			case "DUP":
				cells = append(cells, Cell{cmd: DUP})
			case "OVR":
				cells = append(cells, Cell{cmd: OVR})
			case "TVR":
				cells = append(cells, Cell{cmd: TVR})
			case "TWP":
				cells = append(cells, Cell{cmd: TWP})
			case "QDP":
				cells = append(cells, Cell{cmd: QDP})
			case "ROT":
				cells = append(cells, Cell{cmd: ROT})
			case "TDP":
				cells = append(cells, Cell{cmd: TDP})
			case "DRP":
				cells = append(cells, Cell{cmd: DRP})
			case "SWP":
				cells = append(cells, Cell{cmd: SWP})
			case "ADI":
				cells = append(cells, Cell{cmd: ADI})
			case "JMP":
				cells = append(cells, Cell{cmd: JMP, argStr: scmd[1]})
			case "JIN":
				cells = append(cells, Cell{cmd: JIN, argStr: scmd[1]})
			case "SBI":
				cells = append(cells, Cell{cmd: SBI})
			case "DVI":
				cells = append(cells, Cell{cmd: DVI})
			case "LSI":
				cells = append(cells, Cell{cmd: LSI})
			case "GRI":
				cells = append(cells, Cell{cmd: GRI})
			case "MLI":
				cells = append(cells, Cell{cmd: MLI})
			case "OR":
				cells = append(cells, Cell{cmd: OR})
			case "AND":
				cells = append(cells, Cell{cmd: AND})
			case "NOT":
				cells = append(cells, Cell{cmd: NOT})
			case "EQI":
				cells = append(cells, Cell{cmd: EQI})
			case "LV":
				cells = append(cells, Cell{cmd: LV})
			case "L":
				value, err := strconv.ParseInt(scmd[1], 10, 64)
				if err != nil {
					fmt.Println(err)
					os.Exit(2)
				}
				cells = append(cells, Cell{cmd: L, arg: value})
			case "STR":
				cells = append(cells, Cell{cmd: STR})
			case "SYS":
				cells = append(cells, Cell{cmd: SYS})
			case "STP":
				cells = append(cells, Cell{cmd: STP})
			case "SUB":
				cells = append(cells, Cell{cmd: SUB, argStr: scmd[1]})
			case "END":
				cells = append(cells, Cell{cmd: END})
			case "MAIN":
				cells = append(cells, Cell{cmd: MAIN})
			case "LCTX":
				cells = append(cells, Cell{cmd: LCTX})
			case "LSET":
				cells = append(cells, Cell{cmd: LSET, argStr: scmd[1]})
			case "LDEF":
				cells = append(cells, Cell{cmd: LDEF, argStr: scmd[1]})
			case "LCL":
				cells = append(cells, Cell{cmd: LCL, argStr: scmd[1]})
			case "LCLR":
				cells = append(cells, Cell{cmd: LCLR})
			case "CALL":
				cells = append(cells, Cell{cmd: CALL, argStr: scmd[1]})
			default:
				fmt.Println("ERROR: Unknown command [" + cmd + "]")
				os.Exit(2)
			}
		}
	}

	return cells
}

// (SUB xx ... END)* MAIN ... STP delimited by semicolon
func (fvm *ForthVM) Run(codeStr string) {
	labels := make(map[string]int)
	code := parseCode(codeStr)
	progPtr := 0
	locals := new(Stack)

	for pos, cell := range code {
		// fmt.Printf("%v\n", cell)

		switch cell.cmd {
		case NOP:
			labels[cell.argStr] = pos
		case MAIN:
			progPtr = pos + 1
		case SUB:
			labels[cell.argStr] = pos
		case LDEF:
			if !locals.Contains(cell.argStr) {
				locals.Push(cell.argStr)
			}
		}
	}

	// init locals

	// TODO: lege in jeden LDEF als cell.arg den Index aus locals hinein, so spart man sich GetIndex
	// TODO: Shebang fuer Skripte
	// TODO: verbose Flag um ByteCode auszublenden
	// TODO: Stack und Memory anzeigen

	fvm.ln = -1
	fvm.l_len = locals.Len()
	fvm.lstack = make([]Local, fvm.l_len*100)

	// end init locals

	done := false
	returnStack := make([]int, 0, 100)

	// fmt.Printf("progPtr: %d\n", progPtr)
	numCmds := 0
	start := time.Now()

	for ; !done; progPtr++ {
		numCmds++
		command := code[progPtr]

		switch command.cmd {
		case RDI:
			fvm.rdi()
		case PRI:
			fvm.pri()
		case PRA:
			fvm.pra()
		case DUP:
			fvm.dup()
		case OVR:
			fvm.ovr()
		case TVR:
			fvm.tvr()
		case TWP:
			fvm.twp()
		case QDP:
			fvm.qdp()
		case ROT:
			fvm.rot()
		case TDP:
			fvm.tdp()
		case DRP:
			fvm.drp()
		case SWP:
			fvm.swp()
		case ADI:
			fvm.adi()
		case NOP:
			// pass
		case JMP:
			progPtr = labels[command.argStr] - 1
		case JIN:
			if fvm.jin() {
				progPtr = labels[command.argStr] - 1
			}
		case SBI:
			fvm.sbi()
		case DVI:
			fvm.dvi()
		case LSI:
			fvm.lsi()
		case GRI:
			fvm.gri()
		case MLI:
			fvm.mli()
		case OR:
			fvm.or()
		case AND:
			fvm.and()
		case NOT:
			fvm.not()
		case EQI:
			fvm.eqi()
		case LV:
			fvm.lv()
		case L:
			fvm.push(command.arg)
		case STR:
			fvm.str()
		case SYS:
			fvm.sys()
		case STP:
			done = true
		case SUB:
			// pass
		case END:
			// pop callstack
			index := len(returnStack) - 1
			progPtr = returnStack[index]
			returnStack = returnStack[:index]
		case MAIN:
			// pass
		case LCTX:
			fvm.lctx()
		case LSET:
			fvm.lset(locals.GetIndex(command.argStr))
		case LDEF:
			fvm.ldef(locals.GetIndex(command.argStr))
		case LCL:
			fvm.lcl(locals.GetIndex(command.argStr))
		case LCLR:
			fvm.lclr()
		case CALL:
			// push callstack
			returnStack = append(returnStack, progPtr)
			progPtr = labels[command.argStr]
		default:
			fmt.Printf("ERROR: Unknown command %v", command)
			os.Exit(2)
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("\n\nexecution time: %s\nNumber of Cmds: %d\nSpeed: %f cmd/ns", elapsed, numCmds, float32(numCmds)/float32(elapsed.Nanoseconds()))
}
