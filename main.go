package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// this struct provides push and pop methods for a stack
type Stack struct {
	data []interface{}
}

type listMarker struct{}

func (s *Stack) Push(x interface{}) {
	s.data = append(s.data, x)
}

func (s *Stack) Pop() interface{} {
	if len(s.data) == 0 {
		return nil
	}
	x := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return x
}

func (s *Stack) Top() interface{} {
	return s.data[len(s.data)-1]
}

type state struct {
	stack     *Stack
	altStack  *Stack
	input     *bufio.Scanner
	variables map[string]interface{}
	functions map[string]func(*state)
}

func newState() *state {
	s := &state{
		stack:     &Stack{},
		altStack:  &Stack{},
		variables: map[string]interface{}{},
		input:     bufio.NewScanner(os.Stdin),
		functions: map[string]func(*state){
			"+": func(s *state) {
				a := s.stack.Pop().(int)
				b := s.stack.Pop().(int)
				s.stack.Push(a + b)
			},
			"-": func(s *state) {
				a := s.stack.Pop().(int)
				b := s.stack.Pop().(int)
				s.stack.Push(b - a)
			},
			"*": func(s *state) {
				a := s.stack.Pop().(int)
				b := s.stack.Pop().(int)
				s.stack.Push(a * b)
			},
			"/": func(s *state) {
				a := s.stack.Pop().(int)
				b := s.stack.Pop().(int)
				s.stack.Push(b / a)
			},
			"%": func(s *state) {
				a := s.stack.Pop().(int)
				b := s.stack.Pop().(int)
				s.stack.Push(b % a)
			},
			"^": func(s *state) {
				a := s.stack.Pop().(int)
				b := s.stack.Pop().(int)
				s.stack.Push(b ^ a)
			},
			"dup": func(s *state) {
				s.stack.Push(s.stack.Top())
			},
			"pop": func(s *state) {
				s.stack.Pop()
			},
			"swap": func(s *state) {
				a := s.stack.Pop()
				b := s.stack.Pop()
				s.stack.Push(a)
				s.stack.Push(b)
			},
			"->": func(s *state) {
				s.altStack.Push(s.stack.Pop())
			},
			"<-": func(s *state) {
				s.stack.Push(s.altStack.Pop())
			},
			".": func(s *state) {
				fmt.Println(s.stack.Pop())
			},
			"sigil:#": func(s *state) {
				word := s.stack.Pop().(string)
				// if the received word is a number, push it to the stack
				if num, err := strconv.Atoi(word); err == nil {
					s.stack.Push(num)
				} else {
					panic("bad number")
				}
			},
			"sigil:'": func(s *state) {
				word := s.stack.Pop().(string)
				s.stack.Push(strings.ReplaceAll(word, "_", " "))
			},
			"sigil:&": func(s *state) {
				word := s.stack.Pop().(string)
				if fn, ok := s.functions[word]; ok {
					s.stack.Push(fn)
				} else {
					panic("unknown word: " + word)
				}
			},
			"sigil:!": func(s *state) {
				word := s.stack.Pop().(string)
				val := s.stack.Pop()
				s.variables[word] = val
			},
			"sigil:@": func(s *state) {
				word := s.stack.Pop().(string)
				s.stack.Push(s.variables[word])
			},
			"call": func(s *state) {
				fn := s.stack.Pop().(func(*state))
				fn(s)
			},
			"def": func(s *state) {
				fn := s.stack.Pop().(func(*state))
				word := s.stack.Pop().(string)
				s.functions[word] = fn
			},
			"true": func(s *state) {
				s.stack.Push(true)
			},
			"false": func(s *state) {
				s.stack.Push(false)
			},
			"if": func(s *state) {
				fn := s.stack.Pop().(func(*state))
				cond := s.stack.Pop().(bool)
				if cond {
					fn(s)
				}
			},
			"choose": func(s *state) {
				iffalse := s.stack.Pop().(func(*state))
				iftrue := s.stack.Pop().(func(*state))
				cond := s.stack.Pop().(bool)
				if cond {
					iftrue(s)
				} else {
					iffalse(s)
				}
			},
			"while": func(s *state) {
				body := s.stack.Pop().(func(*state))
				cond := s.stack.Pop().(func(*state))

				cond(s)
				for s.stack.Pop().(bool) {
					body(s)
					cond(s)
				}
			},
			">": func(s *state) {
				a := s.stack.Pop().(int)
				b := s.stack.Pop().(int)
				s.stack.Push(b > a)
			},
			"<": func(s *state) {
				a := s.stack.Pop().(int)
				b := s.stack.Pop().(int)
				s.stack.Push(b < a)
			},
			"dbg": func(s *state) {
				fmt.Println(s.stack)
			},
			"{": func(s *state) {
				s.stack.Push(listMarker{})
			},
			"}": func(s *state) {
				var items []interface{}

				for {
					switch t := s.stack.Pop().(type) {
					case listMarker:
						var reversedItems []interface{}
						for i := len(items) - 1; i >= 0; i-- {
							reversedItems = append(reversedItems, items[i])
						}
						s.stack.Push(reversedItems)
						return
					default:
						items = append(items, t)
					}
				}
			},
			"sigil:/": func(s *state) {
				word := s.stack.Pop().(string)
				arr := s.stack.Pop().([]interface{})
				num, err := strconv.Atoi(word)
				if err == nil {
					s.stack.Push(arr[num])
				} else {
					panic("bad number")
				}
			},
			"nth": func(s *state) {
				num := s.stack.Pop().(int)
				arr := s.stack.Pop().([]interface{})
				s.stack.Push(arr[num])
			},
			"set-nth": func(s *state) {
				val := s.stack.Pop()
				num := s.stack.Pop().(int)
				arr := s.stack.Pop().([]interface{})

				arr[num] = val
				s.stack.Push(arr)
			},
			"new-arr": func(s *state) {
				s.stack.Push(make([]interface{}, s.stack.Pop().(int)))
			},
			"push-arr": func(s *state) {
				val := s.stack.Pop()
				arr := s.stack.Pop().([]interface{})

				arr = append(arr, val)
				s.stack.Push(arr)
			},
			"arr-pop-front": func(s *state) {
				arr := s.stack.Pop().([]interface{})
				s.stack.Push(arr[1:])
				s.stack.Push(arr[0])
			},
			"arr-pop-back": func(s *state) {
				arr := s.stack.Pop().([]interface{})
				s.stack.Push(arr[:len(arr)-1])
				s.stack.Push(arr[len(arr)-1])
			},
			"arr-spill": func(s *state) {
				arr := s.stack.Pop().([]interface{})
				for _, it := range arr {
					s.stack.Push(it)
				}
			},
			"for-each": func(s *state) {
				combinator := s.stack.Pop().(func(*state))
				array := s.stack.Pop().([]interface{})

				for _, it := range array {
					s.stack.Push(it)
					combinator(s)
				}
			},
			"curry": func(s *state) {
				combinator := s.stack.Pop().(func(*state))
				value := s.stack.Pop()

				s.stack.Push(func(s *state) {
					s.stack.Push(value)
					combinator(s)
				})
			},
			"concat": func(s *state) {
				a := s.stack.Pop().(string)
				b := s.stack.Pop().(string)

				s.stack.Push(b + a)
			},
			"0-through": func(s *state) {
				comb := s.stack.Pop().(func(*state))
				through := s.stack.Pop().(int)

				for i := 0; i <= through; i++ {
					s.stack.Push(i)
					comb(s)
				}
			},
			"eq?": func(s *state) {
				a := s.stack.Pop()
				b := s.stack.Pop()

				s.stack.Push(reflect.DeepEqual(a, b))
			},
		},
	}
	s.input.Split(bufio.ScanWords)
	return s
}

