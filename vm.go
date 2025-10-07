package goforth

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type Local struct {
	active bool
	data   int64
}

type ForthVM struct {
	Vars       map[string]int64
	Mem        []int64
	Stack      []int64
	Rstack     []int64
	lstack     []Local
	ln         int
	l_len      int
	Sysfunc    func(*ForthVM, int64)
	Out        io.Writer
	CodeData   *Code
	ExitStatus int
}

func NewForthVM() *ForthVM {
	fvm := new(ForthVM)

	fvm.Vars = make(map[string]int64)
	fvm.Stack = make([]int64, 0, 100)
	fvm.Rstack = make([]int64, 0, 100)
	fvm.Out = os.Stdout

	return fvm
}

func (fvm *ForthVM) Push(i int64) {
	fvm.Stack = append(fvm.Stack, i)
}

func (fvm *ForthVM) Pop() int64 {
	n := len(fvm.Stack) - 1
	value := fvm.Stack[n]
	fvm.Stack = fvm.Stack[:n]
	return value
}

func (fvm *ForthVM) Rpush(i int64) {
	fvm.Rstack = append(fvm.Rstack, i)
}

func (fvm *ForthVM) Rpop() int64 {
	rn := len(fvm.Rstack) - 1
	value := fvm.Rstack[rn]
	fvm.Rstack = fvm.Rstack[:rn]
	return value
}

func (fvm *ForthVM) Fpop() float64 {
	value := fvm.Pop()
	return *(*float64)(unsafe.Pointer(&value))
}

func (fvm *ForthVM) Fpush(f float64) {
	val := *(*int64)(unsafe.Pointer(&f))
	fvm.Push(val)
}

func (fvm *ForthVM) Lctx() {
	fvm.ln += 1
	for i := 0; i < fvm.l_len; i++ {
		fvm.local_get(i, fvm.ln).active = false
	}
}

func (fvm *ForthVM) local_get(name, ctx int) *Local {
	return &fvm.lstack[fvm.l_len*ctx+name]
}

func (fvm *ForthVM) Ldef(name int) {
	local := fvm.local_get(name, fvm.ln)
	local.data = fvm.Pop()
	local.active = true
}

func (fvm *ForthVM) Lset(name int) {
	v := fvm.Pop()
	for i := fvm.ln; i > -1; i-- {
		if local := fvm.local_get(name, i); local.active {
			local.data = v
			return
		}
	}
}

func (fvm *ForthVM) Lcl(name int) {
	for i := fvm.ln; i > -1; i-- {
		if local := fvm.local_get(name, i); local.active {
			fvm.Push(local.data)
			return
		}
	}
}

func (fvm *ForthVM) Lclr() {
	fvm.ln -= 1
}

func (fvm *ForthVM) Lv() {
	fvm.Push(fvm.Mem[fvm.Pop()])
}

func (fvm *ForthVM) Lsi() {
	var v int64

	a := fvm.Pop()
	b := fvm.Pop()

	if a > b {
		v = 1
	}

	fvm.Push(v)
}

func (fvm *ForthVM) Gri() {
	var v int64

	a := fvm.Pop()
	b := fvm.Pop()

	if a < b {
		v = 1
	}

	fvm.Push(v)
}

func (fvm *ForthVM) Jin() bool {
	return fvm.Pop() == 0
}

func (fvm *ForthVM) Adi() {
	// fvm.Push(fvm.Pop() + fvm.Pop())

	a := fvm.Pop()
	n := len(fvm.Stack) - 1

	fvm.Stack[n] = a + fvm.Stack[n]
}

func (fvm *ForthVM) Sbi() {
	a := fvm.Pop()
	b := fvm.Pop()
	fvm.Push(b - a)
}

func (fvm *ForthVM) Dvi() {
	a := fvm.Pop()
	b := fvm.Pop()
	fvm.Push(b / a)
}

func (fvm *ForthVM) Mli() {
	fvm.Push(fvm.Pop() * fvm.Pop())
}

