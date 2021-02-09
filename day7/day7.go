package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type BagCount struct {
	bagIdx, count int
}

type BagContent []BagCount

type Ruleset struct {
	bagColorIndices map[string]int
	rules           []BagContent
}

func (ruleset *Ruleset) GetBagColorIndex(bagColor string) int {
	if _, exists := ruleset.bagColorIndices[bagColor]; !exists {
		ruleset.bagColorIndices[bagColor] = len(ruleset.rules)
		ruleset.rules = append(ruleset.rules, nil)
	}

	return ruleset.bagColorIndices[bagColor]
}

func (ruleset *Ruleset) DefineRule(bagColor string, bagContent map[string]int) {
	rule := make(BagContent, 0, len(bagContent))
	for bag, count := range bagContent {
		bagIdx := ruleset.GetBagColorIndex(bag)
		rule = append(rule, BagCount{bagIdx, count})
	}

	bagColorIdx := ruleset.GetBagColorIndex(bagColor)
	ruleset.rules[bagColorIdx] = rule
}

func readRules(filename string) Ruleset {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	ruleset := Ruleset{bagColorIndices: make(map[string]int)}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words := strings.Split(scanner.Text(), " ")

		bagColor := words[0] + " " + words[1]
		bagRules := map[string]int{}
		if words[4] != "no" {
			for readIdx := 4; readIdx < len(words); readIdx += 4 {
				bagColor := words[readIdx+1] + " " + words[readIdx+2]
				count, _ := strconv.Atoi(words[readIdx])
				bagRules[bagColor] = count
			}
		}

		ruleset.DefineRule(bagColor, bagRules)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return ruleset
}

func computeContainers(ruleset Ruleset) [][]int {
	containers := make([][]int, len(ruleset.rules))
	for idx, rule := range ruleset.rules {
		for _, bagCount := range rule {
			containers[bagCount.bagIdx] = append(containers[bagCount.bagIdx], idx)
		}
	}
	return containers
}

func countIndirectContainers(ruleset Ruleset, bagColor string) int {
	containers := computeContainers(ruleset)

	visited := make([]bool, len(ruleset.rules))
	searchStack := []int{ruleset.GetBagColorIndex(bagColor)}
	for len(searchStack) > 0 {
		idx := searchStack[len(searchStack)-1]
		searchStack = searchStack[:len(searchStack)-1]

		for _, containerIdx := range containers[idx] {
			if !visited[containerIdx] {
				searchStack = append(searchStack, containerIdx)
				visited[containerIdx] = true
			}
		}
	}

	count := 0
	for _, isVisited := range visited {
		if isVisited {
			count += 1
		}
	}

	return count
}

func countBagsInside(ruleset Ruleset, bagColor string) int {
	containers := computeContainers(ruleset)

	bagSizes := make([]int, len(ruleset.rules))
	insideBagsDone := make([]int, len(ruleset.rules))

	var searchStack []int
	for idx, rule := range ruleset.rules {
		if len(rule) == 0 {
			searchStack = append(searchStack, idx)
		}
	}

	for len(searchStack) > 0 {
		idx := searchStack[len(searchStack)-1]
		searchStack = searchStack[:len(searchStack)-1]

		contentSize := 0
		for _, innerBag := range ruleset.rules[idx] {
			contentSize += innerBag.count + innerBag.count*bagSizes[innerBag.bagIdx]
		}
		bagSizes[idx] = contentSize

		for _, containerIdx := range containers[idx] {
			insideBagsDone[containerIdx] += 1
			if insideBagsDone[containerIdx] == len(ruleset.rules[containerIdx]) {
				searchStack = append(searchStack, containerIdx)
			}
		}
	}

	return bagSizes[ruleset.GetBagColorIndex(bagColor)]
}

func main() {
	ruleset := readRules("day7/input.txt")
	fmt.Println(countIndirectContainers(ruleset, "shiny gold"))
	fmt.Println(countBagsInside(ruleset, "shiny gold"))
}
