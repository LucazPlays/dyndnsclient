package main
import (
	"fmt"
	"os"
	"golang.org/x/term"
)
type terminalRW struct {
	in  *os.File
	out *os.File
}
func (trw terminalRW) Read(p []byte) (n int, err error) { return trw.in.Read(p) }
func (trw terminalRW) Write(p []byte) (n int, err error) { return trw.out.Write(p) }

func main() {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer term.Restore(fd, oldState)

	t := term.NewTerminal(terminalRW{os.Stdin, os.Stdout}, "> ")
	line, _ := t.ReadLine()
	fmt.Printf("\r\nYou typed: %s\r\n", line)
}
