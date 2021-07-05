package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

// this struct provides push and pop methods for a stack
type Stack struct {
	data []interface{}
}

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
	input     *bufio.Scanner
	functions map[string]func(*state)
}

func newState() *state {
	s := &state{
		stack: &Stack{},
		input: bufio.NewScanner(os.Stdin),
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
				a := s.stack.Pop().(int)
				b := s.stack.Pop().(int)
				s.stack.Push(a)
				s.stack.Push(b)
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
				s.stack.Push(word)
			},
			"sigil:&": func(s *state) {
				word := s.stack.Pop().(string)
				if fn, ok := s.functions[word]; ok {
					s.stack.Push(fn)
				} else {
					panic("unknown word: " + word)
				}
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

func main() {
	state := newState()
	state.run()
}
