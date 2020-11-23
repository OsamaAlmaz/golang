package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Step struct {
	cell Cell

	scale int
}

type Path struct {
	steps []Step
}

type Warehouse struct {
	name   string
	demand int
}

type Factory struct {
	name   string
	supply int
}

type Cell struct {
	factoryIndex, warehouseIndex int

	price  int
	supply int
}

type Table struct {
	warehouses []Warehouse
	factories  []Factory

	// First index = Factory, second index = Warehouse
	cells [][]Cell

	cost int
}

const (
	WAREHOUSES = 0
	COSTS      = 1
	DEMAND     = 2
)

func parseWarehousesData(lineData []string) {

	for i := 1; i < len(lineData)-1; i++ {
		warehouse := Warehouse{}
		warehouse.name = lineData[i]
		table.warehouses = append(table.warehouses, warehouse)
	}

}

func parseCostsData(lineData []string, factoryIndex int) {

	factory := Factory{}

	factory.name = lineData[0]
	factory.supply, _ = strconv.Atoi(lineData[len(lineData)-1])

	table.factories = append(table.factories, factory)

	cellRow := make([]Cell, 0)

	for i := 1; i < len(lineData)-1; i++ {
		cell := Cell{}
		cell.price, _ = strconv.Atoi(lineData[i])
		cell.warehouseIndex = i - 1
		cell.factoryIndex = factoryIndex
		cellRow = append(cellRow, cell)
	}

	table.cells = append(table.cells, cellRow)

}

// Assumes warehouses already created.
func parseDemandData(lineData []string) {

	for i := 1; i < len(lineData); i++ {
		(&table.warehouses[i-1]).demand, _ = strconv.Atoi(lineData[i])
	}

}

func parseDefinitionFile(fileName string) {

	file, err := os.Open(fileName)

	defer file.Close()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)

	var parseMode int

	table = Table{}

	warehouses := make([]Warehouse, 0)
	factories := make([]Factory, 0)

	table.cells = make([][]Cell, 0)

	table.warehouses = warehouses
	table.factories = factories

	factoryIndex := 0

	for scanner.Scan() {

		line := scanner.Text()

		lineData := strings.Fields(line)

		if len(lineData) <= 0 {
			continue
		}

		lineId := strings.ToUpper(lineData[0])

		if lineId == "COSTS" {
			parseMode = WAREHOUSES
		} else if lineId == "DEMAND" {
			parseMode = DEMAND
		} else {
			parseMode = COSTS
		}

		switch parseMode {
		case WAREHOUSES:

			parseWarehousesData(lineData)

			break

		case COSTS:

			parseCostsData(lineData, factoryIndex)
			factoryIndex++

			break

		case DEMAND:

			parseDemandData(lineData)

			break

		}

	}

}

func parseInitialSolutionFile(fileName string) {

	file, err := os.Open(fileName)
	if err != nil {
		//use a print statement 
		panic(err)
	}
	
	defer file.Close()


	scanner := bufio.NewScanner(file)

	factoryIndex := 0

	for scanner.Scan() {

		line := scanner.Text()

		lineData := strings.Fields(line)

		if len(lineData) <= 0 {
			continue
		}

		lineId := strings.ToUpper(lineData[0])

		if lineId == "INITIAL" || lineId == "DEMAND" {
			continue
		}

		for i := 1; i < len(lineData)-1; i++ {

			warehouseIndex := i - 1

			cell := &table.cells[factoryIndex][warehouseIndex]

			supply := lineData[i]

			var err error

			cell.supply, err = strconv.Atoi(supply)

			if err != nil {
				cell.supply = 0
			}

		}

		factoryIndex++

	}

}

func computeInitialSolution() {

}