// f+
func (fvm *ForthVM) Adf() {
	// fvm.Fpush(fvm.Fpop() + fvm.Fpop())

	a := fvm.Fpop()
	n := len(fvm.Stack) - 1
	s := a + *(*float64)(unsafe.Pointer(&fvm.Stack[n]))

	fvm.Stack[n] = *(*int64)(unsafe.Pointer(&s))
}

// f-
func (fvm *ForthVM) Sbf() {
	// a := fvm.Fpop()
	// b := fvm.Fpop()
	// fvm.Fpush(b - a)

	a := fvm.Fpop()
	n := len(fvm.Stack) - 1
	s := *(*float64)(unsafe.Pointer(&fvm.Stack[n])) - a

	fvm.Stack[n] = *(*int64)(unsafe.Pointer(&s))
}

// f/
func (fvm *ForthVM) Dvf() {
	// a := fvm.Fpop()
	// b := fvm.Fpop()
	// fvm.Fpush(b / a)

	a := fvm.Fpop()
	n := len(fvm.Stack) - 1
	s := *(*float64)(unsafe.Pointer(&fvm.Stack[n])) / a

	fvm.Stack[n] = *(*int64)(unsafe.Pointer(&s))
}

// f*
func (fvm *ForthVM) Mlf() {
	// fvm.Fpush(fvm.Fpop() * fvm.Fpop())

	a := fvm.Fpop()
	n := len(fvm.Stack) - 1
	s := *(*float64)(unsafe.Pointer(&fvm.Stack[n])) * a

	fvm.Stack[n] = *(*int64)(unsafe.Pointer(&s))
}

func (fvm *ForthVM) Pri() {
	fmt.Fprintf(fvm.Out, "%d", fvm.Pop())
}

// f.
func (fvm *ForthVM) Prf() {
	fmt.Fprintf(fvm.Out, "%f", fvm.Fpop())
}

// f<
func (fvm *ForthVM) Lsf() {
	var v int64

	a := fvm.Fpop()
	b := fvm.Fpop()

	if a > b {
		v = 1
	}

	fvm.Push(v)
}

// f>
func (fvm *ForthVM) Grf() {
	var v int64

	a := fvm.Fpop()
	b := fvm.Fpop()

	if a < b {
		v = 1
	}

	fvm.Push(v)
}

func (fvm *ForthVM) Pra() {
	fmt.Fprintf(fvm.Out, "%c", fvm.Pop())
}

func (fvm *ForthVM) Rdi() {
	var i int64
	fmt.Scanf("%d", &i)
	fvm.Push(i)
}

func (fvm *ForthVM) Eqi() {
	var v int64

	a := fvm.Pop()
	b := fvm.Pop()

	if a == b {
		v = 1
	}

	fvm.Push(v)
}

func (fvm *ForthVM) Xor() {
	a := fvm.Pop()
	b := fvm.Pop()

	fvm.Push(a ^ b)
}

func (fvm *ForthVM) And() {
	var v int64

	a := fvm.Pop()
	b := fvm.Pop()

	if a != 0 && b != 0 {
		v = 1
	}

	fvm.Push(v)
}

func (fvm *ForthVM) Or() {
	var v int64

	a := fvm.Pop()
	b := fvm.Pop()

	if a != 0 || b != 0 {
		v = 1
	}

	fvm.Push(v)
}

func (fvm *ForthVM) Not() {
	var v int64

	a := fvm.Pop()

	if a == 0 {
		v = 1
	}

	fvm.Push(v)
}

func (fvm *ForthVM) Str() {
	a := fvm.Pop()
	b := fvm.Pop()

	fvm.Mem[a] = b
}

// Forth functions

func (fvm *ForthVM) Dup() {
	fvm.Push(fvm.Stack[len(fvm.Stack)-1])
}

func (fvm *ForthVM) Pick() {
	// value := int(fvm.Pop())
	// fvm.Push(fvm.Stack[(len(fvm.Stack)-1)-value])

	n := len(fvm.Stack) - 1
	value := int(fvm.Stack[n])
	fvm.Stack[n] = fvm.Stack[n-1-value]
}

func (fvm *ForthVM) Ovr() {
	fvm.Push(fvm.Stack[len(fvm.Stack)-2])
}

