package domain

type Result struct {
	Percent  float64
	Graded   int
	Criteria int
}

// ResultStrategy определяет, какие критерии входят в расчёт общего веса.
type ResultStrategy interface {
	Includes(graded bool) bool
}

type FinalStrategy struct{}

func (FinalStrategy) Includes(bool) bool { return true }

type ProgressStrategy struct{}

func (ProgressStrategy) Includes(graded bool) bool { return graded }

// CalculateResult считает процент по текущим оценкам студента и рубрикатору.
func CalculateResult(book *Gradebook, studentID string, strategy ResultStrategy) (Result, error) {
	if _, err := book.Student(studentID); err != nil {
		return Result{}, err
	}

	criteria := book.Criteria()
	result := Result{Criteria: len(criteria)}
	weightScale := 0.0
	for _, criterion := range criteria {
		_, graded := book.evaluations[evaluationKey{studentID, criterion.ID}]
		if graded {
			result.Graded++
		}
		if strategy.Includes(graded) {
			weightScale = max(weightScale, criterion.Weight)
		}
	}
	if weightScale == 0 {
		return result, nil
	}

	weightedScore, totalWeight := 0.0, 0.0
	for _, criterion := range criteria {
		evaluation, graded := book.evaluations[evaluationKey{studentID, criterion.ID}]
		if strategy.Includes(graded) {
			// Деление на максимальный вес сохраняет пропорции и защищает сумму от переполнения.
			weight := criterion.Weight / weightScale
			totalWeight += weight
			weightedScore += (evaluation.Points() / criterion.MaxScore) * weight
		}
	}
	result.Percent = 100 * (weightedScore / totalWeight)
	return result, nil
}
