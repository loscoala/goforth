package goforth

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type Local struct {
	active bool
	data   int64
}

type VM interface {
	Push(int64)
	Pop() int64
	Rpush(int64)
	Rpop() int64
	Fpush(float64)
	Fpop() float64
	GetString() string
	SetString(string)
	GetMemory() *[1000]int64
}

type ForthVM struct {
	mem     [1000]int64
	stack   [100]int64
	n       int
	rstack  [100]int64
	rn      int
	lstack  []Local
	ln      int
	l_len   int
	Sysfunc func(VM, int64)
	Out     io.Writer
}

func NewForthVM() *ForthVM {
	fvm := new(ForthVM)

	fvm.n = -1
	fvm.rn = -1
	fvm.Out = os.Stdout

	return fvm
}

func (fvm *ForthVM) Push(i int64) {
	fvm.n += 1
	fvm.stack[fvm.n] = i
}

func (fvm *ForthVM) Pop() int64 {
	v := fvm.stack[fvm.n]
	fvm.n -= 1
	return v
}

func (fvm *ForthVM) Rpush(i int64) {
	fvm.rn += 1
	fvm.rstack[fvm.rn] = i
}

func (fvm *ForthVM) Rpop() int64 {
	v := fvm.rstack[fvm.rn]
	fvm.rn -= 1
	return v
}

func (fvm *ForthVM) Fpop() float64 {
	value := fvm.Pop()
	return *(*float64)(unsafe.Pointer(&value))
}

func (fvm *ForthVM) Fpush(f float64) {
	val := *(*int64)(unsafe.Pointer(&f))
	fvm.Push(val)
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
	v := fvm.Pop()
	local := fvm.local_get(name, fvm.ln)
	local.data = v
	local.active = true
}

func (fvm *ForthVM) lset(name int) {
	v := fvm.Pop()
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
			fvm.Push(local.data)
			return
		}
	}
}

func (fvm *ForthVM) lclr() {
	fvm.ln -= 1
}

func (fvm *ForthVM) lv() {
	fvm.Push(fvm.mem[fvm.Pop()])
}

func (fvm *ForthVM) lsi() {
	var v int64 = 0

	a := fvm.Pop()
	b := fvm.Pop()

	if a > b {
		v = 1
	}

	fvm.Push(v)
}

func (fvm *ForthVM) gri() {
	var v int64 = 0

	a := fvm.Pop()
	b := fvm.Pop()

	if a < b {
		v = 1
	}

	fvm.Push(v)
}

func (fvm *ForthVM) jin() bool {
	return fvm.Pop() == 0
}

func (fvm *ForthVM) adi() {
	fvm.Push(fvm.Pop() + fvm.Pop())
}

func (fvm *ForthVM) sbi() {
	a := fvm.Pop()
	b := fvm.Pop()
	fvm.Push(b - a)
}

func (fvm *ForthVM) dvi() {
	a := fvm.Pop()
	b := fvm.Pop()
	fvm.Push(b / a)
}

func (fvm *ForthVM) mli() {
	fvm.Push(fvm.Pop() * fvm.Pop())
}

// f+
func (fvm *ForthVM) adf() {
	fvm.Fpush(fvm.Fpop() + fvm.Fpop())
}

// f-
func (fvm *ForthVM) sbf() {
	a := fvm.Fpop()
	b := fvm.Fpop()
	fvm.Fpush(b - a)
}

// f/
func (fvm *ForthVM) dvf() {
	a := fvm.Fpop()
	b := fvm.Fpop()
	fvm.Fpush(b / a)
}

// f*
func (fvm *ForthVM) mlf() {
	fvm.Fpush(fvm.Fpop() * fvm.Fpop())
}

func (fvm *ForthVM) pri() {
	fmt.Fprintf(fvm.Out, "%d", fvm.Pop())
}

// f.
func (fvm *ForthVM) prf() {
	fmt.Fprintf(fvm.Out, "%f", fvm.Fpop())
}

// f<
func (fvm *ForthVM) lsf() {
	var v int64 = 0

	a := fvm.Fpop()
	b := fvm.Fpop()

	if a > b {
		v = 1
	}

	fvm.Push(v)
}

// f>
func (fvm *ForthVM) grf() {
	var v int64 = 0

	a := fvm.Fpop()
	b := fvm.Fpop()

	if a < b {
		v = 1
	}

	fvm.Push(v)
}

func (fvm *ForthVM) pra() {
	fmt.Fprintf(fvm.Out, "%c", fvm.Pop())
}

func (fvm *ForthVM) rdi() {
	var i int64
	fmt.Scanf("%d", &i)
	fvm.Push(i)
}

func (fvm *ForthVM) eqi() {
	var v int64 = 0

	a := fvm.Pop()
	b := fvm.Pop()

	if a == b {
		v = 1
	}

	fvm.Push(v)
}

func (fvm *ForthVM) and() {
	var v int64 = 0

	a := fvm.Pop()
	b := fvm.Pop()

	if a != 0 && b != 0 {
		v = 1
	}

	fvm.Push(v)
}