func (fvm *ForthVM) Tvr() {
	n := len(fvm.Stack) - 1
	c := fvm.Stack[n-2]
	d := fvm.Stack[n-3]
	fvm.Push(d)
	fvm.Push(c)
}

func (fvm *ForthVM) Twp() {
	a := fvm.Pop()
	b := fvm.Pop()
	c := fvm.Pop()
	d := fvm.Pop()
	fvm.Push(b)
	fvm.Push(a)
	fvm.Push(d)
	fvm.Push(c)
}

func (fvm *ForthVM) Qdp() {
	n := len(fvm.Stack) - 1
	a := fvm.Stack[n]

	if a != 0 {
		fvm.Push(a)
	}
}

func (fvm *ForthVM) Rot() {
	a := fvm.Pop()
	b := fvm.Pop()
	c := fvm.Pop()
	fvm.Push(b)
	fvm.Push(a)
	fvm.Push(c)
}

func (fvm *ForthVM) Nrot() {
	a := fvm.Pop()
	b := fvm.Pop()
	c := fvm.Pop()
	fvm.Push(a)
	fvm.Push(c)
	fvm.Push(b)
}

func (fvm *ForthVM) Tdp() {
	n := len(fvm.Stack) - 1
	a := fvm.Stack[n]
	b := fvm.Stack[n-1]
	fvm.Push(b)
	fvm.Push(a)
}

func (fvm *ForthVM) Drp() {
	fvm.Pop()
}

func (fvm *ForthVM) Swp() {
	n := len(fvm.Stack) - 1
	a := fvm.Stack[n]
	fvm.Stack[n] = fvm.Stack[n-1]
	fvm.Stack[n-1] = a
}

func (fvm *ForthVM) Tr() {
	fvm.Rpush(fvm.Pop())
}

func (fvm *ForthVM) Fr() {
	fvm.Push(fvm.Rpop())
}

func (fvm *ForthVM) Rf() {
	rn := len(fvm.Rstack) - 1
	fvm.Push(fvm.Rstack[rn])
}

func (fvm *ForthVM) Ttr() {
	a := fvm.Pop()
	b := fvm.Pop()

	fvm.Rpush(b)
	fvm.Rpush(a)
}

func (fvm *ForthVM) Tfr() {
	a := fvm.Rpop()
	b := fvm.Rpop()

	fvm.Push(b)
	fvm.Push(a)
}

func (fvm *ForthVM) Trf() {
	rn := len(fvm.Rstack) - 1
	fvm.Push(fvm.Rstack[rn-1])
	fvm.Push(fvm.Rstack[rn])
}

func (fvm *ForthVM) Inc() {
	n := len(fvm.Stack) - 1
	fvm.Stack[n]++
}

func (fvm *ForthVM) Dec() {
	n := len(fvm.Stack) - 1
	fvm.Stack[n]--
}

func (fvm *ForthVM) Gdef(name string) {
	fvm.Vars[name] = 0
}

func (fvm *ForthVM) Gbl(name string) {
	fvm.Push(fvm.Vars[name])
}

func (fvm *ForthVM) Gset(name string) {
	value := fvm.Pop()
	fvm.Vars[name] = value
}

func (fvm *ForthVM) GetString() string {
	value := fvm.Pop()
	length := fvm.Mem[value]
	data := fvm.Mem[value+1]
	var builder strings.Builder

	for i := int64(0); i < length; i++ {
		err := builder.WriteByte(byte(fvm.Mem[data+i]))

		if err != nil {
			log.Fatal(err)
		}
	}

	return builder.String()
}

// Push str to the fvm stack
func (fvm *ForthVM) StringToStack(str string) {
	fvm.Push(0)
	length := int64(len(str))

	for i := length - 1; i >= 0; i-- {
		fvm.Push(int64(str[i]))
	}

	fvm.Push(length)
}

