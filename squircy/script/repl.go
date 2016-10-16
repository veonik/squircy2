package script

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/peterh/liner"
)

var lastExpressionRegex = regexp.MustCompile(`[a-zA-Z0-9]([a-zA-Z0-9\.]*[a-zA-Z0-9])?\.?$`)

func (m *ScriptManager) Repl() {
	hist := filepath.Join(m.conf.RootPath, ".history_repl")

	cli := liner.NewLiner()
	defer func() {
		if f, err := os.Create(hist); err == nil {
			cli.WriteHistory(f)
			f.Close()
		}
		cli.Close()
	}()
	cli.SetWordCompleter(func(line string, pos int) (head string, completions []string, tail string) {
		lastExpression := lastExpressionRegex.FindString(line[:pos])

		bits := strings.Split(lastExpression, ".")

		first := bits[:len(bits)-1]
		last := bits[len(bits)-1]

		var l []string

		if len(first) == 0 {
			c := m.jsDriver.vm.Context()

			l = make([]string, len(c.Symbols))

			i := 0
			for k := range c.Symbols {
				l[i] = k
				i++
			}
		} else {
			m.jsDriver.vm.Interrupt = make(chan func(), 1)
			r, err := m.jsDriver.vm.Eval(strings.Join(bits[:len(bits)-1], "."))
			if err != nil {
				return line[:pos], []string{}, line[pos:]
			}

			if o := r.Object(); o != nil {
				for _, v := range o.KeysByParent() {
					l = append(l, v...)
				}
			}
		}

		var r []string
		for _, s := range l {
			if strings.HasPrefix(s, last) {
				r = append(r, strings.TrimPrefix(s, last))
			}
		}

		return line[:pos], r, line[pos:]
	})
	if f, err := os.Open(hist); err == nil {
		cli.ReadHistory(f)
		f.Close()
	}
	fmt.Println("Starting javascript REPL...")
	fmt.Println("Type 'exit' and hit enter to exit the REPL.")
	for {
		str, _ := cli.Prompt("repl> ")
		if str == "exit" {
			fmt.Println("Closing REPL...")
			break
		}
		cli.AppendHistory(str)
		v, _ := m.RunUnsafe(Javascript, str)
		fmt.Println(v)
	}
}