func (fvm *ForthVM) or() {
	var v int64 = 0

	a := fvm.Pop()
	b := fvm.Pop()

	if a != 0 || b != 0 {
		v = 1
	}

	fvm.Push(v)
}

func (fvm *ForthVM) not() {
	var v int64 = 0

	a := fvm.Pop()

	if a == 0 {
		v = 1
	}

	fvm.Push(v)
}

func (fvm *ForthVM) str() {
	a := fvm.Pop()
	b := fvm.Pop()

	fvm.mem[a] = b
}

// Forth functions

func (fvm *ForthVM) dup() {
	fvm.Push(fvm.stack[fvm.n])
}

func (fvm *ForthVM) pick() {
	value := int(fvm.Pop())
	fvm.Push(fvm.stack[fvm.n-value])
}

func (fvm *ForthVM) ovr() {
	fvm.Push(fvm.stack[fvm.n-1])
}

func (fvm *ForthVM) tvr() {
	c := fvm.stack[fvm.n-2]
	d := fvm.stack[fvm.n-3]
	fvm.Push(d)
	fvm.Push(c)
}

func (fvm *ForthVM) twp() {
	a := fvm.Pop()
	b := fvm.Pop()
	c := fvm.Pop()
	d := fvm.Pop()
	fvm.Push(b)
	fvm.Push(a)
	fvm.Push(d)
	fvm.Push(c)
}

func (fvm *ForthVM) qdp() {
	a := fvm.stack[fvm.n]

	if a != 0 {
		fvm.Push(a)
	}
}

func (fvm *ForthVM) rot() {
	a := fvm.Pop()
	b := fvm.Pop()
	c := fvm.Pop()
	fvm.Push(b)
	fvm.Push(a)
	fvm.Push(c)
}

func (fvm *ForthVM) nrot() {
	a := fvm.Pop()
	b := fvm.Pop()
	c := fvm.Pop()
	fvm.Push(a)
	fvm.Push(c)
	fvm.Push(b)
}

func (fvm *ForthVM) tdp() {
	a := fvm.stack[fvm.n]
	b := fvm.stack[fvm.n-1]
	fvm.Push(b)
	fvm.Push(a)
}

func (fvm *ForthVM) drp() {
	fvm.Pop()
}

func (fvm *ForthVM) swp() {
	a := fvm.stack[fvm.n]
	fvm.stack[fvm.n] = fvm.stack[fvm.n-1]
	fvm.stack[fvm.n-1] = a
}

func (fvm *ForthVM) tr() {
	fvm.Rpush(fvm.Pop())
}

func (fvm *ForthVM) fr() {
	fvm.Push(fvm.Rpop())
}

func (fvm *ForthVM) rf() {
	fvm.Push(fvm.rstack[fvm.rn])
}

func (fvm *ForthVM) ttr() {
	a := fvm.Pop()
	b := fvm.Pop()

	fvm.Rpush(b)
	fvm.Rpush(a)
}

func (fvm *ForthVM) tfr() {
	a := fvm.Rpop()
	b := fvm.Rpop()

	fvm.Push(b)
	fvm.Push(a)
}

func (fvm *ForthVM) trf() {
	fvm.Push(fvm.rstack[fvm.rn-1])
	fvm.Push(fvm.rstack[fvm.rn])
}

func (fvm *ForthVM) GetString() string {
	value := fvm.Pop()
	length := fvm.mem[value]
	var builder strings.Builder

	for i := int64(0); i < length; i++ {
		err := builder.WriteByte(byte(fvm.mem[value+1+i]))

		if err != nil {
			log.Fatal(err)
		}
	}

	return builder.String()
}

func (fvm *ForthVM) SetString(str string) {
	addr := fvm.Pop()
	length := int64(len(str))

	fvm.mem[addr] = length

	for i := int64(0); i < length; i++ {
		fvm.mem[addr+1+i] = int64(str[i])
	}
}

func (fvm *ForthVM) GetMemory() *[1000]int64 {
	return &fvm.mem
}

func (fvm *ForthVM) sys() {
	syscall := fvm.Pop()

	switch syscall {
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
		// dest-addr name-addr readfile
		name := fvm.GetString()
		content, err := os.ReadFile(name)

		if err != nil {
			log.Fatal(err)
		}

		fvm.SetString(string(content))
	case 6:
		// read memory from image
		// name-addr readimage
		name := fvm.GetString()
		content, err := os.ReadFile(name)

		if err != nil {
			log.Fatal(err)
		}

		buf := bytes.NewReader(content)
		err = binary.Read(buf, binary.LittleEndian, &fvm.mem)

		if err != nil {
			log.Fatal(err)
		}
	case 7:
		// write memory into image
		// name-addr writeimage
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, fvm.mem)

		if err != nil {
			log.Fatal(err)
		}

		name := fvm.GetString()
		err = os.WriteFile(name, buf.Bytes(), 0666)

		if err != nil {
			log.Fatal(err)
		}
	case 8:
		nbytes := fvm.Pop()
		buf := make([]byte, nbytes)
		if n, err := os.Stdin.Read(buf); err != nil {
			fvm.SetString("")
		} else {
			str := string(buf[:n])
			fvm.SetString(str)
		}
	case 9:
		v := fvm.Pop()
		if v != 0 {
			ShowByteCode = true
		} else {
			ShowByteCode = false
		}
	default:
		if fvm.Sysfunc != nil {
			fvm.Sysfunc(VM(fvm), syscall)
		} else {
			log.Fatalf("ERROR: sys() - Unknown call \"%d\"\n", syscall)
		}
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
	LV
	L
	LF
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
)

