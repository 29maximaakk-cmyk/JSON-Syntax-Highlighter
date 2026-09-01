// json_viewer.ts
// Версия на TypeScript с классами, интерфейсами, декораторами

import { Command } from 'commander';
import * as fs from 'fs';
import chalk from 'chalk';
import { spawn } from 'child_process';

// Цветовая схема
const colors = {
    key: chalk.cyan,
    string: chalk.green,
    number: chalk.yellow,
    boolean: chalk.magenta,
    null: chalk.red,
    brace: chalk.gray,
    bracket: chalk.gray,
    comma: chalk.gray,
    colon: chalk.gray,
};

interface ViewerOptions {
    compact?: boolean;
    minify?: boolean;
    colors?: boolean;
    filter?: string;
}

class JSONViewer {
    private data: any;
    private compact: boolean;
    private minify: boolean;
    private colors: boolean;
    private filterPath: string | null;
    private indent: number;

    constructor(data: any, options: ViewerOptions = {}) {
        this.data = data;
        this.compact = options.compact || false;
        this.minify = options.minify || false;
        this.colors = options.colors !== false;
        this.filterPath = options.filter || null;
        this.indent = this.compact || this.minify ? 0 : 2;

        if (this.filterPath) {
            this.data = this._filterData(this.data, this.filterPath);
        }
    }

    private _filterData(obj: any, path: string): any {
        const parts = path.split('.');
        let current = obj;
        for (const part of parts) {
            if (current && typeof current === 'object' && current[part] !== undefined) {
                current = current[part];
            } else {
                throw new Error(`Путь '${path}' не найден`);
            }
        }
        return current;
    }

    private _colorizeValue(value: any): string {
        if (typeof value === 'string') {
            return this.colors ? colors.string(JSON.stringify(value)) : JSON.stringify(value);
        } else if (typeof value === 'boolean') {
            return this.colors ? colors.boolean(String(value)) : String(value);
        } else if (value === null) {
            return this.colors ? colors.null('null') : 'null';
        } else if (typeof value === 'number') {
            return this.colors ? colors.number(String(value)) : String(value);
        } else if (Array.isArray(value) || (typeof value === 'object' && value !== null)) {
            return this._colorizeComplex(value);
        } else {
            return String(value);
        }
    }

    private _colorizeComplex(obj: any): string {
        if (Array.isArray(obj)) {
            const items = obj.map(item => this._colorizeValue(item));
            if (this.compact || this.minify) {
                return `[${items.join(', ')}]`;
            } else {
                const indent = ' '.repeat(this.indent);
                const inner = items.map(item => `${indent}${item}`).join(',\n');
                return `[\n${inner}\n]`;
            }
        } else if (obj !== null && typeof obj === 'object') {
            const entries = Object.entries(obj);
            const items = entries.map(([k, v]) => {
                const key = this.colors ? colors.key(`"${k}"`) : `"${k}"`;
                const val = this._colorizeValue(v);
                return `${key}: ${val}`;
            });
            if (this.compact || this.minify) {
                return `{${items.join(', ')}}`;
            } else {
                const indent = ' '.repeat(this.indent);
                const inner = items.map(item => `${indent}${item}`).join(',\n');
                return `{\n${inner}\n}`;
            }
        } else {
            return String(obj);
        }
    }

    render(): string {
        return this._colorizeComplex(this.data);
    }
}

const program = new Command();
program
    .name('json_viewer')
    .description('JSON Syntax Highlighter (TypeScript)')
    .argument('[file]', 'JSON файл')
    .option('-p, --pretty', 'Pretty-print (по умолчанию)', true)
    .option('-c, --compact', 'Компактный вывод')
    .option('-m, --minify', 'Минифицированный вывод')
    .option('-o, --output <file>', 'Сохранить результат в файл')
    .option('--colors', 'Включить цвета', true)
    .option('--no-colors', 'Отключить цвета')
    .option('--theme <light|dark>', 'Цветовая тема')
    .option('--filter <path>', 'Фильтр по пути')
    .option('--page', 'Использовать пейджер (less)')
    .action(async (file, options) => {
        try {
            let input: string;
            if (file) {
                input = fs.readFileSync(file, 'utf-8');
            } else {
                input = fs.readFileSync(0, 'utf-8');
            }
            const data = JSON.parse(input);
            const viewer = new JSONViewer(data, {
                compact: options.compact || options.minify,
                minify: options.minify,
                colors: options.colors,
                filter: options.filter,
            });
            const output = viewer.render();

            if (options.output) {
                fs.writeFileSync(options.output, output, 'utf-8');
                console.log(`Результат сохранён в ${options.output}`);
            } else {
                if (options.page) {
                    const less = spawn('less', ['-R'], { stdio: ['pipe', 'inherit', 'inherit'] });
                    less.stdin.write(output);
                    less.stdin.end();
                } else {
                    console.log(output);
                }
            }
        } catch (err: any) {
            console.error(`Ошибка: ${err.message}`);
            process.exit(1);
        }
    });

program.parse(process.argv);
