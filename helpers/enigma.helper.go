package sdk_helper

import (
	"strings"
	"unicode"
)

type (
	rotor struct {
		forward  [26]byte
		backward [26]byte
		position int
	}

	plugboard struct {
		wiring [26]byte
	}

	reflector struct {
		wiring [26]byte
	}

	enigmaMachine struct {
		rotors    []*rotor
		reflector *reflector
		plugboard *plugboard
	}

	IEnigma interface {
		Encode(plainText string) string
		Decode(cipherText string) string
	}

	enigma struct {
		machine enigmaMachine
	}
)

func NewEnigma() IEnigma {
	r1W := "EKMFLGDQVZNTOWYHXUSPAIBRCJ"
	r2W := "AJDKSIRUXBLHWTMCQGZNPYFVOE"
	r3W := "BDFHJLCPRTXVZNYEIWGAKMU SQO"
	refW := "YRUHQSLDPXNGOKMIEBFZCWVJAT"

	createRotor := func(w string) *rotor {
		r := &rotor{}
		for i := 0; i < 26; i++ {
			r.forward[i] = w[i]
			r.backward[w[i]-'A'] = byte(i + 'A')
		}
		return r
	}

	machine := enigmaMachine{
		rotors: []*rotor{
			createRotor(r1W),
			createRotor(r2W),
			createRotor(r3W),
		},
		reflector: &reflector{},
		plugboard: &plugboard{},
	}

	for i := 0; i < 26; i++ {
		machine.reflector.wiring[i] = refW[i]
		machine.plugboard.wiring[i] = byte(i + 'A')
	}

	return &enigma{machine: machine}
}

func (r *rotor) rotate() {
	r.position = (r.position + 1) % 26
}

func (r *rotor) encode(letter byte, forward bool) byte {
	idx := int(letter - 'A')
	shiftIdx := (idx + r.position) % 26

	var out byte
	if forward {
		out = r.forward[shiftIdx]
	} else {
		out = r.backward[shiftIdx]
	}

	res := (int(out-'A') - r.position + 26) % 26
	return byte(res + 'A')
}

func (e *enigmaMachine) process(message string) string {
	var sb strings.Builder
	sb.Grow(len(message))

	for i := 0; i < len(message); i++ {
		char := message[i]
		if char < 'A' || char > 'Z' {
			continue
		}

		for _, r := range e.rotors {
			r.rotate()
		}

		res := e.plugboard.wiring[char-'A']

		for _, r := range e.rotors {
			res = r.encode(res, true)
		}

		res = e.reflector.wiring[res-'A']

		for i := len(e.rotors) - 1; i >= 0; i-- {
			res = e.rotors[i].encode(res, false)
		}

		res = e.plugboard.wiring[res-'A']

		sb.WriteByte(res)
	}

	return sb.String()
}

func (h *enigma) Encode(plainText string) string {
	var cleaned strings.Builder
	cleaned.Grow(len(plainText))

	for _, r := range plainText {
		if !unicode.IsSpace(r) {
			cleaned.WriteRune(unicode.ToUpper(r))
		}
	}

	return h.machine.process(cleaned.String())
}

func (h *enigma) Decode(cipherText string) string {
	for _, r := range h.machine.rotors {
		r.position = 0
	}

	return h.machine.process(cipherText)
}
