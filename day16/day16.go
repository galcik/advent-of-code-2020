package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Interval struct {
	start, end int
}

type Field struct {
	rules []Interval
}

func (field *Field) canBeValidValue(val int) bool {
	for _, interval := range field.rules {
		if interval.start <= val && val <= interval.end {
			return true
		}
	}

	return false
}

type Ticket []int

type Puzzle struct {
	fields        map[string]Field
	ticket        Ticket
	nearbyTickets []Ticket
}

func readPuzzle(filename string) Puzzle {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	puzzle := Puzzle{fields: map[string]Field{}}

	scanner := bufio.NewScanner(file)
	// fields
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			break
		}

		fieldName, field := parseField(line)
		puzzle.fields[fieldName] = field
	}

	// ticket
	for scanner.Scan() && scanner.Text() != "your ticket:" {
	}
	scanner.Scan()
	puzzle.ticket = parseTicket(scanner.Text())

	// nearbyTickets
	for scanner.Scan() && scanner.Text() != "nearby tickets:" {
	}
	for scanner.Scan() {
		puzzle.nearbyTickets = append(puzzle.nearbyTickets, parseTicket(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return puzzle
}

func parseField(fieldLine string) (string, Field) {
	fieldName, rulesSpec := divideString(fieldLine, ":")
	var intervals []Interval
	for _, intervalSpec := range strings.Split(rulesSpec, "or") {
		start, end := divideString(intervalSpec, "-")
		startVal, _ := strconv.Atoi(start)
		endVal, _ := strconv.Atoi(end)
		intervals = append(intervals, Interval{start: startVal, end: endVal})
	}

	return fieldName, Field{rules: intervals}
}

func divideString(s string, sep string) (string, string) {
	parsed := strings.Split(s, sep)
	if len(parsed) != 2 {
		panic("Too many splits")
	}
	return strings.TrimSpace(parsed[0]), strings.TrimSpace(parsed[1])
}

func parseTicket(ticketSpec string) Ticket {
	var values []int
	for _, strVal := range strings.Split(ticketSpec, ",") {
		val, _ := strconv.Atoi(strVal)
		values = append(values, val)
	}
	return values
}

func getTicketScanningErrorRate(puzzle Puzzle) int {
	errorRate := 0
	for _, ticket := range puzzle.nearbyTickets {
		for _, val := range ticket {
			if !validForAnyField(val, puzzle.fields) {
				errorRate += val
			}
		}
	}

	return errorRate
}

func validForAnyField(val int, fields map[string]Field) bool {
	for _, field := range fields {
		if field.canBeValidValue(val) {
			return true
		}
	}

	return false
}

func getValidTickets(puzzle Puzzle) []Ticket {
	var validTickets []Ticket
	for _, ticket := range puzzle.nearbyTickets {
		isValidTicket := true
		for _, val := range ticket {
			if !validForAnyField(val, puzzle.fields) {
				isValidTicket = false
				break
			}
		}

		if isValidTicket {
			validTickets = append(validTickets, ticket)
		}
	}

	validTickets = append(validTickets, puzzle.ticket)
	return validTickets
}

func detectOrder(puzzle Puzzle) []string {
	var validTickets = getValidTickets(puzzle)

	fieldNames := make([]string, 0, len(puzzle.fields))
	fields := make([]Field, 0, len(puzzle.fields))
	for fieldName, field := range puzzle.fields {
		fieldNames = append(fieldNames, fieldName)
		fields = append(fields, field)
	}

	ticketPosField := make([][]bool, len(puzzle.ticket))
	for i := 0; i < len(ticketPosField); i++ {
		ticketPosField[i] = make([]bool, len(fields))
		for j := 0; j < len(fields); j++ {
			ticketPosField[i][j] = true
		}
	}

	for _, ticket := range validTickets {
		for ticketPos, val := range ticket {
			for fieldIdx, field := range fields {
				ticketPosField[ticketPos][fieldIdx] = ticketPosField[ticketPos][fieldIdx] && field.canBeValidValue(val)
			}
		}
	}

	type pair struct {
		idx, count int
	}

	searchOrder := make([]pair, 0, len(fields))
	for idx, fieldCandidates := range ticketPosField {
		count := 0
		for _, val := range fieldCandidates {
			if val {
				count++
			}
		}
		searchOrder = append(searchOrder, pair{idx, count})
	}
	sort.Slice(searchOrder, func(i, j int) bool { return searchOrder[i].count < searchOrder[j].count })

	stopComputation := false
	permutation := make([]int, len(fields))
	inPermutation := make([]bool, len(fields))

	var genPerm func(int)
	genPerm = func(fromIdx int) {
		if stopComputation {
			return
		}

		if fromIdx == len(fields) {
			stopComputation = true
			return
		}

		ticketIdx := searchOrder[fromIdx].idx
		for fieldIdx, isCandidate := range ticketPosField[ticketIdx] {
			if isCandidate && !stopComputation && !inPermutation[fieldIdx] {
				inPermutation[fieldIdx] = true
				permutation[ticketIdx] = fieldIdx
				genPerm(fromIdx + 1)
				inPermutation[fieldIdx] = false
			}
		}
	}

	genPerm(0)
	if !stopComputation {
		panic("No valid permutation of fields")
	}

	orderedFields := make([]string, len(fields))
	for idx, fieldIdx := range permutation {
		orderedFields[idx] = fieldNames[fieldIdx]
	}

	return orderedFields
}

func main() {
	puzzle := readPuzzle("day16/input.txt")
	fmt.Println(getTicketScanningErrorRate(puzzle))
	orderedFields := detectOrder(puzzle)
	solution := 1
	for idx, field := range orderedFields {
		if strings.HasPrefix(field, "departure ") {
			solution *= puzzle.ticket[idx]
		}
	}
	fmt.Println(solution)
}
