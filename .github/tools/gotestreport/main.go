// gotestreport превращает вывод `go test -json` в понятный отчёт для GitHub Actions:
//
//   - в лог шага — итог по пакетам, список упавших тестов и полный вывод каждого
//     из них в сворачиваемых группах;
//   - аннотации ::error на строках тестов, где случилась ошибка, — они видны
//     на странице запуска и прямо в коде пулл-реквеста;
//   - Markdown-отчёт в $GITHUB_STEP_SUMMARY: таблицы, упавшие тесты с выводом,
//     пропущенные и самые медленные тесты.
//
// Код выхода 1, если хоть что-то упало или не собралось, — шаг CI краснеет.
// Только стандартная библиотека, поэтому запускается без go.mod:
//
//	go run .github/tools/gotestreport/main.go -module taygant_backend -dir backend report.json
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
)

// event — одна строка `go test -json` (см. `go doc test2json`).
type event struct {
	Action      string
	Package     string
	Test        string
	Elapsed     float64
	Output      string
	ImportPath  string // у build-output / build-fail
	FailedBuild string // у fail пакета, который не собрался
}

type outputLine struct {
	test string // "" — вывод самого пакета (TestMain, паника вне теста)
	text string
}

type testResult struct {
	status  string // pass | fail | skip
	elapsed float64
}

type pkgResult struct {
	name     string
	status   string
	elapsed  float64
	lines    []outputLine
	tests    map[string]*testResult
	topLevel []string // тесты верхнего уровня в порядке запуска
	buildKey string   // непусто, если пакет не собрался
}

type report struct {
	pkgs  []*pkgResult
	index map[string]*pkgResult
	build map[string][]string // ImportPath → вывод компилятора
	stray []string            // строки не в формате JSON
}

var (
	// "    auth_test.go:85: ожидался код 400..." — место ошибки внутри теста.
	testFailureLine = regexp.MustCompile(`^\s*([\w.\-]+_test\.go):(\d+): (.*)$`)
	// "internal/api/x_test.go:12:3: undefined: foo" — ошибка компиляции.
	compileErrorLine = regexp.MustCompile(`^(\S+\.go):(\d+):(?:(\d+):)? (.*)$`)
)

func main() {
	module := flag.String("module", "", "имя Go-модуля из go.mod — чтобы превратить пакет в путь к файлу")
	dir := flag.String("dir", ".", "каталог модуля относительно корня репозитория")
	title := flag.String("title", "Тесты", "заголовок отчёта")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "использование: gotestreport [-module имя] [-dir каталог] [-title заголовок] report.json")
		os.Exit(2)
	}

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "не удалось открыть отчёт:", err)
		os.Exit(2)
	}
	rep, err := parse(file)
	file.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, "не удалось прочитать отчёт:", err)
		os.Exit(2)
	}

	w := writer{module: *module, dir: *dir, gha: os.Getenv("GITHUB_ACTIONS") == "true"}
	ok := w.printLog(rep)
	if summaryPath := os.Getenv("GITHUB_STEP_SUMMARY"); summaryPath != "" {
		if err := w.writeSummary(summaryPath, *title, rep); err != nil {
			fmt.Fprintln(os.Stderr, "не удалось записать отчёт в summary:", err)
		}
	}
	if !ok {
		os.Exit(1)
	}
}

func parse(r io.Reader) (*report, error) {
	rep := &report{index: map[string]*pkgResult{}, build: map[string][]string{}}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for sc.Scan() {
		raw := sc.Bytes()
		if len(raw) == 0 {
			continue
		}
		var e event
		if raw[0] != '{' || json.Unmarshal(raw, &e) != nil {
			rep.stray = append(rep.stray, string(raw))
			continue
		}
		text := strings.TrimRight(e.Output, "\n")

		switch e.Action {
		case "build-output":
			rep.build[e.ImportPath] = append(rep.build[e.ImportPath], text)
			continue
		case "build-fail", "start":
			continue
		}
		if e.Package == "" {
			continue
		}

		p := rep.pkg(e.Package)
		if e.Test == "" {
			switch e.Action {
			case "output":
				p.lines = append(p.lines, outputLine{"", text})
			case "pass", "fail", "skip":
				p.status, p.elapsed, p.buildKey = e.Action, e.Elapsed, e.FailedBuild
			}
			continue
		}

		t, seen := p.tests[e.Test]
		if !seen {
			t = &testResult{}
			p.tests[e.Test] = t
			if !strings.Contains(e.Test, "/") {
				p.topLevel = append(p.topLevel, e.Test)
			}
		}
		switch e.Action {
		case "output":
			p.lines = append(p.lines, outputLine{e.Test, text})
		case "pass", "fail", "skip":
			t.status, t.elapsed = e.Action, e.Elapsed
		}
	}
	return rep, sc.Err()
}

