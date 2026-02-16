package customers

import (
	"fmt"
	"gsupert/internal/common"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
	"io"
)

type Service struct {
	repo         *Repository
	emailService *common.EmailService
}

func NewService(repo *Repository, emailService *common.EmailService) *Service {
	return &Service{repo: repo, emailService: emailService}
}

func (s *Service) SendGreetingEmail(id string) error {
	customer, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if customer.Email == "" {
		return fmt.Errorf("customer has no email address")
	}

	subject := "Chúc Mừng Năm Mới - GSuperT"
	body := fmt.Sprintf("Chào %s,\n\nChúc mừng năm mới! Chúc bạn một năm mới an khang thịnh vượng và hạnh phúc.\n\nTrân trọng,\nTeam GSuperT", customer.Name)

	return s.emailService.SendEmail(customer.Email, subject, body)
}

func (s *Service) ExportPDF(w io.Writer) error {
	customers, err := s.repo.List()
	if err != nil {
		return err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Customers List")
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(10, 10, "ID")
	pdf.Cell(40, 10, "Name")
	pdf.Cell(60, 10, "Email")
	pdf.Cell(40, 10, "Phone")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	for i, c := range customers {
		pdf.Cell(10, 10, fmt.Sprintf("%d", i+1))
		pdf.Cell(40, 10, c.Name)
		pdf.Cell(60, 10, c.Email)
		pdf.Cell(40, 10, c.Phone)
		pdf.Ln(10)
	}

	return pdf.Output(w)
}

func (s *Service) ExportExcel(w io.Writer) error {
	customers, err := s.repo.List()
	if err != nil {
		return err
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetName := "Customers"
	f.SetSheetName("Sheet1", sheetName)

	headers := []string{"ID", "Name", "Email", "Phone", "Address"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	for i, c := range customers {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), c.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), c.Name)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), c.Email)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), c.Phone)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), c.Address)
	}

	return f.Write(w)
}

func (s *Service) CreateCustomer(name, email, phone, address string) (*Customer, error) {
	customer := &Customer{
		Name:    name,
		Email:   email,
		Phone:   phone,
		Address: address,
	}

	if err := s.repo.Create(customer); err != nil {
		return nil, err
	}

	return customer, nil
}

func (s *Service) GetCustomer(id string) (*Customer, error) {
	return s.repo.FindByID(id)
}

func (s *Service) ListCustomers() ([]Customer, error) {
	return s.repo.List()
}

func (s *Service) UpdateCustomer(id string, name, email, phone, address string) (*Customer, error) {
	customer, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	customer.Name = name
	customer.Email = email
	customer.Phone = phone
	customer.Address = address

	if err := s.repo.Update(customer); err != nil {
		return nil, err
	}

	return customer, nil
}

func (s *Service) DeleteCustomer(id string) error {
	return s.repo.Delete(id)
}
