package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func readInstructions(filename string, hal *HAL_INTERN) {
	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		log.Fatalf("failed to open")
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	lineCount := 0
	for scanner.Scan() {
		args := strings.Split(scanner.Text(), " ")

		command := args[0]
		data_s := ""
		data_i := 0
		data_f := 0.0
		if len(args) > 1 {
			data_s = args[1]
			data_i, _ = strconv.Atoi(data_s)
			data_f, _ = strconv.ParseFloat(data_s, 64)
		}
		instruction := Instruction{command: command, data_s: data_s, data_f: data_f, data_i: data_i}
		hal.instructions[lineCount] = instruction
		lineCount++
		//fmt.Println(hal.instructions)
	}
}

func create_hal_instance() HAL_INTERN {
	mmu := create_mmu_instance()
	tmp := HAL_INTERN{START: true, accumulator: 0.00, processCounter: 0, mmu: mmu, instructions: make([]Instruction, 200), connections: make([]chan string, 5), stdio: Std{in: nil, out: nil}, waitGroup: nil}
	for i, _ := range tmp.connections {
		tmp.connections[i] = make(chan string)
	}
	return tmp
}

func (hal *HAL_INTERN) alive() bool {
	if strings.EqualFold(hal.instructions[hal.processCounter].command, "STOP") {
		hal.STOP = true
		hal.START = false
	}
	if hal.processCounter >= 200 {
		hal.STOP = true
		hal.START = false
	}
	if strings.EqualFold(hal.instructions[hal.processCounter].command, "START") {
		hal.START = true
		hal.STOP = false
	}
	return hal.START && !hal.STOP
}



func getUserInput(port string) float64 {
	fmt.Println("Eingabe-Port: " + port)
	var input float64
	fmt.Scanln(&input)
	return input
}
func executeCommand(instruction Instruction, hal *HAL_INTERN) {

	switch instruction.command {
	case "STOP":
		hal.STOP = true
	case "IN":
		IN(hal, instruction.data_i)
	case "OUT":
		OUT(hal, instruction.data_i)
	case "STORE":
		STORE(hal, instruction.data_i)
	case "LOAD":
		LOAD(hal, instruction.data_i)
	case "LOADNUM":
		LOADNUM(hal, instruction.data_f)
	case "JUMP":
		JUMP(hal, instruction.data_i)
	case "JUMPNEG":
		JUMPNEG(hal, instruction.data_i)
	case "JUMPPOS":
		JUMPPOS(hal, instruction.data_i)
	case "JUMPNULL":
		JUMPNULL(hal, instruction.data_i)
	case "ADD":
		ADD(hal, instruction.data_i)
	case "ADDNUM":
		ADDNUM(hal, instruction.data_f)
	case "SUB":
		SUB(hal, instruction.data_i)
	case "SUBNUM":
		SUBNUM(hal, instruction.data_f)
	case "MUL":
		MUL(hal, instruction.data_i)
	case "MULNUM":
		MULNUM(hal, instruction.data_f)
	case "DIV":
		DIV(hal, instruction.data_i)
	case "DIVNUM":
		DIVNUM(hal, instruction.data_f)
	case "LOADIND":
		LOADIND(hal, instruction.data_i)
	case "STOREIND":
		STOREIND(hal, instruction.data_i)
	case "DUMPREG":
		DUMPREG(hal)
	case "DUMPPROG":
		DUMPPROG(hal)
	}
	hal.processCounter++
}



func executeHAL(hal *HAL_INTERN) {
	defer hal.waitGroup.Done()
	readInstructions(hal.scFileName, hal)
	startTime := time.Now()
	DEBUG := true

	if hal.stdio.out != nil {
		go stdout(hal)
	}
	if hal.stdio.in != nil {
		go stdin(hal)
	}

	for hal.alive() {
		pc := &(hal.processCounter)
		executeCommand(hal.instructions[*pc], hal)
	}
	if DEBUG {
		fmt.Println(hal.scFileName)
		fmt.Println(hal.STOP)
		fmt.Println("-------")
	}
	endTime := time.Now()
	totalTime := endTime.Sub(startTime)
	fmt.Println("Elapsed Time: ", totalTime)

}

/*
func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: " + os.Args[0] + " <halcode file>")
		os.Exit(0)
	}
	var filename = os.Args[1]
	hal := create_hal_instance()
	readInstructions(filename, &hal)
	executeHAL(&hal)
}
*/
