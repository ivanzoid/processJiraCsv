package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		log.Fatalf("Usage: %s <inputFile>\n", os.Args[0])
		return
	}

	filePath := args[0]

	file, err := os.Open(filePath)
	if err != nil {
		return
	}

	csvReader := csv.NewReader(file)
	csvReader.Comma = ';'

	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatalln("Can't read file:", err)
	}

	if len(records) == 0 {
		return
	}

	firstLine := records[0]

	summaryIndex := -1
	issudIdIndex := -1
	parentIdIndex := -1
	originalEstimateIndex := -1
	for index, item := range firstLine {
		if strings.EqualFold(item, "Summary") {
			summaryIndex = index
		} else if strings.EqualFold(item, "Issue ID") {
			issudIdIndex = index
		} else if strings.EqualFold(item, "Parent ID") {
			parentIdIndex = index
		} else if strings.EqualFold(item, "Original Estimate") {
			originalEstimateIndex = index
		}
	}

	if summaryIndex == -1 {
		fmt.Fprintf(os.Stderr, "Missing Summary field in first line")
		return
	}
	if issudIdIndex == -1 {
		fmt.Fprintf(os.Stderr, "Missing Issue ID field in first line")
		return
	}
	if parentIdIndex == -1 {
		fmt.Fprintf(os.Stderr, "Missing Parent ID field in first line")
		return
	}

	currentParentIssueSummary := ""
	maxRecordLength := 0
	for index := 1; index < len(records); index++ {
		record := &records[index]

		recordLength := len(*record)

		if recordLength < summaryIndex || recordLength < issudIdIndex || recordLength < parentIdIndex {
			continue
		}

		if recordLength > maxRecordLength {
			maxRecordLength = recordLength
		}

		summaryString := (*record)[summaryIndex]
		issudIdString := (*record)[issudIdIndex]
		parentIdString := (*record)[parentIdIndex]

		if len(summaryString) == 0 {
			fmt.Fprintf(os.Stderr, "Warning: empty Summary at line %d\n", index)
			continue
		}

		if len(issudIdString) != 0 {
			currentParentIssueSummary = summaryString
		}

		if summaryString[0] == '\t' {
			summaryString = summaryString[1:]
		}

		if len(parentIdString) != 0 {
			summaryString = fmt.Sprintf("%s | %s", currentParentIssueSummary, summaryString)
		}

		summaryString = fmt.Sprintf("iOS | %s", summaryString)

		(*record)[summaryIndex] = summaryString

		if recordLength >= originalEstimateIndex {
			originalEstimateString := (*record)[originalEstimateIndex]
			if len(originalEstimateString) != 0 {
				estimate, err := strconv.ParseInt(originalEstimateString, 10, 64)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Can't parse Original Estimate at line %d (%s)\n", index, originalEstimateString)
				} else {
					estimate *= 3600000 // Jira expects milliseconds, csv contains hours
					originalEstimateString = fmt.Sprintf("%d", estimate)
				}
				(*record)[originalEstimateIndex] = originalEstimateString
			}
		}
	}

	// for _, record := range records {
	// 	fmt.Println(record)
	// }

	csvWriter := csv.NewWriter(os.Stdout)
	csvWriter.WriteAll(records)

	if err := csvWriter.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}
}
