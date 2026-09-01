// JsonViewer.cs
// Версия на C# с использованием System.Text.Json, консольных цветов

using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Text.Json;

namespace JsonViewer
{
    class Program
    {
        private const string Reset = "\u001b[0m";
        private const string Cyan = "\u001b[96m";
        private const string Green = "\u001b[92m";
        private const string Yellow = "\u001b[93m";
        private const string Magenta = "\u001b[95m";
        private const string Red = "\u001b[91m";
        private const string Gray = "\u001b[90m";

        static void Main(string[] args)
        {
            string file = null;
            bool compact = false, minify = false, colors = true, page = false;
            string outputFile = null, filter = null;

            for (int i = 0; i < args.Length; i++)
            {
                switch (args[i])
                {
                    case "-c": compact = true; break;
                    case "-m": minify = true; break;
                    case "-p": break;
                    case "-o": if (i + 1 < args.Length) outputFile = args[++i]; break;
                    case "--no-colors": colors = false; break;
                    case "--filter": if (i + 1 < args.Length) filter = args[++i]; break;
                    case "--page": page = true; break;
                    default:
                        if (!args[i].StartsWith("-")) file = args[i];
                        break;
                }
            }

            string jsonText;
            if (file != null)
            {
                jsonText = File.ReadAllText(file);
            }
            else
            {
                jsonText = Console.In.ReadToEnd();
            }

            var options = new JsonDocumentOptions { AllowTrailingCommas = true, CommentHandling = JsonCommentHandling.Skip };
            using var doc = JsonDocument.Parse(jsonText, options);
            var root = doc.RootElement;

            // Фильтрация
            if (!string.IsNullOrEmpty(filter))
            {
                root = FilterElement(root, filter);
            }

            var output = RenderElement(root, compact || minify, colors, minify);

            if (!string.IsNullOrEmpty(outputFile))
            {
                File.WriteAllText(outputFile, output);
                Console.WriteLine($"Результат сохранён в {outputFile}");
            }
            else
            {
                if (page)
                {
                    // Используем less
                    var psi = new System.Diagnostics.ProcessStartInfo("less", "-R")
                    {
                        RedirectStandardInput = true,
                        UseShellExecute = false
                    };
                    using var p = System.Diagnostics.Process.Start(psi);
                    using (var sw = p.StandardInput)
                    {
                        sw.Write(output);
                    }
                    p.WaitForExit();
                }
                else
                {
                    Console.WriteLine(output);
                }
            }
        }

        private static JsonElement FilterElement(JsonElement el, string path)
        {
            var parts = path.Split('.');
            var current = el;
            foreach (var part in parts)
            {
                if (current.ValueKind == JsonValueKind.Object && current.TryGetProperty(part, out var next))
                {
                    current = next;
                }
                else
                {
                    throw new Exception($"Путь '{path}' не найден");
                }
            }
            return current;
        }

        private static string RenderElement(JsonElement el, bool compact, bool colors, bool minify)
        {
            var sb = new StringBuilder();
            RenderElementRecursive(el, sb, compact ? "" : "  ", 0, colors, minify);
            return sb.ToString();
        }

        private static void RenderElementRecursive(JsonElement el, StringBuilder sb, string indent, int level, bool colors, bool minify)
        {
            string currentIndent = minify ? "" : string.Concat(Enumerable.Repeat(indent, level));
            switch (el.ValueKind)
            {
                case JsonValueKind.Object:
                    sb.Append('{');
                    if (!minify) sb.Append('\n');
                    bool first = true;
                    foreach (var prop in el.EnumerateObject())
                    {
                        if (!first) { sb.Append(','); if (!minify) sb.Append('\n'); }
                        first = false;
                        if (!minify) sb.Append(currentIndent + indent);
                        string key = $"\"{prop.Name}\"";
                        if (colors) key = Cyan + key + Reset;
                        sb.Append(key + ": ");
                        RenderElementRecursive(prop.Value, sb, indent, level + 1, colors, minify);
                    }
                    if (!minify) sb.Append('\n' + currentIndent);
                    sb.Append('}');
                    break;

                case JsonValueKind.Array:
                    sb.Append('[');
                    if (!minify) sb.Append('\n');
                    first = true;
                    foreach (var item in el.EnumerateArray())
                    {
                        if (!first) { sb.Append(','); if (!minify) sb.Append('\n'); }
                        first = false;
                        if (!minify) sb.Append(currentIndent + indent);
                        RenderElementRecursive(item, sb, indent, level + 1, colors, minify);
                    }
                    if (!minify) sb.Append('\n' + currentIndent);
                    sb.Append(']');
                    break;

                case JsonValueKind.String:
                    string str = $"\"{el.GetString()}\"";
                    sb.Append(colors ? Green + str + Reset : str);
                    break;

                case JsonValueKind.Number:
                    string num = el.GetRawText();
                    sb.Append(colors ? Yellow + num + Reset : num);
                    break;

                case JsonValueKind.True:
                    sb.Append(colors ? Magenta + "true" + Reset : "true");
                    break;

                case JsonValueKind.False:
                    sb.Append(colors ? Magenta + "false" + Reset : "false");
                    break;

                case JsonValueKind.Null:
                    sb.Append(colors ? Red + "null" + Reset : "null");
                    break;

                default:
                    sb.Append(el.GetRawText());
                    break;
            }
        }
    }
}
