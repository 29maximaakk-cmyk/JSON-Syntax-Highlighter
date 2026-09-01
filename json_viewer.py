# json_viewer.py
# Версия на Python с использованием json, argparse, цветной вывод

import sys
import json
import argparse
import os
import subprocess
from typing import Any, Optional

# ANSI-цвета
class Colors:
    RESET = '\033[0m'
    GRAY = '\033[90m'
    RED = '\033[91m'
    GREEN = '\033[92m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    MAGENTA = '\033[95m'
    CYAN = '\033[96m'
    WHITE = '\033[97m'
    BOLD = '\033[1m'

    @staticmethod
    def enable(enable: bool):
        if not enable:
            for attr in dir(Colors):
                if not attr.startswith('_') and isinstance(getattr(Colors, attr), str):
                    setattr(Colors, attr, '')

class JSONViewer:
    def __init__(self, data: Any, colors: bool = True, compact: bool = False,
                 minify: bool = False, filter_path: Optional[str] = None):
        self.data = data
        self.colors = colors
        self.compact = compact
        self.minify = minify
        self.filter_path = filter_path
        self.indent = None if compact or minify else 2
        self.separators = (',', ':') if compact else (', ', ': ')

        if self.filter_path:
            self.data = self._filter_data(self.data, self.filter_path)

    def _filter_data(self, obj: Any, path: str) -> Any:
        parts = path.split('.')
        current = obj
        for part in parts:
            if isinstance(current, dict) and part in current:
                current = current[part]
            else:
                raise ValueError(f"Путь '{path}' не найден")
        return current

    def _colorize_value(self, value: Any) -> str:
        if isinstance(value, str):
            return f"{Colors.GREEN}{json.dumps(value, ensure_ascii=False)}{Colors.RESET}"
        elif isinstance(value, bool):
            return f"{Colors.MAGENTA}{str(value).lower()}{Colors.RESET}"
        elif value is None:
            return f"{Colors.RED}null{Colors.RESET}"
        elif isinstance(value, (int, float)):
            return f"{Colors.YELLOW}{value}{Colors.RESET}"
        elif isinstance(value, (dict, list)):
            # Рекурсивно обработаем
            return self._colorize_complex(value)
        else:
            return str(value)

    def _colorize_complex(self, obj: Any) -> str:
        if isinstance(obj, dict):
            items = []
            for k, v in obj.items():
                key = f"{Colors.CYAN}{json.dumps(str(k), ensure_ascii=False)}{Colors.RESET}"
                val = self._colorize_value(v)
                items.append(f"{key}: {val}")
            if self.compact or self.minify:
                return f"{{{', '.join(items)}}}"
            else:
                indent = ' ' * 2
                inner = f',\n{indent}'.join(items)
                return f"{{\n{indent}{inner}\n}}"
        elif isinstance(obj, list):
            items = [self._colorize_value(item) for item in obj]
            if self.compact or self.minify:
                return f"[{', '.join(items)}]"
            else:
                indent = ' ' * 2
                inner = f',\n{indent}'.join(items)
                return f"[\n{indent}{inner}\n]"
        else:
            return self._colorize_value(obj)

    def render(self) -> str:
        if self.colors:
            return self._colorize_complex(self.data)
        else:
            # Обычный JSON
            if self.minify:
                return json.dumps(self.data, ensure_ascii=False, separators=(',', ':'))
            elif self.compact:
                return json.dumps(self.data, ensure_ascii=False, separators=(',', ':'))
            else:
                return json.dumps(self.data, indent=2, ensure_ascii=False)

def main():
    parser = argparse.ArgumentParser(description='JSON Syntax Highlighter (Python)')
    parser.add_argument('file', nargs='?', help='Входной JSON файл (если не указан, читается stdin)')
    parser.add_argument('-p', '--pretty', action='store_true', default=True, help='Pretty-print (по умолчанию)')
    parser.add_argument('-c', '--compact', action='store_true', help='Компактный вывод')
    parser.add_argument('-m', '--minify', action='store_true', help='Минифицированный вывод')
    parser.add_argument('-o', '--output', help='Сохранить результат в файл')
    parser.add_argument('--colors', action='store_true', default=True, help='Включить цвета')
    parser.add_argument('--no-colors', action='store_false', dest='colors', help='Отключить цвета')
    parser.add_argument('--theme', choices=['light', 'dark'], help='Цветовая тема')
    parser.add_argument('--filter', help='Фильтр по пути (например, "user.address.city")')
    parser.add_argument('--page', action='store_true', help='Использовать пейджер (less)')
    args = parser.parse_args()

    # Читаем данные
    try:
        if args.file:
            with open(args.file, 'r', encoding='utf-8') as f:
                data = json.load(f)
        else:
            data = json.load(sys.stdin)
    except json.JSONDecodeError as e:
        print(f"Ошибка парсинга JSON: {e}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Ошибка чтения: {e}", file=sys.stderr)
        sys.exit(1)

    # Если тема указана, можно менять цвета (в демо оставим стандартные)
    if args.theme == 'light':
        # Можно изменить цвета для светлой темы, но оставим как есть
        pass

    # Создаём viewer
    viewer = JSONViewer(
        data=data,
        colors=args.colors,
        compact=args.compact or args.minify,
        minify=args.minify,
        filter_path=args.filter
    )
    output = viewer.render()

    # Сохраняем или выводим
    if args.output:
        with open(args.output, 'w', encoding='utf-8') as f:
            f.write(output)
        print(f"Результат сохранён в {args.output}")
    else:
        if args.page:
            # Используем less с опцией -R для цветов
            pager = subprocess.Popen(['less', '-R'], stdin=subprocess.PIPE)
            pager.communicate(output.encode('utf-8'))
        else:
            print(output)

if __name__ == '__main__':
    main()
