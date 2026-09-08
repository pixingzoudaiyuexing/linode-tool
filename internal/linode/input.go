package linode

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// Prompter owns one buffered reader so piped and interactive input behave consistently.
type Prompter struct {
	in     io.Reader
	out    io.Writer
	reader *bufio.Reader
}

func NewPrompter(in io.Reader, out io.Writer) *Prompter {
	return &Prompter{in: in, out: out, reader: bufio.NewReader(in)}
}

func (p *Prompter) Read(prompt string) (string, error) {
	if _, err := fmt.Fprint(p.out, prompt); err != nil {
		return "", err
	}
	value, err := p.reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	value = strings.TrimSpace(value)
	if errors.Is(err, io.EOF) && value == "" {
		return "", io.EOF
	}
	return value, nil
}

func (p *Prompter) ReadPositiveInt(prompt string) (int, error) {
	for {
		value, err := p.Read(prompt)
		if err != nil {
			return 0, err
		}
		count, err := strconv.Atoi(value)
		if err == nil && count > 0 {
			return count, nil
		}
		fmt.Fprintln(p.out, "请输入大于 0 的整数")
	}
}

func (p *Prompter) ReadChoice(prompt string, max int) (int, error) {
	for {
		choice, err := p.ReadPositiveInt(prompt)
		if err != nil {
			return 0, err
		}
		if choice <= max {
			return choice, nil
		}
		fmt.Fprintf(p.out, "请输入 1-%d 之间的序号\n", max)
	}
}

func (p *Prompter) ReadPassword(prompt string) (string, error) {
	for {
		var password string
		var err error
		if inputFile, ok := p.in.(*os.File); ok && term.IsTerminal(int(inputFile.Fd())) {
			if _, err = fmt.Fprint(p.out, prompt); err == nil {
				var passwordBytes []byte
				passwordBytes, err = term.ReadPassword(int(inputFile.Fd()))
				password = string(passwordBytes)
				fmt.Fprintln(p.out)
			}
		} else {
			password, err = p.Read(prompt)
		}
		if err != nil {
			return "", err
		}
		if password != "" {
			return password, nil
		}
		fmt.Fprintln(p.out, "Root Password 不能为空")
	}
}