func (fvm *ForthVM) Sys() {
	syscall := fvm.Pop()

	switch syscall {
	case 0:
		// depth
		fvm.Push(int64(len(fvm.Stack)))
	case 1:
		mod := fvm.Pop()
		n := fvm.Pop()
		fvm.Push(n % mod)
	case 2:
		// fsqrt
		value := fvm.Fpop()
		fvm.Fpush(math.Sqrt(value))
	case 3:
		// i->f
		value := fvm.Pop()
		fvm.Fpush(float64(value))
	case 4:
		// f->i
		value := fvm.Fpop()
		fvm.Push(int64(value))
	case 5:
		// name-addr readfile
		name := fvm.GetString()
		content, err := os.ReadFile(name)

		if err != nil {
			log.Fatal(err)
		}

		fvm.StringToStack(string(content))
	case 6:
		// read memory from image
		// name-addr readimage
		name := fvm.GetString()
		content, err := os.ReadFile(name)

		if err != nil {
			log.Fatal(err)
		}

		buf := bytes.NewReader(content)
		err = binary.Read(buf, binary.LittleEndian, &fvm.Mem)

		if err != nil {
			log.Fatal(err)
		}
	case 7:
		// write memory into image
		// name-addr writeimage
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, fvm.Mem)

		if err != nil {
			log.Fatal(err)
		}

		name := fvm.GetString()
		err = os.WriteFile(name, buf.Bytes(), 0666)

		if err != nil {
			log.Fatal(err)
		}
	case 8:
		// num-bytes read
		nbytes := fvm.Pop()
		buf := make([]byte, nbytes)
		if n, err := os.Stdin.Read(buf); err != nil {
			fvm.StringToStack("")
		} else {
			str := string(buf[:n])
			fvm.StringToStack(str)
		}
	case 9:
		ShowByteCode = fvm.Pop() != 0
		ShowExecutionTime = ShowByteCode
	case 10:
		n := fvm.Pop()
		mem := make([]int64, n)
		copy(mem, fvm.Mem)
		fvm.Mem = mem
	case 11:
		fvm.Push(int64(len(fvm.Mem)))
	case 12:
		// compare
		str1 := fvm.GetString()
		str2 := fvm.GetString()

		if str1 == str2 {
			fvm.Push(1)
		} else {
			fvm.Push(0)
		}
	case 13:
		// shell
		str := fvm.GetString()
		cmd := exec.Command("sh", "-c", str)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			PrintError(err)
		}
	case 14:
		// system
		str := fvm.GetString()
		cmd := exec.Command(str)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			PrintError(err)
		}
	case 15:
		// file
		str := fvm.GetString()
		info, err := os.Stat(str)

		if os.IsNotExist(err) {
			fvm.Push(0)
		} else if !info.IsDir() {
			fvm.Push(1)
		} else {
			fvm.Push(0)
		}
	case 16:
		// argc
		fvm.Push(int64(len(os.Args)))
	case 17:
		// i argv
		n := fvm.Pop()
		arg := os.Args[n]
		fvm.StringToStack(arg)
	default:
		if fvm.Sysfunc != nil {
			fvm.Sysfunc(fvm, syscall)
		} else {
			log.Fatalf("ERROR: sys() - Unknown call \"%d\"\n", syscall)
		}
	}
}

type Opcode int

const (
	RDI Opcode = iota
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
	NOP
	JMP
	JIN
	ADI
	SBI
	DVI
	LSI
	GRI
	MLI
	ADF
	SBF
	MLF
	DVF
	PRF
	LSF
	GRF
	OR
	AND
	NOT
	EQI
	XOR
	LV
	L
	LF
	STR
	SYS
	STP
	SUB
	END
	MAIN
	GDEF
	GSET
	GBL
	LCTX
	LDEF
	LSET
	LCL
	LCLR
	CALL
	REF
	EXC
	PCK
	NRT
	TR  // to r
	FR  // from r
	RF  // r fetch
	TTR // 2 to r
	TFR // 2 from r
	TRF // 2 r fetch
	INC
	DEC
)

