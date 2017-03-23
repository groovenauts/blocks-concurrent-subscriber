package main

import (
	"io"
	"os"
	"os/exec"

	"github.com/groovenauts/blocks-variable"

	log "github.com/Sirupsen/logrus"
)

var (
	CommandStdout io.Writer = os.Stdout
	CommandStderr io.Writer = os.Stderr
)

type Pattern struct {
	Completed string `json:"completed"`
	Level     string `json:"level"`
	Data      string
	Command   []string `json:"command"`
}

func (p *Pattern) execute(msg *Message) error {
	log.WithFields(log.Fields{"pattern": p}).Debugln("Executing command 0")
	cmd, err := p.build(msg)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"cmd": cmd}).Debugln("Executing command")
	return cmd.Run()
}

func (p *Pattern) match(msg *Message) bool {
	if p.Completed != "" {
		if p.Completed != msg.completed {
			return false
		}
	}
	if p.Level != "" {
		if p.Level != msg.level {
			return false
		}
	}
	return true
}

func (p *Pattern) build(msg *Message) (*exec.Cmd, error) {
	// logAttrs is another object of msgMap.
	// logAttrs is modified for log.
	logAttrs := log.Fields(msg.buildMap())
	msgMap := msg.buildMap()
	v := &bvariable.Variable{Data: msgMap}
	command := []string{}
	for _, part := range p.Command {
		logAttrs["template"] = part
		log.WithFields(logAttrs).Debugln("Expanding variables")
		expanded, err := v.Expand(part)
		if err != nil {
			logAttrs["error"] = err
			logAttrs["template"] = part
			log.WithFields(logAttrs).Errorln("Failed to expand variables")
			return nil, err
		}
		command = append(command, expanded)
	}
	name := command[0]
	args := command[1:]
	cmd := exec.Command(name, args...)
	cmd.Stdout = CommandStdout
	cmd.Stderr = CommandStderr
	return cmd, nil
}

type Patterns []*Pattern

func (pa Patterns) allFor(msg *Message) Patterns {
	result := Patterns{}
	for _, ptn := range pa {
		if ptn.match(msg) {
			result = append(result, ptn)
		}
	}
	return result
}

func (pa Patterns) oneFor(msg *Message) *Pattern {
	for _, ptn := range pa {
		if ptn.match(msg) {
			return ptn
		}
	}
	return nil
}

func (pa Patterns) execute(msg *Message) error {
	matched := pa.allFor(msg)
	log.WithFields(log.Fields{"matched": matched, "msg": msg}).Debugln("Matched patterns")
	for _, ptn := range matched {
		err := ptn.execute(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pa Patterns) executeOne(msg *Message) error {
	ptn := pa.oneFor(msg)
	if ptn == nil {
		return nil
	}
	return ptn.execute(msg)
}
