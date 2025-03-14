package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mfonda/simhash"
)

func main() {
	start := time.Now()

	stepStart := time.Now()
	filePaths := getFilePaths()
	fmt.Printf("getFilePaths execution time: %.4fs\n", time.Since(stepStart).Seconds())

	stepStart = time.Now()
	htmlFiles := readFiles(filePaths)
	fmt.Printf("readFiles execution time: %.4fs\n", time.Since(stepStart).Seconds())

	stepStart = time.Now()
	fileFeatures := processHTMLFiles(htmlFiles)
	fmt.Printf("processHTMLFiles execution time: %.4fs\n", time.Since(stepStart).Seconds())

	stepStart = time.Now()
	simhashes := computeSimhashes(fileFeatures)
	fmt.Printf("computeSimhashes execution time: %.4fs\n", time.Since(stepStart).Seconds())

	fmt.Printf("Total execution time: %.4fs\n\n", time.Since(start).Seconds())

	for i, hash := range simhashes {
		fmt.Printf("File %d: %x\n", i, hash)
	}
}

func getFilePaths() []string {
	files, err := filepath.Glob("files/*.html")
	if err != nil {
		log.Fatal(err)
	}

	sort.Strings(files)

	return files
}

func readFiles(files []string) []string {
	var htmlFiles []string
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			content, err := os.ReadFile(file)
			if err != nil {
				log.Fatal(err)
			}
			mu.Lock()
			htmlFiles = append(htmlFiles, string(content))
			mu.Unlock()
		}(file)
	}

	wg.Wait()
	return htmlFiles
}

func processHTMLFiles(htmlFiles []string) []map[string]int {
	var wg sync.WaitGroup
	var mu sync.Mutex
	fileFeatures := make([]map[string]int, len(htmlFiles))

	for i, html := range htmlFiles {
		wg.Add(1)
		go func(i int, html string) {
			defer wg.Done()
			// Remove script and style elements
			re := regexp.MustCompile(`(?s)<(script|style).*?>.*?</\\1>`)
			html = re.ReplaceAllString(html, "")

			// Get lowercase text
			html = strings.ToLower(html)

			// Remove all punctuation
			re = regexp.MustCompile(`[^\w\s]`)
			html = re.ReplaceAllString(html, "")

			// Break into lines and remove leading and trailing space on each
			lines := strings.Split(html, "\n")
			for i, line := range lines {
				lines[i] = strings.TrimSpace(line)
			}

			// Break multi-headlines into a line each
			var processedLines []string
			for _, line := range lines {
				processedLines = append(processedLines, strings.Split(line, ".")...)
			}

			// Drop blank lines and update features map
			fileFeature := make(map[string]int)
			for _, line := range processedLines {
				if line != "" {
					fileFeature[line]++
				}
			}

			mu.Lock()
			fileFeatures[i] = fileFeature
			mu.Unlock()
		}(i, html)
	}

	wg.Wait()
	return fileFeatures
}

func computeSimhashes(fileFeatures []map[string]int) []uint64 {
	var wg sync.WaitGroup
	simhashes := make([]uint64, len(fileFeatures))

	for i, features := range fileFeatures {
		wg.Add(1)
		go func(i int, features map[string]int) {
			defer wg.Done()
			var keys []string
			for k := range features {
				keys = append(keys, k)
			}
			combinedKeys := strings.Join(keys, " ")
			hash := simhash.Simhash(simhash.NewWordFeatureSet([]byte(combinedKeys)))
			simhashes[i] = hash
		}(i, features)
	}
	wg.Wait()
	return simhashes
}
