package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"pos/internal/domain"
)

type JSONStore struct {
	path string
}

func NewJSONStore(path string) *JSONStore {
	return &JSONStore{path: path}
}

type savedEvaluation struct {
	StudentID   string  `json:"student_id"`
	CriterionID string  `json:"criterion_id"`
	Points      float64 `json:"points"`
	Comment     string  `json:"comment,omitempty"`
}

type snapshot struct {
	Students    []domain.Student   `json:"students"`
	Criteria    []domain.Criterion `json:"criteria"`
	Evaluations []savedEvaluation  `json:"evaluations"`
}

func (s *JSONStore) Read() (*domain.Gradebook, error) {
	book := domain.NewGradebook()
	content, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return book, nil
	}
	if err != nil {
		return nil, err
	}

	var data snapshot
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("чтение ведомости %q: %w", s.path, err)
	}
	if err := data.restore(book); err != nil {
		return nil, fmt.Errorf("некорректные данные в %q: %w", s.path, err)
	}
	return book, nil
}

func (s *JSONStore) Update(change func(*domain.Gradebook) error) error {
	lockPath := s.path + ".lock"
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("ведомость занята другой записью: %s", lockPath)
	}
	if err != nil {
		return err
	}
	defer os.Remove(lockPath)
	if err := lock.Close(); err != nil {
		return err
	}

	book, err := s.Read()
	if err != nil {
		return err
	}
	if err := change(book); err != nil {
		return err
	}
	data, err := makeSnapshot(book)
	if err != nil {
		return err
	}
	return writeFile(s.path, func(w io.Writer) error {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)
	})
}

func (data snapshot) restore(book *domain.Gradebook) error {
	// Восстанавливаем ведомость через доменные методы, чтобы повторно проверить её правила.
	for _, student := range data.Students {
		if err := book.AddStudent(student.ID, student.Name); err != nil {
			return err
		}
	}
	for _, criterion := range data.Criteria {
		if err := book.AddCriterion(criterion.ID, criterion.Name, criterion.MaxScore, criterion.Weight); err != nil {
			return err
		}
	}
	for _, evaluation := range data.Evaluations {
		err := book.SetEvaluation(evaluation.StudentID, evaluation.CriterionID, evaluation.Points, evaluation.Comment)
		if err != nil {
			return err
		}
	}
	return nil
}

func makeSnapshot(book *domain.Gradebook) (snapshot, error) {
	data := snapshot{Students: book.Students(), Criteria: book.Criteria(), Evaluations: []savedEvaluation{}}
	for _, student := range data.Students {
		evaluations, err := book.Evaluations(student.ID)
		if err != nil {
			return snapshot{}, err
		}
		for _, evaluation := range evaluations {
			data.Evaluations = append(data.Evaluations, savedEvaluation{
				StudentID: student.ID, CriterionID: evaluation.CriterionID,
				Points: evaluation.Points(), Comment: evaluation.Comment,
			})
		}
	}
	return data, nil
}
