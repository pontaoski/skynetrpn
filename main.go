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

func main() {
	// make a reader that reads words from stdin
	b := bufio.NewScanner(os.Stdin)
	b.Split(bufio.ScanWords)

	s := new(Stack)

	// map of names to functions
	functions := map[string]func(*Stack){
		"+": func(s *Stack) {
			a := s.Pop().(int)
			b := s.Pop().(int)
			s.Push(a + b)
		},
		"-": func(s *Stack) {
			a := s.Pop().(int)
			b := s.Pop().(int)
			s.Push(b - a)
		},
		"*": func(s *Stack) {
			a := s.Pop().(int)
			b := s.Pop().(int)
			s.Push(a * b)
		},
		"/": func(s *Stack) {
			a := s.Pop().(int)
			b := s.Pop().(int)
			s.Push(b / a)
		},
		"%": func(s *Stack) {
			a := s.Pop().(int)
			b := s.Pop().(int)
			s.Push(b % a)
		},
		"^": func(s *Stack) {
			a := s.Pop().(int)
			b := s.Pop().(int)
			s.Push(b ^ a)
		},
		"dup": func(s *Stack) {
			s.Push(s.Top())
		},
		"pop": func(s *Stack) {
			s.Pop()
		},
		"swap": func(s *Stack) {
			a := s.Pop().(int)
			b := s.Pop().(int)
			s.Push(a)
			s.Push(b)
		},
		".": func(s *Stack) {
			fmt.Println(s.Pop())
		},
	}

	// read a stream of words from the scanner
	for b.Scan() {
		// get the word from the scanner
		word := b.Text()

		// if the word is in our functions dict, call the function
		if f, ok := functions[word]; ok {
			f(s)
			continue
		}

		// if the received word is a number, push it to the stack
		if num, err := strconv.Atoi(word); err == nil {
			s.Push(num)
			continue
		}

		// if the word starts with a colon, begin reading words until a semicolon is found,
		// at which it is compiled into a function inserted into the functions dict
		if word[0] == ':' {
			// get the function name
			name := word[1:]
			// get the function body
			words := []string{}
			for b.Scan() {
				if b.Text() == ";" {
					break
				}
				words = append(words, b.Text())
			}
			func(wrds []string) {
				// compile the function
				functions[name] = func(s *Stack) {
					for _, wrd := range wrds {
						// if the word is in our functions dict, call the function
						if f, ok := functions[wrd]; ok {
							f(s)
							continue
						}

						// if the received word is a number, push it to the stack
						if num, err := strconv.Atoi(wrd); err == nil {
							s.Push(num)
							continue
						}

						panic("bad word" + wrd)
					}
				}
			}(words)
			continue
		}
	}
}