var CellName = map[Opcode]string{
	RDI:  "RDI",
	PRI:  "PRI",
	PRA:  "PRA",
	DUP:  "DUP",
	OVR:  "OVR",
	TVR:  "TVR",
	TWP:  "TWP",
	QDP:  "QDP",
	ROT:  "ROT",
	TDP:  "TDP",
	DRP:  "DRP",
	SWP:  "SWP",
	NOP:  "NOP",
	JMP:  "JMP",
	JIN:  "JIN",
	ADI:  "ADI",
	SBI:  "SBI",
	DVI:  "DVI",
	LSI:  "LSI",
	GRI:  "GRI",
	MLI:  "MLI",
	ADF:  "ADF",
	SBF:  "SBF",
	MLF:  "MLF",
	DVF:  "DVF",
	PRF:  "PRF",
	LSF:  "LSF",
	GRF:  "GRF",
	OR:   "OR",
	AND:  "AND",
	NOT:  "NOT",
	EQI:  "EQI",
	XOR:  "XOR",
	LV:   "LV",
	L:    "L",
	LF:   "LF",
	STR:  "STR",
	SYS:  "SYS",
	STP:  "STP",
	SUB:  "SUB",
	END:  "END",
	MAIN: "MAIN",
	GDEF: "GDEF",
	GSET: "GSET",
	GBL:  "GBL",
	LCTX: "LCTX",
	LDEF: "LDEF",
	LSET: "LSET",
	LCL:  "LCL",
	LCLR: "LCLR",
	CALL: "CALL",
	REF:  "REF",
	EXC:  "EXC",
	PCK:  "PCK",
	NRT:  "NRT",
	TR:   "TR",
	FR:   "FR",
	RF:   "RF",
	TTR:  "TTR",
	TFR:  "TFR",
	TRF:  "TRF",
	INC:  "INC",
	DEC:  "DEC",
}

type Cell struct {
	cmd        Opcode
	arg        int64
	argf       float64
	argStr     string
	localIndex int
}

func (c Cell) String() string {
	switch c.cmd {
	case L:
		return fmt.Sprintf("%s %d", CellName[c.cmd], c.arg)
	case LDEF, LSET, CALL, JIN, JMP, NOP, REF, LCL, GDEF, GSET, GBL:
		return fmt.Sprintf("%s %s", CellName[c.cmd], c.argStr)
	case LF:
		return fmt.Sprintf("%s %f", CellName[c.cmd], c.argf)
	default:
		return CellName[c.cmd]
	}
}

type Code struct {
	cells     []Cell         // actual code
	labels    map[string]int // labels indices of NOP and SUB
	numLocals int            // number of locals
	PosMain   int            // position of MAIN
	ProgPtr   int            // program pointer, used in RunStep
	Command   *Cell          // current command to execute, used in RunStep
}

