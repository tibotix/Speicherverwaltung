package main

import (
	"sync"
	"fmt"
	"strconv"
)

type Instruction struct {
	command string
	data_f  float64
	data_i  int
	data_s  string
}
type Std struct {
	in  *chan string
	out *chan string
}

type HAL_INTERN struct {
	START          bool
	STOP           bool
	accumulator    float64
	processCounter int
	mmu            MMU
	connections    []chan string
	instructions   []Instruction
	scFileName     string
	stdio          Std
	waitGroup      *sync.WaitGroup
}

func OUT(hal *HAL_INTERN, port int) {
	hal.connections[port] <- fmt.Sprintf("%f", hal.accumulator)
}
func IN(hal *HAL_INTERN, port int) {
	f := <-hal.connections[port]
	s, _ := strconv.ParseFloat(f, 64)
	hal.accumulator = s
}
func STORE(hal *HAL_INTERN, index int) {
	hal.mmu.write(index, hal.accumulator)
}
func LOAD(hal *HAL_INTERN, index int) {
	hal.accumulator = hal.mmu.read(index)
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
	hal.accumulator += hal.mmu.read(index)
}
func ADDNUM(hal *HAL_INTERN, number float64) {
	hal.accumulator += number
}
func SUB(hal *HAL_INTERN, index int) {
	hal.accumulator -= hal.mmu.read(index)
}
func SUBNUM(hal *HAL_INTERN, number float64) {
	hal.accumulator -= number
}
func MUL(hal *HAL_INTERN, index int) {
	hal.accumulator *= hal.mmu.read(index)
}
func MULNUM(hal *HAL_INTERN, number float64) {
	hal.accumulator *= number
}
func DIV(hal *HAL_INTERN, index int) {
	hal.accumulator /= hal.mmu.read(index)
}
func DIVNUM(hal *HAL_INTERN, number float64) {
	hal.accumulator /= number
}
func LOADIND(hal *HAL_INTERN, index int) {
	hal.accumulator = hal.mmu.read(int(hal.mmu.read(index)))
}
func STOREIND(hal *HAL_INTERN, index int) {
	hal.mmu.write(int(hal.mmu.read(index)), hal.accumulator)
}
func DUMPREG(hal *HAL_INTERN){
	for idx, register := range hal.mmu.memory {
		payload := fmt.Sprintf("%d : %f", idx, register)
		hal.connections[1] <- payload
	}
}
func DUMPPROG(hal* HAL_INTERN){
	for _, instruction := range hal.instructions {
		hal.connections[1] <- instruction.command + " " + instruction.data_s
	}
}
