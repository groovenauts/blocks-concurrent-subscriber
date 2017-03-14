package main

import (
	"os"
	"os/exec"
)

type Pattern struct {
	Completed string `json:"completed"`
	Level string     `json:"level"`
	Command []string `json:"command"`
}

func (p *Pattern) execute(msg *Message) error {
	if !p.match(msg) {
		return nil
	}
	cmd, err := p.build(msg)
	if err != nil {
		return err
	}
	return cmd.Run()
}

func (p *Pattern) match(msg *Message) bool {
	return false
}

func (p *Pattern) build(msg *Message) (*exec.Cmd, error) {

	name := p.Command[0]
	args := p.Command[1:] // TODO how should the parameters be passed
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}

type Patterns []*Pattern

func (pa Patterns) execute(msg *Message) error {
	for _, ptn := range pa {
		err := ptn.execute(msg)
		if err != nil {
			return err
		}
	}
	return nil
}