// (SUB xx ... END)* MAIN ... STP delimited by semicolon
func parseCode(codeStr string) *Code {
	code := new(Code)
	code.labels = make(map[string]int)

	cmds := strings.Split(codeStr, ";")
	cells := make([]Cell, 0, len(cmds)+1)
	locals := NewStack[string]()

	for pos, cmd := range cmds {
		if cmd == "" {
			//fmt.Println("EMPTY")
			continue
		}
		//fmt.Println(cmd)
		scmd := strings.Split(cmd, " ")

		switch scmd[0] {
		case "NOP":
			code.labels[scmd[1]] = pos
			cells = append(cells, Cell{cmd: NOP, argStr: scmd[1]})
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
		case "ADF":
			cells = append(cells, Cell{cmd: ADF})
		case "SBF":
			cells = append(cells, Cell{cmd: SBF})
		case "MLF":
			cells = append(cells, Cell{cmd: MLF})
		case "DVF":
			cells = append(cells, Cell{cmd: DVF})
		case "PRF":
			cells = append(cells, Cell{cmd: PRF})
		case "LSF":
			cells = append(cells, Cell{cmd: LSF})
		case "GRF":
			cells = append(cells, Cell{cmd: GRF})
		case "OR":
			cells = append(cells, Cell{cmd: OR})
		case "AND":
			cells = append(cells, Cell{cmd: AND})
		case "NOT":
			cells = append(cells, Cell{cmd: NOT})
		case "EQI":
			cells = append(cells, Cell{cmd: EQI})
		case "XOR":
			cells = append(cells, Cell{cmd: XOR})
		case "LV":
			cells = append(cells, Cell{cmd: LV})
		case "L":
			value, err := strconv.ParseInt(scmd[1], 10, 64)
			if err != nil {
				log.Fatal(err)
			}
			cells = append(cells, Cell{cmd: L, arg: value})
		case "LF":
			value, err := strconv.ParseFloat(scmd[1], 64)
			if err != nil {
				log.Fatal(err)
			}
			cells = append(cells, Cell{cmd: LF, argf: value})
		case "STR":
			cells = append(cells, Cell{cmd: STR})
		case "SYS":
			cells = append(cells, Cell{cmd: SYS})
		case "STP":
			cells = append(cells, Cell{cmd: STP})
		case "SUB":
			code.labels[scmd[1]] = pos
			cells = append(cells, Cell{cmd: SUB, argStr: scmd[1]})
		case "END":
			cells = append(cells, Cell{cmd: END})
		case "MAIN":
			code.PosMain = pos
			cells = append(cells, Cell{cmd: MAIN})
		case "GDEF":
			cells = append(cells, Cell{cmd: GDEF, argStr: scmd[1]})
		case "GSET":
			cells = append(cells, Cell{cmd: GSET, argStr: scmd[1]})
		case "GBL":
			cells = append(cells, Cell{cmd: GBL, argStr: scmd[1]})
		case "LCTX":
			cells = append(cells, Cell{cmd: LCTX})
		case "LSET":
			if !locals.Contains(scmd[1]) {
				locals.Push(scmd[1])
			}
			cells = append(cells, Cell{cmd: LSET, argStr: scmd[1], localIndex: locals.Index(scmd[1])})
		case "LDEF":
			if !locals.Contains(scmd[1]) {
				locals.Push(scmd[1])
			}
			cells = append(cells, Cell{cmd: LDEF, argStr: scmd[1], localIndex: locals.Index(scmd[1])})
		case "LCL":
			if !locals.Contains(scmd[1]) {
				locals.Push(scmd[1])
			}
			cells = append(cells, Cell{cmd: LCL, argStr: scmd[1], localIndex: locals.Index(scmd[1])})
		case "LCLR":
			cells = append(cells, Cell{cmd: LCLR})
		case "CALL":
			cells = append(cells, Cell{cmd: CALL, argStr: scmd[1]})
		case "REF":
			cells = append(cells, Cell{cmd: REF, argStr: scmd[1]})
		case "EXC":
			cells = append(cells, Cell{cmd: EXC})
		case "PCK":
			cells = append(cells, Cell{cmd: PCK})
		case "NRT":
			cells = append(cells, Cell{cmd: NRT})
		case "TR":
			cells = append(cells, Cell{cmd: TR})
		case "FR":
			cells = append(cells, Cell{cmd: FR})
		case "RF":
			cells = append(cells, Cell{cmd: RF})
		case "TTR":
			cells = append(cells, Cell{cmd: TTR})
		case "TFR":
			cells = append(cells, Cell{cmd: TFR})
		case "TRF":
			cells = append(cells, Cell{cmd: TRF})
		case "INC":
			cells = append(cells, Cell{cmd: INC})
		case "DEC":
			cells = append(cells, Cell{cmd: DEC})
		default:
			log.Fatalf("ERROR: Unknown command \"%s\"\n", cmd)
		}
	}

	code.cells = cells
	code.numLocals = locals.Len()

	return code
}

// Initializes the virtual machine.
// You should call the method before RunStep.
func (fvm *ForthVM) PrepareRun(codeStr string) {
	fvm.CodeData = parseCode(codeStr)
	fvm.CodeData.ProgPtr = fvm.CodeData.PosMain
	fvm.CodeData.Command = &fvm.CodeData.cells[fvm.CodeData.ProgPtr]

	fvm.ln = -1
	fvm.l_len = fvm.CodeData.numLocals
	fvm.lstack = make([]Local, fvm.l_len*100)

	fvm.Rstack = fvm.Rstack[:0]
}