// withTests — пакеты, о которых есть что сказать; пакеты без _test.go-файлов
// перечисляются одной строкой, а не пустыми строками таблицы.
func (r *report) withTests() (shown []*pkgResult, empty []string) {
	for _, p := range r.pkgs {
		if len(p.topLevel) == 0 && p.buildKey == "" && p.status != "fail" {
			empty = append(empty, p.name)
			continue
		}
		shown = append(shown, p)
	}
	return shown, empty
}

func (r *report) pkg(name string) *pkgResult {
	if p, ok := r.index[name]; ok {
		return p
	}
	p := &pkgResult{name: name, tests: map[string]*testResult{}}
	r.index[name] = p
	r.pkgs = append(r.pkgs, p)
	return p
}

// counts — число тестов верхнего уровня по статусам (подтесты не считаются отдельно).
func (p *pkgResult) counts() (passed, failed, skipped int) {
	for _, name := range p.topLevel {
		switch p.tests[name].status {
		case "pass":
			passed++
		case "fail":
			failed++
		case "skip":
			skipped++
		}
	}
	return
}

func (p *pkgResult) withStatus(status string) []string {
	var names []string
	for _, name := range p.topLevel {
		if p.tests[name].status == status {
			names = append(names, name)
		}
	}
	return names
}

// testOutput — вывод теста вместе с его подтестами, без служебных строк === RUN.
func (p *pkgResult) testOutput(test string) []string {
	var out []string
	for _, l := range p.lines {
		if l.test != test && !strings.HasPrefix(l.test, test+"/") {
			continue
		}
		// Служебные строки и прошедшие подтесты только мешают найти причину падения.
		if s := strings.TrimSpace(l.text); strings.HasPrefix(s, "=== ") || strings.HasPrefix(s, "--- PASS") {
			continue
		}
		out = append(out, l.text)
	}
	return out
}

// packageOutput — вывод пакета вне тестов: паника, ошибка TestMain, сообщения о пропуске.
func (p *pkgResult) packageOutput() []string {
	var out []string
	for _, l := range p.lines {
		if l.test == "" {
			out = append(out, l.text)
		}
	}
	return out
}

// firstError — первая содержательная строка ошибки для краткой подписи.
func firstError(lines []string) string {
	for _, l := range lines {
		if m := testFailureLine.FindStringSubmatch(l); m != nil {
			return m[3]
		}
	}
	for _, l := range lines {
		if s := strings.TrimSpace(l); s != "" && !strings.HasPrefix(s, "--- FAIL") {
			return s
		}
	}
	return "подробности в логе"
}

// ---------- Вывод ----------

type writer struct {
	module, dir string
	gha         bool
}

func (w writer) short(pkg string) string {
	if w.module == "" {
		return pkg
	}
	if pkg == w.module {
		return "(корень модуля)"
	}
	return strings.TrimPrefix(pkg, w.module+"/")
}

func (w writer) shortList(pkgs []string) string {
	names := make([]string, len(pkgs))
	for i, p := range pkgs {
		names[i] = w.short(p)
	}
	return strings.Join(names, ", ")
}

// repoPath превращает пакет и имя файла в путь от корня репозитория — для аннотаций.
func (w writer) repoPath(pkg, file string) string {
	rel := strings.TrimPrefix(pkg, w.module)
	return path.Join(w.dir, rel, file)
}

func (w writer) group(title string, lines []string) {
	if w.gha {
		fmt.Printf("::group::%s\n", title)
	} else {
		fmt.Printf("── %s\n", title)
	}
	for _, l := range lines {
		fmt.Println(l)
	}
	if w.gha {
		fmt.Println("::endgroup::")
	}
}

