package alpha9wordcounter

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

//WordStruct struct for holding the name of a word and the number of occurrencies
type WordStruct struct {
	word       string
	occurrency int
}

// FileWordMap struct for holding the name of a file and a map that has words and occurrencies of these words in the file
type FileWordMap struct {
	fileName string
	wordMap  WordMap
}

// WordMap map for holding the occurrencies of words
type WordMap map[string]int

var finalWords []WordStruct
var finalMap WordMap
var fileWordMaps []FileWordMap
var sortedWordsOfFiles []WordStruct

// Function that strips a string from given characters
func stripSpecialChars(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if strings.IndexRune(chr, r) < 0 {
			return r
		}
		return -1
	}, str)
}

// Fucntion that splits a string line into an array of word strings
func processLine(line string) []string {
	return strings.Split(stripSpecialChars(strings.ToLower(line), "§!\"#$%&'()*+,-./0123456789:;<=>?@{}[]\\^_`|'~¯´«»„“\r\t\n"), " ")

}

// Function that that puts a proccessed WordMap of each file in a channel and in a WordMap slice
func processFile(directoryPath string, c1 <-chan string, c2 chan WordMap, wg1, wg2 *sync.WaitGroup) {

	for filePath := range c1 {
		wg1.Done()
		wordMap := make(WordMap)

		file, err := os.Open(directoryPath + filePath)
		if err != nil {
			log.Fatal(err)
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			for _, s := range processLine(line) {
				wordMap[s]++
			}
		}
		c2 <- wordMap
		fileWordMap := FileWordMap{filePath, wordMap}
		fileWordMaps = append(fileWordMaps, fileWordMap)

		if e := scanner.Err(); e != nil {
			log.Fatal(e)
		}
		file.Close()
	}
	wg2.Done()
}

// CountWords function that builds the infrastracture so that Search and Common commands are ready for use
func CountWords() {

	fmt.Println("Building infrastracture...")
	//defer profile.Start().Stop()
	//defer profile.Start(profile.MemProfile).Stop()

	start := time.Now()
	var wg sync.WaitGroup
	var wg2 sync.WaitGroup
	var wg3 sync.WaitGroup
	var wg4 sync.WaitGroup

	concurrency := runtime.GOMAXPROCS(runtime.NumCPU())
	wordMapChan := make(chan WordMap, concurrency-1)
	mergedMapChan := make(chan WordMap, concurrency-1)
	path := "./books/"

	fileInfos, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	filePathChan := make(chan string, len(fileInfos))

	for i := 0; i < concurrency; i++ {
		wg2.Add(1)
		go processFile(path, filePathChan, wordMapChan, &wg, &wg2)
	}

	for _, fileInfo := range fileInfos {
		wg.Add(1)
		filePathChan <- fileInfo.Name()
	}
	close(filePathChan)

	for i := 0; i < concurrency; i++ {
		wg3.Add(1)
		go processWordMaps(i, wordMapChan, mergedMapChan, &wg3)
	}

	finalMergedMap := make(WordMap)
	wg4.Add(1)
	go processMergedWordMaps(mergedMapChan, finalMergedMap, &wg4)

	wg.Wait()
	wg2.Wait()
	close(wordMapChan)

	wg3.Wait()
	close(mergedMapChan)

	wg4.Wait()
	sortedWordsOfFiles = makeSortedArrayOfAllFiles(finalMergedMap)

	executionTime := time.Since(start)
	fmt.Println("Ready in: ", executionTime)
}

// Search function that finds the occurrencies of a given word in each file
func Search(word string) string {

	sort.Slice(fileWordMaps, func(i, j int) bool {
		return fileWordMaps[i].wordMap[word] > fileWordMaps[j].wordMap[word]
	})

	s := ""

	for _, w := range fileWordMaps {
		if wordCountPerFile := w.wordMap[word]; wordCountPerFile != 0 {
			s += fmt.Sprintf("[%s: %d] ", w.fileName, wordCountPerFile)
		}
	}
	return s
}

// Common funtion that finds the n most common words of all files
func Common(n int) string {

	s := ""

	for i := 0; i < n; i++ {
		s += fmt.Sprintf("%s %d\n", sortedWordsOfFiles[i].word, sortedWordsOfFiles[i].occurrency)
	}
	return s
}

// Function that converts a hashmap to a sorted slice of WordStructs
func makeSortedArrayOfAllFiles(finalMap WordMap) []WordStruct {

	delete(finalMap, "")

	// convert hashmap to slice of structs
	for k, v := range finalMap {
		finalWords = append(finalWords, WordStruct{k, v})
	}

	// Use golang quicksort to sort slice with occurrency order
	sort.Slice(finalWords, func(i, j int) bool {
		return finalWords[i].occurrency > finalWords[j].occurrency
	})
	return finalWords
}

// Function that merges into the final map the partial merged maps from a given channel of maps
func processMergedWordMaps(c1 chan WordMap, finalMergedMap WordMap, wg *sync.WaitGroup) {

	for mergedMap := range c1 {
		for k, v := range mergedMap {
			finalMergedMap[k] += v
		}
	}
	wg.Done()
}

// Function that merges maps from a given channel of maps
func processWordMaps(i int, c1, c2 chan WordMap, wg *sync.WaitGroup) {

	mergedMap := make(WordMap)
	for m1 := range c1 {
		for k, v := range m1 {
			mergedMap[k] += v
		}
	}
	c2 <- mergedMap
	wg.Done()
}
