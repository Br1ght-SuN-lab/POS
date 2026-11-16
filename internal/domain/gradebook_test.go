package domain

import (
	"math"
	"testing"
)

func TestEvaluationInvariants(t *testing.T) {
	book := testGradebook(t)
	for _, valid := range []float64{0, 10, 7.5} {
		mustSucceed(t, book.SetEvaluation("s1", "quiz", valid, ""))
	}

	cases := []struct {
		name      string
		student   string
		criterion string
		points    float64
	}{
		{"negative", "s1", "quiz", -0.1},
		{"above maximum", "s1", "quiz", 10.1},
		{"NaN", "s1", "quiz", math.NaN()},
		{"infinity", "s1", "quiz", math.Inf(1)},
		{"missing criterion", "s1", "unknown", 5},
		{"missing student", "unknown", "quiz", 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := book.SetEvaluation(tc.student, tc.criterion, tc.points, ""); err == nil {
				t.Fatal("некорректная оценка принята")
			}
			evaluations, err := book.Evaluations("s1")
			mustSucceed(t, err)
			if len(evaluations) != 1 || evaluations[0].Points() != 7.5 {
				t.Fatalf("неверное состояние после отказа: %+v", evaluations)
			}
		})
	}
}

func TestRubricInvariants(t *testing.T) {
	book := testGradebook(t)
	cases := []struct{ maxScore, weight float64 }{
		{0, 10},
		{10, 0},
		{10, -1},
		{10, math.Inf(1)},
		{math.NaN(), 10},
	}
	for _, tc := range cases {
		if err := book.AddCriterion("invalid", "Некорректный", tc.maxScore, tc.weight); err == nil {
			t.Errorf("принят критерий с максимумом %g и весом %g", tc.maxScore, tc.weight)
		}
	}
	if err := book.AddCriterion("quiz", "Дубликат", 20, 1); err == nil {
		t.Error("принят повторный идентификатор критерия")
	}
	if err := book.AddStudent("s1", "Дубликат"); err == nil {
		t.Error("принят повторный идентификатор студента")
	}
	if got := len(book.Criteria()); got != 2 {
		t.Errorf("после отклонения критериев осталось %d, хотели 2", got)
	}
}

func TestCalculateResult(t *testing.T) {
	book := testGradebook(t)
	checkResult(t, book, ProgressStrategy{}, 0, 0, 2)
	mustSucceed(t, book.SetEvaluation("s1", "project", 80, ""))
	checkResult(t, book, FinalStrategy{}, 48, 1, 2)
	checkResult(t, book, ProgressStrategy{}, 80, 1, 2)
	mustSucceed(t, book.SetEvaluation("s1", "quiz", 0, ""))
	checkResult(t, book, ProgressStrategy{}, 48, 2, 2)
	mustSucceed(t, book.SetEvaluation("s1", "quiz", 9, "исправлено"))
	checkResult(t, book, FinalStrategy{}, 84, 2, 2)
	checkResult(t, book, ProgressStrategy{}, 84, 2, 2)
	mustSucceed(t, book.CancelEvaluation("s1", "quiz"))
	checkResult(t, book, FinalStrategy{}, 48, 1, 2)
	if err := book.CancelEvaluation("s1", "quiz"); err == nil {
		t.Error("отмена несуществующей оценки должна вернуть ошибку")
	}
	mustSucceed(t, book.AddCriterion("practice", "Практика", 10, 100))
	checkResult(t, book, FinalStrategy{}, 24, 1, 3)
	checkResult(t, book, ProgressStrategy{}, 80, 1, 3)
}

func testGradebook(t *testing.T) *Gradebook {
	t.Helper()
	book := NewGradebook()
	mustSucceed(t, book.AddStudent("s1", "Студент"))
	mustSucceed(t, book.AddCriterion("project", "Проект", 100, 60))
	mustSucceed(t, book.AddCriterion("quiz", "Тест", 10, 40))
	return book
}

func checkResult(t *testing.T, book *Gradebook, strategy ResultStrategy, percent float64, graded, criteria int) {
	t.Helper()
	result, err := CalculateResult(book, "s1", strategy)
	mustSucceed(t, err)
	if math.Abs(result.Percent-percent) > 1e-9 || result.Graded != graded || result.Criteria != criteria {
		t.Errorf("результат %+v, хотели %g%% (%d/%d)", result, percent, graded, criteria)
	}
}

func mustSucceed(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
