package main

import (
	"sync"
)

type Instruction struct {
	command string
	data_f  float64
	data_i  int
	data_s  string
}
type Std struct {
	in  *chan float64
	out *chan float64
}

type HAL_INTERN struct {
	START          bool
	STOP           bool
	accumulator    float64
	processCounter int
	register       []float64
	connections    []chan float64
	instructions   []Instruction
	scFileName     string
	stdio          Std
	waitGroup      *sync.WaitGroup
}

func OUT(hal *HAL_INTERN, port int) {
	hal.connections[port] <- hal.accumulator
	//fmt.Println("Ausgabe-Port: " + strconv.Itoa(port) + "\n" + fmt.Sprintf("%f", hal.io[port]))
}
func IN(hal *HAL_INTERN, port int) {
	//hal.io[port] = getUserInput(strconv.Itoa(port))
	hal.accumulator = <-hal.connections[port]
}
func STORE(hal *HAL_INTERN, index int) {
	hal.register[index] = hal.accumulator
}
func LOAD(hal *HAL_INTERN, index int) {
	hal.accumulator = hal.register[index]
}
func LOADNUM(hal *HAL_INTERN, number float64) {
	hal.accumulator = number
}
func JUMP(hal *HAL_INTERN, pc int) {
	hal.processCounter = pc - 1
}
func JUMPNEG(hal *HAL_INTERN, pc int) {
	if hal.accumulator < 0 {
		JUMP(hal, pc)
	}
}
func JUMPPOS(hal *HAL_INTERN, pc int) {
	if hal.accumulator > 0 {
		JUMP(hal, pc)
	}
}
func JUMPNULL(hal *HAL_INTERN, pc int) {
	if hal.accumulator == 0 {
		JUMP(hal, pc)
	}
}
func ADD(hal *HAL_INTERN, index int) {
	hal.accumulator += hal.register[index]
}
func ADDNUM(hal *HAL_INTERN, number float64) {
	hal.accumulator += number
}
func SUB(hal *HAL_INTERN, index int) {
	hal.accumulator -= hal.register[index]
}
func SUBNUM(hal *HAL_INTERN, number float64) {
	hal.accumulator -= number
}
func MUL(hal *HAL_INTERN, index int) {
	hal.accumulator *= hal.register[index]
}
func MULNUM(hal *HAL_INTERN, number float64) {
	hal.accumulator *= number
}
func DIV(hal *HAL_INTERN, index int) {
	hal.accumulator /= hal.register[index]
}
func DIVNUM(hal *HAL_INTERN, number float64) {
	hal.accumulator /= number
}
func LOADIND(hal *HAL_INTERN, index int) {
	hal.accumulator = hal.register[int(hal.register[index])]
}
func STOREIND(hal *HAL_INTERN, index int) {
	hal.register[int(hal.register[index])] = hal.accumulator
}
