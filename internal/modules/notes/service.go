package notes

import (
	"fmt"
	"gsupert/internal/common"
)

type Service struct {
	repo         *Repository
	emailService *common.EmailService
}

func NewService(repo *Repository, emailService *common.EmailService) *Service {
	return &Service{repo: repo, emailService: emailService}
}

func (s *Service) CreateNote(content, customerID string) (*Note, error) {
	note := &Note{
		Content:    content,
		CustomerID: customerID,
	}

	if err := s.repo.Create(note); err != nil {
		return nil, err
	}

	// Trigger email notification (fire and forget or handle error)
	// For simplicity, we just log the error if any
	go func() {
		subject := "New Note Created"
		body := fmt.Sprintf("A new note has been created for customer ID: %s\n\nContent: %s", customerID, content)
		// Sending to a dummy admin email or configurable admin email
		err := s.emailService.SendEmail("admin@example.com", subject, body)
		if err != nil {
			fmt.Printf("Failed to send email: %v\n", err)
		}
	}()

	return note, nil
}

func (s *Service) GetNote(id string) (*Note, error) {
	return s.repo.FindByID(id)
}

func (s *Service) ListNotes() ([]Note, error) {
	return s.repo.List()
}

func (s *Service) ListNotesByCustomer(customerID string) ([]Note, error) {
	return s.repo.ListByCustomerID(customerID)
}

func (s *Service) UpdateNote(id string, content string) (*Note, error) {
	note, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	note.Content = content

	if err := s.repo.Update(note); err != nil {
		return nil, err
	}

	return note, nil
}

func (s *Service) DeleteNote(id string) error {
	return s.repo.Delete(id)
}
