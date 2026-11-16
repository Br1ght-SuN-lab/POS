package infrastructure

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"pos/internal/application"
	"pos/internal/domain"
)

type Console struct {
	in      io.Reader
	out     io.Writer
	service *application.Service
}

func NewConsole(in io.Reader, out io.Writer, service *application.Service) *Console {
	return &Console{in: in, out: out, service: service}
}

const help = `ПОС — справка по командам

Подставьте свои значения вместо данных из примеров.
Символ | разделяет параметры. ID студента и критерия видны в списках.

Студенты и критерии
  student-add s01 | Анна Иванова       Добавить студента с ID s01
  students                             Показать студентов
  criterion-add quiz | Тест | 10 | 40  Добавить тест (максимум 10, вес 40)
  criteria                             Показать критерии

Оценки
  grade-set s01 | quiz | 9             Выставить или исправить оценку
  grade-set s01 | quiz | 9 | Хорошо    Сохранить оценку с комментарием
  grades s01                           Показать оценки студента
  grade-cancel s01 | quiz              Отменить оценку за тест

Итоги и экспорт
  result s01                           Показать итог студента
  report                               Показать ведомость и статистику
  export results.csv                   Сохранить ведомость в CSV

По умолчанию режим final учитывает все критерии.
Режим progress учитывает только критерии с оценками.
  result s01 | progress
  report progress
  export results.csv | progress

  help                                 Показать эту справку
  exit                                 Выйти
`

func (c *Console) Run(script bool) error {
	fmt.Fprintln(c.out, "ПОС. Введите help для списка команд.")
	scanner := bufio.NewScanner(c.in)
	for {
		if !script {
			fmt.Fprint(c.out, "> ")
		}
		if !scanner.Scan() {
			return scanner.Err()
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		command, parameters, _ := strings.Cut(line, " ")
		var args []string
		if parameters != "" {
			args = strings.Split(parameters, "|")
			for i := range args {
				args[i] = strings.TrimSpace(args[i])
			}
		}
		if command == "exit" && len(args) == 0 {
			return nil
		}
		if err := c.runCommand(command, args); err != nil {
			if script {
				return fmt.Errorf("команда %q: %w", command, err)
			}
			fmt.Fprintln(c.out, "Ошибка:", err)
		}
	}
}

func (c *Console) runCommand(command string, args []string) error {
	// Проверяем число обязательных и необязательных параметров перед выполнением команды.
	var minArgs, maxArgs int
	switch command {
	case "student-add", "grade-cancel":
		minArgs, maxArgs = 2, 2
	case "criterion-add":
		minArgs, maxArgs = 4, 4
	case "grade-set":
		minArgs, maxArgs = 3, 4
	case "grades":
		minArgs, maxArgs = 1, 1
	case "result", "export":
		minArgs, maxArgs = 1, 2
	case "report":
		minArgs, maxArgs = 0, 1
	case "help", "exit", "students", "criteria":
	default:
		return fmt.Errorf("неизвестная команда %q, введите help", command)
	}
	if len(args) < minArgs || len(args) > maxArgs {
		return fmt.Errorf("неверное число параметров для %s, введите help", command)
	}

	switch command {
	case "help":
		fmt.Fprint(c.out, help)
	case "student-add":
		return c.confirm(c.service.AddStudent(args[0], args[1]), "Студент добавлен.")
	case "students":
		students, err := c.service.Students()
		if err != nil {
			return err
		}
		fmt.Fprintf(c.out, "Студенты (%d)\n", len(students))
		for _, student := range students {
			fmt.Fprintf(c.out, "  %s — %s\n", student.ID, student.Name)
		}
	case "criterion-add":
		maxScore, err := parseNumber(args[2])
		if err != nil {
			return err
		}
		weight, err := parseNumber(args[3])
		if err != nil {
			return err
		}
		return c.confirm(c.service.AddCriterion(args[0], args[1], maxScore, weight), "Критерий добавлен.")
	case "criteria":
		criteria, err := c.service.Criteria()
		if err != nil {
			return err
		}
		fmt.Fprintf(c.out, "Критерии (%d)\n", len(criteria))
		for _, criterion := range criteria {
			fmt.Fprintf(c.out, "  %s — %s, максимум %g, вес %g\n", criterion.ID, criterion.Name, criterion.MaxScore, criterion.Weight)
		}
	case "grade-set":
		points, err := parseNumber(args[2])
		if err != nil {
			return err
		}
		comment := optionalArg(args[3:], "")
		return c.confirm(c.service.SetEvaluation(args[0], args[1], points, comment), "Оценка сохранена.")
	case "grade-cancel":
		return c.confirm(c.service.CancelEvaluation(args[0], args[1]), "Оценка отменена.")
	case "grades":
		evaluations, err := c.service.Evaluations(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(c.out, "Оценки %s (%d)\n", args[0], len(evaluations))
		for _, evaluation := range evaluations {
			fmt.Fprintf(c.out, "  %s — %g", evaluation.CriterionID, evaluation.Points())
			if evaluation.Comment != "" {
				fmt.Fprintf(c.out, " (%s)", evaluation.Comment)
			}
			fmt.Fprintln(c.out)
		}
	case "result":
		mode := optionalArg(args[1:], "final")
		result, err := c.service.Result(args[0], mode)
		if err != nil {
			return err
		}
		c.printResult(args[0]+" ["+mode+"]", result)
	case "report":
		report, err := c.service.Report(optionalArg(args, "final"))
		if err != nil {
			return err
		}
		fmt.Fprintf(c.out, "Ведомость [%s], студентов: %d\n", report.Mode, len(report.Rows))
		for _, row := range report.Rows {
			c.printResult(row.Student.ID+" — "+row.Student.Name, row.Result)
		}
		if len(report.Rows) > 0 {
			fmt.Fprintf(c.out, "Среднее %.2f%%, минимум %.2f%%, максимум %.2f%%\n", report.Mean, report.Min, report.Max)
		}
	case "export":
		return c.confirm(c.service.Export(args[0], optionalArg(args[1:], "final")), "Ведомость экспортирована в "+args[0])
	}
	return nil
}

func parseNumber(input string) (float64, error) {
	value, err := strconv.ParseFloat(strings.ReplaceAll(input, ",", "."), 64)
	if err != nil {
		return 0, fmt.Errorf("ожидалось число, получено %q", input)
	}
	return value, nil
}

func optionalArg(args []string, fallback string) string {
	if len(args) == 0 {
		return fallback
	}
	return args[0]
}

func (c *Console) printResult(label string, result domain.Result) {
	fmt.Fprintf(c.out, "  %s — %.2f%% (оценено %d/%d)\n", label, result.Percent, result.Graded, result.Criteria)
}

func (c *Console) confirm(err error, message string) error {
	if err != nil {
		return err
	}
	fmt.Fprintln(c.out, message)
	return nil
}