type Cell struct {
	cmd        int
	arg        int64
	argf       float64
	argStr     string
	localIndex int
}

func (c Cell) String() string {
	return fmt.Sprintf("{cmd:%d, arg:%d, argStr:%s}", c.cmd, c.arg, c.argStr)
}

type Code struct {
	cells     []Cell         // actual code
	labels    map[string]int // labels indices of NOP and SUB
	numLocals int            // number of locals
	posMain   int            // position of MAIN
}

// (SUB xx ... END)* MAIN ... STP delimited by semicolon
func parseCode(codeStr string) *Code {
	code := new(Code)
	code.labels = make(map[string]int)

	cmds := strings.Split(codeStr, ";")
	cells := make([]Cell, 0, len(cmds)+1)
	locals := NewStack()

	for pos, cmd := range cmds {
		if cmd == "" {
			//fmt.Println("EMPTY")
			continue
		}
		//fmt.Println(cmd)
		scmd := strings.Split(cmd, " ")

		if len(scmd) == 2 && scmd[0][0] == '#' {
			// NOP
			code.labels[scmd[0]] = pos
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
				code.posMain = pos
				cells = append(cells, Cell{cmd: MAIN})
			case "LCTX":
				cells = append(cells, Cell{cmd: LCTX})
			case "LSET":
				if !locals.Contains(scmd[1]) {
					locals.Push(scmd[1])
				}
				cells = append(cells, Cell{cmd: LSET, argStr: scmd[1], localIndex: locals.GetIndex(scmd[1])})
			case "LDEF":
				if !locals.Contains(scmd[1]) {
					locals.Push(scmd[1])
				}
				cells = append(cells, Cell{cmd: LDEF, argStr: scmd[1], localIndex: locals.GetIndex(scmd[1])})
			case "LCL":
				if !locals.Contains(scmd[1]) {
					locals.Push(scmd[1])
				}
				cells = append(cells, Cell{cmd: LCL, argStr: scmd[1], localIndex: locals.GetIndex(scmd[1])})
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
			default:
				log.Fatalf("ERROR: Unknown command \"%s\"\n", cmd)
			}
		}
	}

	code.cells = cells
	code.numLocals = locals.Len()

	return code
}

// (SUB xx ... END)* MAIN ... STP delimited by semicolon
func (fvm *ForthVM) Run(codeStr string) {
	codeData := parseCode(codeStr)

	fvm.ln = -1
	fvm.l_len = codeData.numLocals
	fvm.lstack = make([]Local, fvm.l_len*100)

	done := false
	fvm.rn = -1

	numCmds := int64(0)
	start := time.Now()

	for progPtr := codeData.posMain + 1; !done; progPtr++ {
		numCmds++
		command := codeData.cells[progPtr]

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
			progPtr = codeData.labels[command.argStr] - 1
		case JIN:
			if fvm.jin() {
				progPtr = codeData.labels[command.argStr] - 1
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
		case ADF:
			fvm.adf()
		case SBF:
			fvm.sbf()
		case MLF:
			fvm.mlf()
		case DVF:
			fvm.dvf()
		case PRF:
			fvm.prf()
		case LSF:
			fvm.lsf()
		case GRF:
			fvm.grf()
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
			fvm.Push(command.arg)
		case LF:
			fvm.Fpush(command.argf)
		case STR:
			fvm.str()
		case SYS:
			fvm.sys()
		case STP:
			done = true
		case SUB:
			// pass
		case END:
			progPtr = int(fvm.Rpop())
		case MAIN:
			// pass
		case LCTX:
			fvm.lctx()
		case LSET:
			fvm.lset(command.localIndex)
		case LDEF:
			fvm.ldef(command.localIndex)
		case LCL:
			fvm.lcl(command.localIndex)
		case LCLR:
			fvm.lclr()
		case CALL:
			fvm.Rpush(int64(progPtr))
			progPtr = codeData.labels[command.argStr]
		case REF:
			fvm.Push(int64(codeData.labels[command.argStr]))
		case EXC:
			fvm.Rpush(int64(progPtr))
			progPtr = int(fvm.Pop())
		case PCK:
			fvm.pick()
		case NRT:
			fvm.nrot()
		case TR:
			fvm.tr()
		case FR:
			fvm.fr()
		case RF:
			fvm.rf()
		case TTR:
			fvm.ttr()
		case TFR:
			fvm.tfr()
		case TRF:
			fvm.trf()
		default:
			log.Fatalf("ERROR: Unknown command %v\n", command)
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("\n\nexecution time: %s\nNumber of Cmds: %d\nSpeed: %f cmd/ns", elapsed, numCmds, float64(numCmds)/float64(elapsed.Nanoseconds()))
}
