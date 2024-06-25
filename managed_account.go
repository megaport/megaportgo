package megaport

import (
	"context"
	"encoding/json"
	"io"
)

type ManagedAccountService interface {
	// ListManagedAccounts retrieves a list of managed accounts. Megaport Partners can use this command to list all the managed companies linked to their account.
	ListManagedAccounts(ctx context.Context) ([]*ManagedAccount, error)
	// CreateManagedAccount creates a new managed account. As a Megaport Partner, use this endpoint to create a new managed company.
	CreateManagedAccount(ctx context.Context, req *ManagedAccountRequest) (*ManagedAccount, error)
	// UpdateManagedAccount updates an existing managed account. As a Megaport Partner, use this endpoint to update an existing managed company. You identify the company by providing the companyUid as a parameter for the endpoint.
	UpdateManagedAccount(ctx context.Context, companyUID string, req *ManagedAccountRequest) (*ManagedAccount, error)
	// GetManagedAccount retrieves a managed account by name. As a Megaport Partner, use this endpoint to retrieve a managed company by name.
	GetManagedAccount(ctx context.Context, companyUID string, managedAccountName string) (*ManagedAccount, error)
}

type ManagedAccountServiceOp struct {
	Client *Client
}

type ManagedAccount struct {
	AccountRef  string `json:"accountRef"`
	AccountName string `json:"accountName"`
	CompanyUID  string `json:"companyUid"`
}

type ManagedAccountRequest struct {
	AccountName string `json:"accountName"` // A required string that specifies a unique, easily identifiable name for the account. The length can range from 1 to 128 characters.
	AccountRef  string `json:"accountRef"`  // A required string that specifies a reference ID for the managed account. The accountRef is typically an identifier used in partner systems (for example, CRM or billing). This value is shown on the invoices as the Managed Account Reference. The accountRef also identifies the account in email notifications. (The accountRef value maps to the Managed Account UID in the Portal interface.)
}

type ManagedAccountAPIResponse struct {
	Message string          `json:"message"`
	Terms   string          `json:"terms"`
	Data    *ManagedAccount `json:"data"`
}

type ManagedAccountListAPIResponse struct {
	Message string            `json:"message"`
	Terms   string            `json:"terms"`
	Data    []*ManagedAccount `json:"data"`
}

// NewManagedAccountService creates a new instance of the ManagedAccount Service.
func NewManagedAccountService(c *Client) *ManagedAccountServiceOp {
	return &ManagedAccountServiceOp{
		Client: c,
	}
}

func (svc *ManagedAccountServiceOp) ListManagedAccounts(ctx context.Context) ([]*ManagedAccount, error) {
	path := "/v2/managedCompanies"
	req, err := svc.Client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, req, nil)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, fileErr := io.ReadAll(response.Body)

	if fileErr != nil {
		return nil, fileErr
	}

	var apiResponse *ManagedAccountListAPIResponse

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

func (svc *ManagedAccountServiceOp) CreateManagedAccount(ctx context.Context, req *ManagedAccountRequest) (*ManagedAccount, error) {
	path := "/v2/managedCompanies"
	clientReq, err := svc.Client.NewRequest(ctx, "POST", path, req)
	if err != nil {
		return nil, err
	}
	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var createManagedAccountResponse *ManagedAccountAPIResponse
	if err := json.Unmarshal(body, &createManagedAccountResponse); err != nil {
		return nil, err
	}
	return createManagedAccountResponse.Data, nil
}

func (svc *ManagedAccountServiceOp) UpdateManagedAccount(ctx context.Context, companyUID string, req *ManagedAccountRequest) (*ManagedAccount, error) {
	path := "/v2/managedCompanies/" + companyUID
	clientReq, err := svc.Client.NewRequest(ctx, "PUT", path, req)
	if err != nil {
		return nil, err
	}
	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var updateManagedAccountResponse *ManagedAccountAPIResponse
	if err := json.Unmarshal(body, &updateManagedAccountResponse); err != nil {
		return nil, err
	}
	return updateManagedAccountResponse.Data, nil
}

func (svc *ManagedAccountServiceOp) GetManagedAccount(ctx context.Context, companyUID string, managedAccountName string) (*ManagedAccount, error) {
	accounts, err := svc.ListManagedAccounts(ctx)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		if account.AccountName == managedAccountName {
			return account, nil
		}
	}
	return nil, ErrManagedAccountNotFound
}
