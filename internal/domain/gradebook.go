package domain

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
	"unicode"
)

type Student struct {
	ID   string
	Name string
}

// Gradebook хранит ведомость одного курса и следит за связями между её записями.
type Gradebook struct {
	students    map[string]Student
	rubric      rubric
	evaluations map[evaluationKey]Evaluation
}

func NewGradebook() *Gradebook {
	return &Gradebook{
		students:    make(map[string]Student),
		rubric:      rubric{criteria: make(map[string]Criterion)},
		evaluations: make(map[evaluationKey]Evaluation),
	}
}

func validateIdentity(id, name string) error {
	if id == "" || strings.ContainsRune(id, '|') || strings.ContainsFunc(id, unicode.IsSpace) {
		return fmt.Errorf("идентификатор должен быть непустым и без пробелов или символа |")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("имя или название не может быть пустым")
	}
	return nil
}

func (g *Gradebook) AddStudent(id, name string) error {
	if err := validateIdentity(id, name); err != nil {
		return err
	}
	if _, exists := g.students[id]; exists {
		return fmt.Errorf("студент %q уже существует", id)
	}
	g.students[id] = Student{ID: id, Name: strings.TrimSpace(name)}
	return nil
}

func (g *Gradebook) AddCriterion(id, name string, maxScore, weight float64) error {
	return g.rubric.addCriterion(id, name, maxScore, weight)
}

func (g *Gradebook) Student(id string) (Student, error) {
	student, ok := g.students[id]
	if !ok {
		return Student{}, fmt.Errorf("студент %q не найден", id)
	}
	return student, nil
}

// SetEvaluation выставляет или заменяет оценку после проверки студента и рубрикатора.
func (g *Gradebook) SetEvaluation(studentID, criterionID string, points float64, comment string) error {
	if _, err := g.Student(studentID); err != nil {
		return err
	}
	evaluation, err := g.rubric.newEvaluation(studentID, criterionID, points, comment)
	if err != nil {
		return err
	}
	g.evaluations[evaluationKey{studentID, criterionID}] = evaluation
	return nil
}

func (g *Gradebook) CancelEvaluation(studentID, criterionID string) error {
	key := evaluationKey{studentID, criterionID}
	if _, ok := g.evaluations[key]; !ok {
		return fmt.Errorf("оценка студента %q по критерию %q не найдена", studentID, criterionID)
	}
	delete(g.evaluations, key)
	return nil
}

func (g *Gradebook) Students() []Student {
	students := make([]Student, 0, len(g.students))
	for _, student := range g.students {
		students = append(students, student)
	}
	slices.SortFunc(students, func(a, b Student) int { return cmp.Compare(a.ID, b.ID) })
	return students
}

func (g *Gradebook) Criteria() []Criterion {
	return g.rubric.listCriteria()
}

func (g *Gradebook) Evaluations(studentID string) ([]Evaluation, error) {
	if _, err := g.Student(studentID); err != nil {
		return nil, err
	}
	evaluations := make([]Evaluation, 0)
	for _, criterion := range g.Criteria() {
		if evaluation, ok := g.evaluations[evaluationKey{studentID, criterion.ID}]; ok {
			evaluations = append(evaluations, evaluation)
		}
	}
	return evaluations, nil
}
