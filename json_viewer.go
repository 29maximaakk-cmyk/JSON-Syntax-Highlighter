// json_viewer.go
// Версия на Go с использованием encoding/json, флагов, ANSI-цветов

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// ANSI-цвета
const (
	reset  = "\033[0m"
	gray   = "\033[90m"
	red    = "\033[91m"
	green  = "\033[92m"
	yellow = "\033[93m"
	blue   = "\033[94m"
	magenta= "\033[95m"
	cyan   = "\033[96m"
)

var colorMap = map[string]string{
	"key":    cyan,
	"string": green,
	"number": yellow,
	"bool":   magenta,
	"null":   red,
	"brace":  gray,
}

type JSONViewer struct {
	data    interface{}
	compact bool
	minify  bool
	colors  bool
	filter  string
	indent  string
}

func NewJSONViewer(data interface{}, compact, minify, colors bool, filter string) *JSONViewer {
	indent := "  "
	if compact || minify {
		indent = ""
	}
	return &JSONViewer{
		data:    data,
		compact: compact || minify,
		minify:  minify,
		colors:  colors,
		filter:  filter,
		indent:  indent,
	}
}

func (v *JSONViewer) filterData(obj interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := obj
	for _, part := range parts {
		switch val := current.(type) {
		case map[string]interface{}:
			if v, ok := val[part]; ok {
				current = v
			} else {
				return nil, fmt.Errorf("путь '%s' не найден", path)
			}
		default:
			return nil, fmt.Errorf("путь '%s' не применим к этому типу данных", path)
		}
	}
	return current, nil
}

func (v *JSONViewer) colorizeValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		if v.colors {
			return colorMap["string"] + fmt.Sprintf(`"%s"`, v) + reset
		}
		return fmt.Sprintf(`"%s"`, v)
	case bool:
		if v.colors {
			return colorMap["bool"] + fmt.Sprintf("%t", v) + reset
		}
		return fmt.Sprintf("%t", v)
	case nil:
		if v.colors {
			return colorMap["null"] + "null" + reset
		}
		return "null"
	case float64, int, int64:
		if v.colors {
			return colorMap["number"] + fmt.Sprintf("%v", v) + reset
		}
		return fmt.Sprintf("%v", v)
	case map[string]interface{}, []interface{}:
		return v.colorizeComplex(val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (v *JSONViewer) colorizeComplex(obj interface{}) string {
	switch val := obj.(type) {
	case []interface{}:
		items := make([]string, len(val))
		for i, item := range val {
			items[i] = v.colorizeValue(item)
		}
		if v.compact || v.minify {
			return "[" + strings.Join(items, ", ") + "]"
		}
		inner := strings.Join(items, ",\n"+v.indent)
		return "[\n" + v.indent + inner + "\n]"
	case map[string]interface{}:
		entries := make([]string, 0, len(val))
		for k, vv := range val {
			key := fmt.Sprintf(`"%s"`, k)
			if v.colors {
				key = colorMap["key"] + key + reset
			}
			entries = append(entries, key+": "+v.colorizeValue(vv))
		}
		if v.compact || v.minify {
			return "{" + strings.Join(entries, ", ") + "}"
		}
		inner := strings.Join(entries, ",\n"+v.indent)
		return "{\n" + v.indent + inner + "\n}"
	default:
		return v.colorizeValue(obj)
	}
}

func (v *JSONViewer) Render() string {
	if v.filter != "" {
		filtered, err := v.filterData(v.data, v.filter)
		if err != nil {
			return "Ошибка фильтрации: " + err.Error()
		}
		v.data = filtered
	}
	return v.colorizeComplex(v.data)
}

func main() {
	var (
		pretty     bool
		compact    bool
		minify     bool
		outputFile string
		noColors   bool
		theme      string
		filter     string
		page       bool
	)
	flag.BoolVar(&pretty, "p", true, "Pretty-print")
	flag.BoolVar(&compact, "c", false, "Компактный вывод")
	flag.BoolVar(&minify, "m", false, "Минифицированный вывод")
	flag.StringVar(&outputFile, "o", "", "Сохранить результат в файл")
	flag.BoolVar(&noColors, "no-colors", false, "Отключить цвета")
	flag.StringVar(&theme, "theme", "", "Цветовая тема (light/dark)")
	flag.StringVar(&filter, "filter", "", "Фильтр по пути")
	flag.BoolVar(&page, "page", false, "Использовать пейджер (less)")
	flag.Usage = func() {
		fmt.Println("Использование: go run json_viewer.go [опции] [файл]")
		fmt.Println("  -p           Pretty-print (по умолчанию)")
		fmt.Println("  -c           Компактный вывод")
		fmt.Println("  -m           Минифицированный вывод")
		fmt.Println("  -o <file>    Сохранить результат")
		fmt.Println("  --no-colors  Отключить цвета")
		fmt.Println("  --theme      Тема (light/dark)")
		fmt.Println("  --filter     Фильтр по пути")
		fmt.Println("  --page       Использовать пейджер")
	}
	flag.Parse()

	// Читаем данные
	var input []byte
	if flag.NArg() > 0 {
		file := flag.Arg(0)
		var err error
		input, err = os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка чтения файла: %v\n", err)
			os.Exit(1)
		}
	} else {
		var err error
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка чтения stdin: %v\n", err)
			os.Exit(1)
		}
	}

	var data interface{}
	err := json.Unmarshal(input, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка парсинга JSON: %v\n", err)
		os.Exit(1)
	}

	colors := !noColors
	viewer := NewJSONViewer(data, compact, minify, colors, filter)
	output := viewer.Render()

	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(output), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка записи: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Результат сохранён в %s\n", outputFile)
	} else {
		if page {
			cmd := exec.Command("less", "-R")
			cmd.Stdin = strings.NewReader(output)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		} else {
			fmt.Println(output)
		}
	}
}
