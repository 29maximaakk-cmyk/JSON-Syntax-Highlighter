# json_viewer.rb
# Версия на Ruby с использованием json, цветного вывода, OptionParser

require 'json'
require 'optparse'

# ANSI-цвета
COLORS = {
  reset: "\033[0m",
  cyan: "\033[96m",
  green: "\033[92m",
  yellow: "\033[93m",
  magenta: "\033[95m",
  red: "\033[91m",
  gray: "\033[90m"
}

class JSONViewer
  def initialize(data, compact: false, minify: false, colors: true, filter: nil)
    @data = data
    @compact = compact || minify
    @minify = minify
    @colors = colors
    @filter = filter
    @indent = @compact ? "" : "  "
  end

  def filter_data(obj, path)
    parts = path.split('.')
    current = obj
    parts.each do |part|
      if current.is_a?(Hash) && current.key?(part)
        current = current[part]
      else
        raise "Путь '#{path}' не найден"
      end
    end
    current
  end

  def colorize_value(value)
    case value
    when String
      s = value.to_json
      @colors ? "#{COLORS[:green]}#{s}#{COLORS[:reset]}" : s
    when TrueClass, FalseClass
      s = value.to_s
      @colors ? "#{COLORS[:magenta]}#{s}#{COLORS[:reset]}" : s
    when NilClass
      @colors ? "#{COLORS[:red]}null#{COLORS[:reset]}" : 'null'
    when Numeric
      @colors ? "#{COLORS[:yellow]}#{value}#{COLORS[:reset]}" : value.to_s
    when Array, Hash
      colorize_complex(value)
    else
      value.to_s
    end
  end

  def colorize_complex(obj)
    if obj.is_a?(Array)
      items = obj.map { |item| colorize_value(item) }
      if @compact || @minify
        "[#{items.join(', ')}]"
      else
        inner = items.map { |item| "#{@indent}#{item}" }.join(",\n")
        "[\n#{inner}\n]"
      end
    elsif obj.is_a?(Hash)
      entries = obj.map do |k, v|
        key = k.to_s.to_json
        key = "#{COLORS[:cyan]}#{key}#{COLORS[:reset]}" if @colors
        val = colorize_value(v)
        "#{key}: #{val}"
      end
      if @compact || @minify
        "{#{entries.join(', ')}}"
      else
        inner = entries.map { |entry| "#{@indent}#{entry}" }.join(",\n")
        "{\n#{inner}\n}"
      end
    else
      colorize_value(obj)
    end
  end

  def render
    data = @data
    if @filter
      data = filter_data(data, @filter)
    end
    colorize_complex(data)
  end
end

options = {}
OptionParser.new do |opts|
  opts.banner = "Использование: ruby json_viewer.rb [опции] [файл]"
  opts.on("-p", "--pretty", "Pretty-print (по умолчанию)") { options[:pretty] = true }
  opts.on("-c", "--compact", "Компактный вывод") { options[:compact] = true }
  opts.on("-m", "--minify", "Минифицированный вывод") { options[:minify] = true }
  opts.on("-o", "--output FILE", "Сохранить результат") { |v| options[:output] = v }
  opts.on("--no-colors", "Отключить цвета") { options[:no_colors] = true }
  opts.on("--filter PATH", "Фильтр по пути") { |v| options[:filter] = v }
  opts.on("--page", "Использовать пейджер (less)") { options[:page] = true }
  opts.on("-h", "--help", "Справка") { puts opts; exit }
end.parse!

file = ARGV[0]
colors = !options[:no_colors]

begin
  json = if file
           File.read(file)
         else
           $stdin.read
         end
  data = JSON.parse(json)
rescue JSON::ParserError => e
  $stderr.puts "Ошибка парсинга JSON: #{e.message}"
  exit 1
rescue => e
  $stderr.puts "Ошибка: #{e.message}"
  exit 1
end

viewer = JSONViewer.new(data,
  compact: options[:compact],
  minify: options[:minify],
  colors: colors,
  filter: options[:filter]
)
output = viewer.render

if options[:output]
  File.write(options[:output], output)
  puts "Результат сохранён в #{options[:output]}"
else
  if options[:page]
    IO.popen("less -R", "w") { |io| io.puts output }
  else
    puts output
  end
end
