package application

import (
	"fmt"

	"pos/internal/domain"
)

type Repository interface {
	Read() (*domain.Gradebook, error)
	Update(func(*domain.Gradebook) error) error
}

type ReportExporter interface {
	Export(path string, report Report) error
}

type ResultRow struct {
	Student domain.Student
	Result  domain.Result
}

type Report struct {
	Mode string
	Rows []ResultRow
	Mean float64
	Min  float64
	Max  float64
}

// Service предоставляет консоли сценарии работы с ведомостью.
type Service struct {
	repo     Repository
	exporter ReportExporter
}

func NewService(repo Repository, exporter ReportExporter) *Service {
	return &Service{repo: repo, exporter: exporter}
}

func (s *Service) AddStudent(id, name string) error {
	return s.repo.Update(func(book *domain.Gradebook) error {
		return book.AddStudent(id, name)
	})
}

func (s *Service) AddCriterion(id, name string, maxScore, weight float64) error {
	return s.repo.Update(func(book *domain.Gradebook) error {
		return book.AddCriterion(id, name, maxScore, weight)
	})
}

func (s *Service) SetEvaluation(studentID, criterionID string, points float64, comment string) error {
	return s.repo.Update(func(book *domain.Gradebook) error {
		return book.SetEvaluation(studentID, criterionID, points, comment)
	})
}

func (s *Service) CancelEvaluation(studentID, criterionID string) error {
	return s.repo.Update(func(book *domain.Gradebook) error {
		return book.CancelEvaluation(studentID, criterionID)
	})
}

func (s *Service) Students() ([]domain.Student, error) {
	book, err := s.repo.Read()
	if err != nil {
		return nil, err
	}
	return book.Students(), nil
}

func (s *Service) Criteria() ([]domain.Criterion, error) {
	book, err := s.repo.Read()
	if err != nil {
		return nil, err
	}
	return book.Criteria(), nil
}

func (s *Service) Evaluations(studentID string) ([]domain.Evaluation, error) {
	book, err := s.repo.Read()
	if err != nil {
		return nil, err
	}
	return book.Evaluations(studentID)
}

func resultStrategy(mode string) (domain.ResultStrategy, error) {
	switch mode {
	case "final":
		return domain.FinalStrategy{}, nil
	case "progress":
		return domain.ProgressStrategy{}, nil
	default:
		return nil, fmt.Errorf("режим %q неизвестен, используйте final или progress", mode)
	}
}

func (s *Service) Result(studentID, mode string) (domain.Result, error) {
	strategy, err := resultStrategy(mode)
	if err != nil {
		return domain.Result{}, err
	}
	book, err := s.repo.Read()
	if err != nil {
		return domain.Result{}, err
	}
	return domain.CalculateResult(book, studentID, strategy)
}

func (s *Service) Report(mode string) (Report, error) {
	strategy, err := resultStrategy(mode)
	if err != nil {
		return Report{}, err
	}
	book, err := s.repo.Read()
	if err != nil {
		return Report{}, err
	}

	report := Report{Mode: mode}
	for _, student := range book.Students() {
		result, err := domain.CalculateResult(book, student.ID, strategy)
		if err != nil {
			return Report{}, err
		}
		if len(report.Rows) == 0 {
			report.Min, report.Max = result.Percent, result.Percent
		}
		report.Rows = append(report.Rows, ResultRow{Student: student, Result: result})
		report.Mean += (result.Percent - report.Mean) / float64(len(report.Rows))
		report.Min = min(report.Min, result.Percent)
		report.Max = max(report.Max, result.Percent)
	}
	return report, nil
}

func (s *Service) Export(path, mode string) error {
	report, err := s.Report(mode)
	if err != nil {
		return err
	}
	return s.exporter.Export(path, report)
}