func marginalCost(cell Cell, result chan Path) {

	initialStep := Step{}

	initialStep.cell = cell
	initialStep.scale = +1

	path := Path{}
	path.steps = []Step{initialStep}

	visitedCells := []Cell{cell}
	currentStep := initialStep

	for {

		if currentStep.scale > 0 {

			factoryIndex := currentStep.cell.factoryIndex

			newStepFound := false

			for warehouseIndex, _ := range table.warehouses {

				cell := table.cells[factoryIndex][warehouseIndex]

				if isCellVisited(cell, visitedCells) {
					continue
				}

				visitedCells = append(visitedCells, cell)

				if cell.supply <= 0 {
					continue
				}

				newStep := Step{}
				newStep.cell = cell
				newStep.scale = currentStep.scale * -1
				path.steps = append(path.steps, newStep)

				currentStep = newStep

				newStepFound = true

				if warehouseIndex == initialStep.cell.warehouseIndex {
					goto solutionFound
				}

				break

			}

			if !newStepFound {

				path.steps = append(path.steps[:len(path.steps)-1])

				if len(path.steps) <= 0 {
					goto solutionNotFound
				}

				currentStep = path.steps[len(path.steps)-1]

			}

		} else {

			warehouseIndex := currentStep.cell.warehouseIndex

			newStepFound := false

			for factoryIndex, _ := range table.factories {
				cell := table.cells[factoryIndex][warehouseIndex]

				if isCellVisited(cell, visitedCells) {
					continue
				}

				visitedCells = append(visitedCells, cell)

				if cell.supply <= 0 {
					continue
				}

				step := Step{}
				step.cell = cell
				step.scale = currentStep.scale * -1

				path.steps = append(path.steps, step)

				currentStep = step

				newStepFound = true

				if factoryIndex == initialStep.cell.factoryIndex {
					goto solutionFound
				}

				break

			}

			if !newStepFound {

				path.steps = append(path.steps[:len(path.steps)-1])

				if len(path.steps) <= 0 {
					goto solutionNotFound
				}

				currentStep = path.steps[len(path.steps)-1]

			}

		}

	}

solutionFound:
solutionNotFound:

	result <- path

	return

}

func optimizeSolution() {

optimize:

	emptyCells := make([]Cell, 0)

	for factoryIndex, _ := range table.factories {
		for warehouseIndex, _ := range table.warehouses {

			cell := table.cells[factoryIndex][warehouseIndex]

			if cell.supply <= 0 {
				emptyCells = append(emptyCells, cell)
			}

		}
	}

	result := make(chan Path, len(emptyCells))

	for _, cell := range emptyCells {
		go marginalCost(cell, result)

	}

	var lowestCostPath *Path
	var lowestCost int

	for _, _ = range emptyCells {

		path, _ := <-result
		cost := calculateCost(path)

		if lowestCostPath == nil || cost < lowestCost {
			lowestCostPath = &path
			lowestCost = cost
		}

	}

	close(result)

	if lowestCost < 0 {
		applySolutionPath(*lowestCostPath)

		table.cost = lowestCost
		goto optimize
	}

	return

}

func applySolutionPath(path Path) {

	leaving := 0

	for _, step := range path.steps {

		if step.scale > 0 {
			continue
		}

		if leaving == 0 || step.cell.supply < leaving {
			leaving = step.cell.supply
		}

	}

	for _, step := range path.steps {

		table.cells[step.cell.factoryIndex][step.cell.warehouseIndex].supply = step.cell.supply + (step.scale * leaving)

	}

}

func calculateCost(path Path) int {

	cost := 0

	for _, step := range path.steps {

		cost += step.cell.price * step.scale

	}

	return cost

}

func writeSolution(fileName string) {

	file, err := os.Create(fileName)

	defer file.Close()

	if err != nil {
		fmt.Printf("Could not open file %s, error: %v\n", fileName, err)
		return
	}

	file.WriteString("FINAL\t")

	for _, warehouse := range table.warehouses {
		file.WriteString(warehouse.name + "\t")
	}

	file.WriteString("SUPPLY\n")

	for factoryIndex, factory := range table.factories {

		file.WriteString(factory.name + "\t")

		for warehouseIndex, _ := range table.warehouses {

			supplyString := "-"

			supply := table.cells[factoryIndex][warehouseIndex].supply

			if supply > 0 {
				supplyString = strconv.Itoa(supply)
			}

			file.WriteString(supplyString + "\t")

		}

		file.WriteString(strconv.Itoa(factory.supply) + "\n")

	}

	file.WriteString("DEMAND\t")

	for _, warehouse := range table.warehouses {

		file.WriteString(strconv.Itoa(warehouse.demand) + "\t")

	}

	file.WriteString("\nTotal Cost is " + strconv.Itoa(table.cost))

	file.Sync()

}

var table Table

func main() {

	var definitionFileName string
	var initialSolutionFileName string

	if len(os.Args) >= 3 {

		definitionFileName = os.Args[1]
		initialSolutionFileName = os.Args[2]

	} else if len(os.Args) >= 2 {

		definitionFileName = os.Args[1]
		initialSolutionFileName = ""

	} else {
		panic(fmt.Sprintln("Usage: Assignment0Part2 <definition_file> <initial_solution_file>"))
	}

	parseDefinitionFile(definitionFileName)

	if initialSolutionFileName == "" {
		computeInitialSolution()
	} else {
		parseInitialSolutionFile(initialSolutionFileName)
	}

	optimizeSolution()
	writeSolution("output/solution.txt")

}

func isCellVisited(cell Cell, visitedCells []Cell) bool {

	for _, c := range visitedCells {
		if c == cell {
			return true
		}
	}

	return false

}