// Executes the ByteCode passed as a parameter.
func (fvm *ForthVM) Run(codeStr string) {
	fvm.PrepareRun(codeStr)

	done := false
	numCmds := int64(0)
	start := time.Now()

	for progPtr := fvm.CodeData.PosMain + 1; !done; progPtr++ {
		numCmds++
		command := &fvm.CodeData.cells[progPtr]

		switch command.cmd {
		case RDI:
			fvm.Rdi()
		case PRI:
			fvm.Pri()
		case PRA:
			fvm.Pra()
		case DUP:
			fvm.Dup()
		case OVR:
			fvm.Ovr()
		case TVR:
			fvm.Tvr()
		case TWP:
			fvm.Twp()
		case QDP:
			fvm.Qdp()
		case ROT:
			fvm.Rot()
		case TDP:
			fvm.Tdp()
		case DRP:
			fvm.Drp()
		case SWP:
			fvm.Swp()
		case ADI:
			fvm.Adi()
		case NOP:
			// pass
		case JMP:
			progPtr = fvm.CodeData.labels[command.argStr] - 1
		case JIN:
			if fvm.Jin() {
				progPtr = fvm.CodeData.labels[command.argStr] - 1
			}
		case SBI:
			fvm.Sbi()
		case DVI:
			fvm.Dvi()
		case LSI:
			fvm.Lsi()
		case GRI:
			fvm.Gri()
		case MLI:
			fvm.Mli()
		case ADF:
			fvm.Adf()
		case SBF:
			fvm.Sbf()
		case MLF:
			fvm.Mlf()
		case DVF:
			fvm.Dvf()
		case PRF:
			fvm.Prf()
		case LSF:
			fvm.Lsf()
		case GRF:
			fvm.Grf()
		case OR:
			fvm.Or()
		case AND:
			fvm.And()
		case NOT:
			fvm.Not()
		case EQI:
			fvm.Eqi()
		case XOR:
			fvm.Xor()
		case LV:
			fvm.Lv()
		case L:
			fvm.Push(command.arg)
		case LF:
			fvm.Fpush(command.argf)
		case STR:
			fvm.Str()
		case SYS:
			fvm.Sys()
		case STP:
			fvm.ExitStatus = int(fvm.Pop())
			done = true
		case SUB:
			// pass
		case END:
			progPtr = int(fvm.Rpop())
		case MAIN:
			// pass
		case GDEF:
			fvm.Gdef(command.argStr)
		case GSET:
			fvm.Gset(command.argStr)
		case GBL:
			fvm.Gbl(command.argStr)
		case LCTX:
			fvm.Lctx()
		case LSET:
			fvm.Lset(command.localIndex)
		case LDEF:
			fvm.Ldef(command.localIndex)
		case LCL:
			fvm.Lcl(command.localIndex)
		case LCLR:
			fvm.Lclr()
		case CALL:
			fvm.Rpush(int64(progPtr))
			progPtr = fvm.CodeData.labels[command.argStr]
		case REF:
			fvm.Push(int64(fvm.CodeData.labels[command.argStr]))
		case EXC:
			fvm.Rpush(int64(progPtr))
			progPtr = int(fvm.Pop())
		case PCK:
			fvm.Pick()
		case NRT:
			fvm.Nrot()
		case TR:
			fvm.Tr()
		case FR:
			fvm.Fr()
		case RF:
			fvm.Rf()
		case TTR:
			fvm.Ttr()
		case TFR:
			fvm.Tfr()
		case TRF:
			fvm.Trf()
		case INC:
			fvm.Inc()
		case DEC:
			fvm.Dec()
		default:
			log.Fatalf("ERROR: Unknown command %v\n", command)
		}
	}

	if ShowExecutionTime {
		elapsed := time.Since(start)
		fmt.Printf("\n\nexecution time: %s\nNumber of Cmds: %d\nSpeed: %f cmd/ns", elapsed, numCmds, float64(numCmds)/float64(elapsed.Nanoseconds()))
	}
}