func (s *state) handleWord(w string) func() {
	f, ok := s.functions["sigil:"+string(w[0])]
	if ok {
		return func() {
			s.stack.Push(w[1:])
			f(s)
		}
	}
	f, ok = s.functions[w]
	if !ok {
		panic("unknown word: " + w)
	}
	return func() {
		f(s)
	}
}

func (s *state) parseLambda() func(s *state) {
	fns := []func(*state){}
	for s.input.Scan() && s.input.Text() != "]" {
		w := s.input.Text()
		f, ok := s.functions["sigil:"+string(w[0])]
		if ok {
			fns = append(fns, func(s *state) {
				s.stack.Push(w[1:])
				f(s)
			})
			continue
		}
		if w == "[" {
			lam := s.parseLambda()
			fns = append(fns, func(s *state) {
				s.stack.Push(lam)
			})
			continue
		}
		f, ok = s.functions[w]
		if !ok {
			panic("unknown word: " + w)
		}
		fns = append(fns, f)
	}
	return func(s *state) {
		for _, f := range fns {
			f(s)
		}
	}
}

func (s *state) run() {
	for s.input.Scan() {
		word := s.input.Text()
		switch word {
		case "[":
			s.stack.Push(s.parseLambda())
		default:
			f := s.handleWord(word)
			f()
		}
	}
}

func (s *state) evaluate(prg string) {
	old := s.input

	s.input = bufio.NewScanner(strings.NewReader(prg))
	s.input.Split(bufio.ScanWords)
	s.run()

	s.input = old
}

const preconception = `
'sigil:( [ pop ] def
'c-> [ dup -> ] def
'<-c [ <- dup -> ] def
'<>swap [ <- <- swap -> -> ] def

'adt [
	[
		dup /0 !name
		/1 !size


		@name @size [
			!size !name

			@size #1 + !arr-size

			@arr-size new-arr

			#0 @name set-nth !arr

			@size #1 - [
				#1 + !index
				!value

				@arr @index @value set-nth
			] 0-through

		] curry curry

		@name swap def

	] for-each
	pop
]
def

'match [
	!currently-matching
] def

'||
[
	!closure !tag

	@currently-matching
	arr-pop-front @tag eq?

	(iftrue)
		[ arr-spill @closure call ]
	(iffalse)
		[ pop ]

	choose
]
def

'possibly {
	{ 'is #1 }
	{ 'isn't #0 }
} adt

'Willkommen_bei_SkynetRPN! .
`

func main() {
	state := newState()
	state.evaluate(preconception)
	state.run()
}
