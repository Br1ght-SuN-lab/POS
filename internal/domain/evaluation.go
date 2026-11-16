package domain

import (
	"fmt"
	"strings"
)

// Score хранит балл, допустимый диапазон которого задаёт критерий.
type Score struct {
	points float64
}

type Evaluation struct {
	StudentID   string
	CriterionID string
	Comment     string
	score       Score
}

func (e Evaluation) Points() float64 { return e.score.points }

type evaluationKey struct {
	studentID   string
	criterionID string
}

// newEvaluation создаёт оценку после проверки критерия и диапазона балла.
func (r rubric) newEvaluation(studentID, criterionID string, points float64, comment string) (Evaluation, error) {
	criterion, ok := r.criteria[criterionID]
	if !ok {
		return Evaluation{}, fmt.Errorf("критерий %q отсутствует в рубрикаторе", criterionID)
	}
	if !finite(points) || points < 0 || points > criterion.MaxScore {
		return Evaluation{}, fmt.Errorf("балл по критерию %q должен быть числом от 0 до %g", criterionID, criterion.MaxScore)
	}
	return Evaluation{
		StudentID: studentID, CriterionID: criterionID, Comment: strings.TrimSpace(comment),
		score: Score{points: points},
	}, nil
}
