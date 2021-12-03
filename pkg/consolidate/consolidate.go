package consolidate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/tmccombs/hcl2json/convert"
)

func jsonToHcl(j interface{}, indentCount int) (string, error) {

	indent := strings.Repeat(" ", indentCount)
	output := ""

	switch jj := j.(type) {
	case map[string]interface{}:
		for k, v := range jj {
			switch c := v.(type) {
			case string:
				// format:
				// k = "v"
				s, err := json.Marshal(c)
				if err != nil {
					return "", err
				}
				output += fmt.Sprintf("%s%s = %s\n", indent, k, string(s))
			case float64:
				// format:
				// k = v
				s, err := json.Marshal(v)
				if err != nil {
					return "", err
				}
				output += fmt.Sprintf("%s%s = %s\n", indent, k, string(s))
			case bool:
				// format:
				// k = v
				output += fmt.Sprintf("%s%s = %t\n", indent, k, c)
			case []interface{}:
				// format:
				// k = [v]
				_output, err := jsonToHcl(c, indentCount+2)
				if err != nil {
					return "", err
				}
				output += fmt.Sprintf("%s%s = [\n%s%s]\n", indent, k, _output, indent)
			case map[string]interface{}:
				// format:
				// k = {v}
				_output, err := jsonToHcl(c, indentCount+2)
				if err != nil {
					return "", err
				}
				output += fmt.Sprintf("%s%s = {\n%s%s}\n", indent, k, _output, indent)
			default:
				return "", fmt.Errorf("not sure what type item %q is, but I think it might be %T", k, c)
			}
		}
	case []interface{}:
		l := len(jj)
		for index, i := range jj {
			comma := ","
			if index >= l-1 {
				comma = ""
			}
			switch c := i.(type) {
			case string:
				// format:
				// "v",
				s, err := json.Marshal(c)
				if err != nil {
					return "", err
				}
				output += fmt.Sprintf("%s%s%s\n", indent, string(s), comma)
			case float64:
				// format:
				// v,
				s, err := json.Marshal(c)
				if err != nil {
					return "", err
				}
				output += fmt.Sprintf("%s%s%s\n", indent, string(s), comma)
			case bool:
				// format:
				// v,
				output += fmt.Sprintf("%s%t%s\n", indent, c, comma)
			case []interface{}:
				// format:
				// [v],
				_output, err := jsonToHcl(c, indentCount+2)
				if err != nil {
					return "", err
				}
				output += fmt.Sprintf("%s[\n%s%s]%s\n", indent, _output, indent, comma)
			case map[string]interface{}:
				// format:
				// {v},
				_output, err := jsonToHcl(c, indentCount+2)
				if err != nil {
					return "", err
				}
				output += fmt.Sprintf("%s{\n%s%s}%s\n", indent, _output, indent, comma)
			default:
				return "", fmt.Errorf("not sure what type item %q is, but I think it might be %T", c, c)
			}
		}
	}

	return output, nil
}

// Consolidate will open the files, read the contents, and try to create a
// single tfvar file.
func Consolidate(out string, files []string, useEnvs bool, backend string) error {

	f, err := ioutil.TempFile("", "e")
	if err == nil {
		files = append(files, f.Name())
		defer os.Remove(f.Name())
	} else if err != nil && useEnvs {
		return err
	}
	if useEnvs {
		envVars := ""
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, "TF_VAR_") {
				k := strings.TrimPrefix(strings.Split(env, "=")[0], "TF_VAR_")
				v := strings.Join(strings.Split(env, "=")[1:], "=")
				if v == "" {
					continue
				}
				if string(v[0]) != "{" && string(v[0]) != "[" {
					v = fmt.Sprintf("\"%s\"", v)
				}
				envVars += fmt.Sprintf("\n%s = %s", k, v)
			}
		}
		_, err := f.Write([]byte(envVars))
		if err != nil {
			return err
		}
	}

	var tfvars []byte
	for _, f := range files {

		b, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}

		if !strings.HasSuffix(f, ".json") {
			b, err = convert.Bytes(b, "", convert.Options{})
			if err != nil {
				return err
			}
		}

		j := make(map[string]interface{})
		err = json.Unmarshal(b, &j)
		if err != nil {
			return err
		}

		output, err := jsonToHcl(j, 0)
		if err != nil {
			return err
		}
		b = []byte(output)

		tfvars = append(tfvars, b...)

	}

	if backend != "" {

		type BackendResource struct {
			Resource map[string][]interface{} `json:"backend"`
		}

		type TerraformBackend struct {
			Backends []BackendResource `json:"terraform"`
		}

		b, err := ioutil.ReadFile(backend)
		if err != nil {
			return err
		}

		if !strings.HasSuffix(backend, ".json") {
			b, err = convert.Bytes(b, "", convert.Options{})
			if err != nil {
				return err
			}
		}

		// For simplicity, this project makes an assumption the backend is a
		// fully defined terraform resource. By removing the backend resource
		// proclamation, only the vars will be extracted.
		//
		// Furthermore, since (as far as I know or as far as I have bothered to
		// test) terraform can be used with a single backend at a time, this
		// project will just read the first backend resource defined.
		j := TerraformBackend{}
		err = json.Unmarshal(b, &j)
		if err != nil {
			return err
		}
		if len(j.Backends) > 0 {
			for _, values := range j.Backends[0].Resource {
				if len(values) > 0 {
					output, err := jsonToHcl(values[0], 0)
					if err != nil {
						return err
					}
					b = []byte(output)
				}
			}
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
		o, _ := ioutil.TempFile("", "t")
		out = o.Name()
		fmt.Fprintf(os.Stderr, "Saving to: ")
		fmt.Fprintln(os.Stdout, out)
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