func (w writer) annotate(file, line, title, message string) {
	if !w.gha {
		return
	}
	props := "title=" + escapeProperty(title)
	if file != "" {
		props = "file=" + escapeProperty(file) + ",line=" + line + "," + props
	}
	fmt.Printf("::error %s::%s\n", props, escapeData(message))
}

// printLog печатает отчёт в лог шага и возвращает false, если что-то упало.
func (w writer) printLog(rep *report) bool {
	var passed, failed, skipped, broken int
	for _, p := range rep.pkgs {
		ps, fl, sk := p.counts()
		passed, failed, skipped = passed+ps, failed+fl, skipped+sk
		if p.buildKey != "" || (p.status == "fail" && fl == 0) {
			broken++
		}
	}

	fmt.Println()
	fmt.Printf("Итог: прошло %d · упало %d · пропущено %d · пакетов с ошибкой сборки или запуска: %d\n\n", passed, failed, skipped, broken)
	shown, empty := rep.withTests()
	fmt.Printf("%-44s %8s %7s %10s %9s\n", "Пакет", "Прошло", "Упало", "Пропущено", "Время")
	for _, p := range shown {
		ps, fl, sk := p.counts()
		fmt.Printf("%s %-42s %8d %7d %10d %8.1fс\n", statusIcon(p), w.short(p.name), ps, fl, sk, p.elapsed)
	}
	if len(empty) > 0 {
		fmt.Printf("Без тестов: %s\n", w.shortList(empty))
	}
	fmt.Println()

	if len(rep.pkgs) == 0 {
		fmt.Println("❌ В отчёте нет ни одного пакета — go test не запустился. Смотрите вывод предыдущего шага.")
		w.annotate("", "", "go test не запустился", "Отчёт пустой — смотрите лог шага с тестами")
		return false
	}

	ok := true
	for _, p := range rep.pkgs {
		if p.buildKey != "" {
			ok = false
			lines := rep.build[p.buildKey]
			w.group("🧱 Не собрался пакет "+w.short(p.name), lines)
			for _, l := range lines {
				if m := compileErrorLine.FindStringSubmatch(l); m != nil {
					w.annotate(path.Join(w.dir, m[1]), m[2], "Ошибка компиляции", m[4])
				}
			}
			continue
		}
		failedTests := p.withStatus("fail")
		if p.status == "fail" && len(failedTests) == 0 {
			ok = false
			w.group("💥 Пакет "+w.short(p.name)+" упал вне тестов", p.packageOutput())
			w.annotate("", "", "Пакет "+w.short(p.name)+" упал", firstError(p.packageOutput()))
		}
		for _, name := range failedTests {
			ok = false
			out := p.testOutput(name)
			w.group("❌ "+w.short(p.name)+" · "+name, out)
			annotated := false
			for _, l := range out {
				if m := testFailureLine.FindStringSubmatch(l); m != nil {
					w.annotate(w.repoPath(p.name, m[1]), m[2], name, m[3])
					annotated = true
				}
			}
			if !annotated {
				w.annotate("", "", name, firstError(out))
			}
		}
	}

	if len(rep.stray) > 0 {
		w.group("Вывод не в формате JSON (обычно сообщения go до запуска тестов)", rep.stray)
	}
	if ok {
		fmt.Println("✅ Все тесты прошли.")
	} else {
		fmt.Println("❌ Есть упавшие тесты — полный вывод каждого в группах выше, сводка на странице запуска (Summary).")
	}
	return ok
}

func statusIcon(p *pkgResult) string {
	_, failed, skipped := p.counts()
	switch {
	case p.buildKey != "" || p.status == "fail" || failed > 0:
		return "❌"
	case len(p.topLevel) == 0:
		return "➖"
	case skipped == len(p.topLevel):
		return "⏭️"
	default:
		return "✅"
	}
}

