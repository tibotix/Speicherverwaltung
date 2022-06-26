package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type HALOS struct {
	processors []HAL_INTERN
}

func splitDetails(info string) (int, int) {
	splitter := strings.Split(info, ":")
	r1, _ := strconv.Atoi(splitter[0])
	r2, _ := strconv.Atoi(splitter[1])
	return r1, r2
}
func connectHALS(halos *HALOS, info1 string, info2 string) {
	idSource, channelSource := splitDetails(info1)
	idDest, channelDest := splitDetails(info2)

	destHAL := halos.processors[idDest]
	sourceHAL := halos.processors[idSource]

	destHAL.connections[channelDest] = sourceHAL.connections[channelSource]
}
func connectStd(halos *HALOS, stdType string, info string) {
	idDest, channelDest := splitDetails(info)

	if strings.EqualFold(stdType, "stdin") {
		halos.processors[idDest].stdio.in = &halos.processors[idDest].connections[channelDest]
	} else if strings.EqualFold(stdType, "stdout") {
		halos.processors[idDest].stdio.out = &halos.processors[idDest].connections[channelDest]
	}
}
func initOS(lines []string) *HALOS {
	halos := HALOS{processors: make([]HAL_INTERN, 0, 50)}
	f_procs := false
	f_connections := false
	f_stdin := false
	f_stdout := false

	for _, line := range lines {
		if strings.Contains(line, "HAL-Prozessoren:") {
			f_procs = true
			f_connections = false
			f_stdin = false
			f_stdout = false
			continue
		}
		if strings.Contains(line, "HAL-Verbindungen:") {
			f_connections = true
			f_procs = false
			f_stdin = false
			f_stdout = false
			continue
		}
		if strings.Contains(line, "HAL-Stdin:") {
			f_connections = false
			f_procs = false
			f_stdin = true
			f_stdout = false
			continue
		}
		if strings.Contains(line, "HAL-Stdout:") {
			f_connections = false
			f_procs = false
			f_stdin = false
			f_stdout = true
			continue
		}

		if f_procs {
			splitter := strings.Split(line, " ")
			id, _ := strconv.Atoi(splitter[0])
			scFile := splitter[1]

			newHAL := create_hal_instance()
			newHAL.scFileName = scFile
			halos.processors = append(halos.processors, create_hal_instance())
			halos.processors[id] = newHAL
		} else if f_connections {
			splitter := strings.Split(line, ">")
			connectHALS(&halos, splitter[0], splitter[1])
		} else if f_stdin {
			splitter := strings.Split(line, ">")
			connectStd(&halos, splitter[0], splitter[1])
		} else if f_stdout {
			splitter := strings.Split(line, ">")
			connectStd(&halos, splitter[1], splitter[0])
		}
	}
	return &halos
}
func getConfig(filename string) []string {
	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		log.Fatalf("failed to open")
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var output []string
	for scanner.Scan() {
		output = append(output, scanner.Text())
	}
	return output
}

func stdin(hal *HAL_INTERN) {
	for hal.alive() {
		var input string
		fmt.Scanln(&input)
		(*hal.stdio.in) <- input
	}
}
func stdout(hal *HAL_INTERN) {
	for hal.alive() {
		output := <-(*hal.stdio.out)
		fmt.Println("Ausgabe: ", output)
	}
}

func executeOS(halos *HALOS) {
	var waitGroup sync.WaitGroup
	for index := range halos.processors {
		hal := &halos.processors[index]
		waitGroup.Add(1)
		hal.waitGroup = &waitGroup
		go executeHAL(hal)
	}
	waitGroup.Wait()
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: " + os.Args[0] + " <config file>")
		os.Exit(0)
	}
	cfgFile := os.Args[1]
	cfgContent := getConfig(cfgFile)
	halos := initOS(cfgContent)
	executeOS(halos)
	fmt.Scanln()
}
