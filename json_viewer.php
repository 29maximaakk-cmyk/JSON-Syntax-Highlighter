<?php
// json_viewer.php
// Версия на PHP с использованием json_decode, ANSI-цветов

// ANSI-цвета
define('RESET', "\033[0m");
define('CYAN', "\033[96m");
define('GREEN', "\033[92m");
define('YELLOW', "\033[93m");
define('MAGENTA', "\033[95m");
define('RED', "\033[91m");
define('GRAY', "\033[90m");

class JSONViewer {
    private $data;
    private $compact;
    private $minify;
    private $colors;
    private $filter;
    private $indent;

    public function __construct($data, $compact = false, $minify = false, $colors = true, $filter = null) {
        $this->data = $data;
        $this->compact = $compact || $minify;
        $this->minify = $minify;
        $this->colors = $colors;
        $this->filter = $filter;
        $this->indent = $compact || $minify ? "" : "  ";
    }

    private function filterData($obj, $path) {
        $parts = explode('.', $path);
        $current = $obj;
        foreach ($parts as $part) {
            if (is_array($current) && array_key_exists($part, $current)) {
                $current = $current[$part];
            } else {
                throw new Exception("Путь '$path' не найден");
            }
        }
        return $current;
    }

    private function colorizeValue($value) {
        if (is_string($value)) {
            $s = json_encode($value, JSON_UNESCAPED_UNICODE);
            return $this->colors ? GREEN . $s . RESET : $s;
        } elseif (is_bool($value)) {
            $s = $value ? 'true' : 'false';
            return $this->colors ? MAGENTA . $s . RESET : $s;
        } elseif ($value === null) {
            return $this->colors ? RED . 'null' . RESET : 'null';
        } elseif (is_numeric($value)) {
            return $this->colors ? YELLOW . $value . RESET : (string)$value;
        } elseif (is_array($value)) {
            return $this->colorizeComplex($value);
        } else {
            return (string)$value;
        }
    }

    private function colorizeComplex($obj) {
        if (array_values($obj) === $obj) {
            // список
            $items = [];
            foreach ($obj as $item) {
                $items[] = $this->colorizeValue($item);
            }
            if ($this->compact || $this->minify) {
                return '[' . implode(', ', $items) . ']';
            }
            $inner = implode(",\n" . $this->indent, $items);
            return "[\n" . $this->indent . $inner . "\n]";
        } else {
            // объект
            $entries = [];
            foreach ($obj as $k => $v) {
                $key = json_encode((string)$k, JSON_UNESCAPED_UNICODE);
                if ($this->colors) $key = CYAN . $key . RESET;
                $val = $this->colorizeValue($v);
                $entries[] = $key . ': ' . $val;
            }
            if ($this->compact || $this->minify) {
                return '{' . implode(', ', $entries) . '}';
            }
            $inner = implode(",\n" . $this->indent, $entries);
            return "{\n" . $this->indent . $inner . "\n}";
        }
    }

    public function render() {
        $data = $this->data;
        if ($this->filter) {
            $data = $this->filterData($data, $this->filter);
        }
        return $this->colorizeComplex($data);
    }
}

// Парсинг аргументов
$opts = getopt('pcmo:', ['compact', 'minify', 'colors', 'no-colors', 'filter:', 'page', 'output:']);
$file = null;
$compact = isset($opts['c']) || isset($opts['compact']);
$minify = isset($opts['m']) || isset($opts['minify']);
$colors = !isset($opts['no-colors']) && !isset($opts['colors']) ? true : (isset($opts['no-colors']) ? false : true);
$filter = $opts['filter'] ?? null;
$output = $opts['o'] ?? $opts['output'] ?? null;
$page = isset($opts['page']);

// Определяем файл
$args = array_values(array_filter($argv, function($a) { return !str_starts_with($a, '-'); }));
$file = $args[1] ?? null;

try {
    if ($file) {
        $json = file_get_contents($file);
    } else {
        $json = file_get_contents('php://stdin');
    }
    if ($json === false) throw new Exception("Не удалось прочитать данные");
    $data = json_decode($json, true);
    if (json_last_error() !== JSON_ERROR_NONE) {
        throw new Exception("Ошибка парсинга JSON: " . json_last_error_msg());
    }
    $viewer = new JSONViewer($data, $compact, $minify, $colors, $filter);
    $result = $viewer->render();

    if ($output) {
        file_put_contents($output, $result);
        echo "Результат сохранён в $output\n";
    } else {
        if ($page) {
            $descriptors = [0 => ['pipe', 'r'], 1 => STDOUT, 2 => STDERR];
            $proc = proc_open('less -R', $descriptors, $pipes);
            if (is_resource($proc)) {
                fwrite($pipes[0], $result);
                fclose($pipes[0]);
                proc_close($proc);
            }
        } else {
            echo $result . "\n";
        }
    }
} catch (Exception $e) {
    fwrite(STDERR, "Ошибка: " . $e->getMessage() . "\n");
    exit(1);
}
