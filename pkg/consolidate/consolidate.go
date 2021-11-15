package consolidate

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
)

// Consolidate will open the files, read the contents, and try to create a
// single tfvar file.
func Consolidate(out string, files []string) error {

	if out == "" {
		o, _ := ioutil.TempFile("", "t")
		out = o.Name()
		fmt.Println("Saving to:", out)
	}

	var tfvars []byte
	for _, f := range files {

		b, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}

		if strings.HasSuffix(f, ".json") {
			// TODO convert json to hcl syntax for tfvar files
			return fmt.Errorf("json tfvars are not yet supported")
		}

		tfvars = append(tfvars, b...)

	}

	var c bytes.Buffer
	var currentKey string
	var currentValue string
	keyIndexer := make(map[string]string)
	var openBrackets int
	for _, line := range strings.Split(string(tfvars), "\n") {
		lineArr := strings.Split(line, "=")
		// ignore blank lines
		if strings.TrimSpace(lineArr[0]) == "" {
			continue
		}
		// ignore comments
		if strings.HasPrefix(strings.TrimSpace(line), "//") || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		if openBrackets > 0 {
			currentValue += "\n" + strings.ReplaceAll(line, "\t", "  ")
			// Check for more open brackets and close brackets
			trimmedLine := strings.TrimSpace(line)
			lastCharIdx := len(trimmedLine) - 1
			lastChar := string(trimmedLine[lastCharIdx])
			lastTwoChar := ""
			if lastCharIdx > 0 {
				lastTwoChar = string(trimmedLine[lastCharIdx-1:])
			}

			if lastChar == "{" || lastChar == "[" {
				openBrackets++
			} else if lastChar == "}" || lastChar == "]" || lastTwoChar == "}," || lastTwoChar == "]," {
				openBrackets--
			}
			if openBrackets == 0 {
				keyIndexer[currentKey] = currentValue
			}
			continue
		}
		currentKey = strings.TrimSpace(lineArr[0])

		if len(lineArr) > 1 {
			lastLineArrIdx := len(lineArr) - 1
			trimmedLine := lineArr[lastLineArrIdx]
			lastCharIdx := len(trimmedLine) - 1
			lastChar := string(trimmedLine[lastCharIdx])
			if lastChar == "{" || lastChar == "[" {
				openBrackets++
			}
		} else {
			return fmt.Errorf("error in parsing tfvars string: %s", line)
		}

		currentValue = line
		if openBrackets > 0 {
			continue
		}
		keyIndexer[currentKey] = currentValue
	}

	keys := make([]string, 0, len(keyIndexer))
	for k := range keyIndexer {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&c, "%s\n\n", keyIndexer[k])
	}

	if out == "" {
		fmt.Printf("%s\n", c.String())
	} else {
		err := ioutil.WriteFile(out, c.Bytes(), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