// Runs a single step of the virtual machine.
// Note: You must call PrepareRun once before calling RunStep for the first time
func (fvm *ForthVM) RunStep() (bool, error) {
	fvm.CodeData.Command = &fvm.CodeData.cells[fvm.CodeData.ProgPtr]

	switch fvm.CodeData.Command.cmd {
	case RDI:
		fvm.Rdi()
	case PRI:
		fvm.Pri()
	case PRA:
		fvm.Pra()
	case DUP:
		fvm.Dup()
	case OVR:
		fvm.Ovr()
	case TVR:
		fvm.Tvr()
	case TWP:
		fvm.Twp()
	case QDP:
		fvm.Qdp()
	case ROT:
		fvm.Rot()
	case TDP:
		fvm.Tdp()
	case DRP:
		fvm.Drp()
	case SWP:
		fvm.Swp()
	case ADI:
		fvm.Adi()
	case NOP:
		// pass
	case JMP:
		fvm.CodeData.ProgPtr = fvm.CodeData.labels[fvm.CodeData.Command.argStr] - 1
	case JIN:
		if fvm.Jin() {
			fvm.CodeData.ProgPtr = fvm.CodeData.labels[fvm.CodeData.Command.argStr] - 1
		}
	case SBI:
		fvm.Sbi()
	case DVI:
		fvm.Dvi()
	case LSI:
		fvm.Lsi()
	case GRI:
		fvm.Gri()
	case MLI:
		fvm.Mli()
	case ADF:
		fvm.Adf()
	case SBF:
		fvm.Sbf()
	case MLF:
		fvm.Mlf()
	case DVF:
		fvm.Dvf()
	case PRF:
		fvm.Prf()
	case LSF:
		fvm.Lsf()
	case GRF:
		fvm.Grf()
	case OR:
		fvm.Or()
	case AND:
		fvm.And()
	case NOT:
		fvm.Not()
	case EQI:
		fvm.Eqi()
	case XOR:
		fvm.Xor()
	case LV:
		fvm.Lv()
	case L:
		fvm.Push(fvm.CodeData.Command.arg)
	case LF:
		fvm.Fpush(fvm.CodeData.Command.argf)
	case STR:
		fvm.Str()
	case SYS:
		fvm.Sys()
	case STP:
		fvm.ExitStatus = int(fvm.Pop())
		return true, nil
	case SUB:
		// pass
	case END:
		fvm.CodeData.ProgPtr = int(fvm.Rpop())
	case MAIN:
		// pass
	case GDEF:
		fvm.Gdef(fvm.CodeData.Command.argStr)
	case GSET:
		fvm.Gset(fvm.CodeData.Command.argStr)
	case GBL:
		fvm.Gbl(fvm.CodeData.Command.argStr)
	case LCTX:
		fvm.Lctx()
	case LSET:
		fvm.Lset(fvm.CodeData.Command.localIndex)
	case LDEF:
		fvm.Ldef(fvm.CodeData.Command.localIndex)
	case LCL:
		fvm.Lcl(fvm.CodeData.Command.localIndex)
	case LCLR:
		fvm.Lclr()
	case CALL:
		fvm.Rpush(int64(fvm.CodeData.ProgPtr))
		fvm.CodeData.ProgPtr = fvm.CodeData.labels[fvm.CodeData.Command.argStr]
	case REF:
		fvm.Push(int64(fvm.CodeData.labels[fvm.CodeData.Command.argStr]))
	case EXC:
		fvm.Rpush(int64(fvm.CodeData.ProgPtr))
		fvm.CodeData.ProgPtr = int(fvm.Pop())
	case PCK:
		fvm.Pick()
	case NRT:
		fvm.Nrot()
	case TR:
		fvm.Tr()
	case FR:
		fvm.Fr()
	case RF:
		fvm.Rf()
	case TTR:
		fvm.Ttr()
	case TFR:
		fvm.Tfr()
	case TRF:
		fvm.Trf()
	case INC:
		fvm.Inc()
	case DEC:
		fvm.Dec()
	default:
		return true, fmt.Errorf("ERROR: Unknown command %v", fvm.CodeData.Command)
	}

	fvm.CodeData.ProgPtr++
	return false, nil
}
