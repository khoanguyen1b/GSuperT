package customers

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
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