func (w writer) writeSummary(summaryPath, title string, rep *report) error {
	var b strings.Builder
	var passed, failed, skipped int
	var elapsed float64
	for _, p := range rep.pkgs {
		ps, fl, sk := p.counts()
		passed, failed, skipped = passed+ps, failed+fl, skipped+sk
		elapsed += p.elapsed
	}

	fmt.Fprintf(&b, "## %s\n\n", title)
	fmt.Fprintf(&b, "| ✅ Прошло | ❌ Упало | ⏭️ Пропущено | ⏱️ Время |\n|:-:|:-:|:-:|:-:|\n| **%d** | **%d** | **%d** | %.1f с |\n\n", passed, failed, skipped, elapsed)

	shown, empty := rep.withTests()
	b.WriteString("| | Пакет | Прошло | Упало | Пропущено | Время |\n|:-:|---|:-:|:-:|:-:|--:|\n")
	for _, p := range shown {
		ps, fl, sk := p.counts()
		fmt.Fprintf(&b, "| %s | `%s` | %d | %d | %d | %.1f с |\n", statusIcon(p), w.short(p.name), ps, fl, sk, p.elapsed)
	}
	b.WriteString("\n")
	if len(empty) > 0 {
		fmt.Fprintf(&b, "Пакеты без тестов: %s\n\n", w.shortList(empty))
	}

	for _, p := range rep.pkgs {
		if p.buildKey != "" {
			fmt.Fprintf(&b, "### 🧱 Не собрался пакет `%s`\n\n", w.short(p.name))
			codeBlock(&b, rep.build[p.buildKey])
		} else if p.status == "fail" && len(p.withStatus("fail")) == 0 {
			fmt.Fprintf(&b, "### 💥 Пакет `%s` упал вне тестов\n\n", w.short(p.name))
			codeBlock(&b, p.packageOutput())
		}
	}

	if failed > 0 {
		fmt.Fprintf(&b, "### ❌ Упавшие тесты (%d)\n\n", failed)
		for _, p := range rep.pkgs {
			for _, name := range p.withStatus("fail") {
				out := p.testOutput(name)
				fmt.Fprintf(&b, "<details><summary><code>%s</code> · <b>%s</b> — %s</summary>\n\n", w.short(p.name), name, htmlEscape(firstError(out)))
				codeBlock(&b, out)
				b.WriteString("</details>\n\n")
			}
		}
	}

	if skipped > 0 {
		fmt.Fprintf(&b, "<details><summary>⏭️ Пропущенные тесты (%d)</summary>\n\n", skipped)
		for _, p := range rep.pkgs {
			for _, name := range p.withStatus("skip") {
				fmt.Fprintf(&b, "- `%s` · %s\n", w.short(p.name), name)
			}
		}
		b.WriteString("\n</details>\n\n")
	}

	type timed struct {
		pkg, name string
		elapsed   float64
	}
	var slow []timed
	for _, p := range rep.pkgs {
		for _, name := range p.topLevel {
			slow = append(slow, timed{w.short(p.name), name, p.tests[name].elapsed})
		}
	}
	sort.Slice(slow, func(i, j int) bool { return slow[i].elapsed > slow[j].elapsed })
	if len(slow) > 5 {
		slow = slow[:5]
	}
	if len(slow) > 0 && slow[0].elapsed > 0 {
		b.WriteString("<details><summary>🐢 Самые медленные тесты</summary>\n\n| Тест | Пакет | Время |\n|---|---|--:|\n")
		for _, s := range slow {
			fmt.Fprintf(&b, "| %s | `%s` | %.2f с |\n", s.name, s.pkg, s.elapsed)
		}
		b.WriteString("\n</details>\n\n")
	}

	f, err := os.OpenFile(summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(b.String())
	return err
}

// codeBlock выводит строки в блоке кода, обрезая слишком длинный вывод:
// у summary есть лимит размера, а полный текст всё равно лежит в логе шага.
func codeBlock(b *strings.Builder, lines []string) {
	const limit = 80
	b.WriteString("```text\n")
	for i, l := range lines {
		if i == limit {
			fmt.Fprintf(b, "… ещё %d строк — полный вывод в логе шага\n", len(lines)-limit)
			break
		}
		b.WriteString(strings.ReplaceAll(l, "```", "ʼʼʼ"))
		b.WriteString("\n")
	}
	b.WriteString("```\n\n")
}

func htmlEscape(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// escapeData и escapeProperty — экранирование для workflow-команд GitHub.
func escapeData(s string) string {
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A").Replace(s)
}

func escapeProperty(s string) string {
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A", ":", "%3A", ",", "%2C").Replace(s)
}
