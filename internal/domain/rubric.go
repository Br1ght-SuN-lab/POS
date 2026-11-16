package domain

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"
)

type Criterion struct {
	ID       string
	Name     string
	MaxScore float64
	Weight   float64
}

type rubric struct {
	criteria map[string]Criterion
}

func (r rubric) addCriterion(id, name string, maxScore, weight float64) error {
	if err := validateIdentity(id, name); err != nil {
		return err
	}
	if _, exists := r.criteria[id]; exists {
		return fmt.Errorf("критерий %q уже существует", id)
	}
	if !finite(maxScore) || maxScore <= 0 || !finite(weight) || weight <= 0 {
		return fmt.Errorf("максимальный балл и вес должны быть конечными положительными числами")
	}
	r.criteria[id] = Criterion{ID: id, Name: strings.TrimSpace(name), MaxScore: maxScore, Weight: weight}
	return nil
}

func (r rubric) listCriteria() []Criterion {
	criteria := make([]Criterion, 0, len(r.criteria))
	for _, criterion := range r.criteria {
		criteria = append(criteria, criterion)
	}
	slices.SortFunc(criteria, func(a, b Criterion) int { return cmp.Compare(a.ID, b.ID) })
	return criteria
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
