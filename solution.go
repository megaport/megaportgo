package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"

	"github.com/google/uuid"
)

// SolutionService is an interface for interfacing with the Solution endpoints of the Megaport API.
// A Solution is a named group of services, such as a DCI, that is ordered and billed as one item.
type SolutionService interface {
	// BuyDCI buys a DCI: a Port at each end and a VXC between them, placed as one order.
	BuyDCI(ctx context.Context, req *BuyDCIRequest) (*BuyDCIResponse, error)
	// GetSolution gets a Solution and its member services.
	GetSolution(ctx context.Context, solutionUID string) (*Solution, error)
	// GetSolutionBillable gets a Solution's current term and auto-renew setting.
	GetSolutionBillable(ctx context.Context, solutionUID string) (*SolutionBillable, error)
	// UpdateSolution changes a Solution's name and service level reference.
	UpdateSolution(ctx context.Context, req *UpdateSolutionRequest) error
}

// NewSolutionService creates a new instance of the Solution Service.
func NewSolutionService(c *Client) *SolutionServiceOp {
	return &SolutionServiceOp{
		Client: c,
	}
}

var _ SolutionService = &SolutionServiceOp{}

// SolutionServiceOp handles communication with the Solution related methods of the Megaport API.
type SolutionServiceOp struct {
	Client *Client
}

// BuyDCI buys a DCI using POST /v4/networkdesign/buy. The order holds both Ports, the VXC between them and the
// DCI itself, so the DCI is created, priced and billed as one item. The API checks the name, term, speed and locations.
func (svc *SolutionServiceOp) BuyDCI(ctx context.Context, req *BuyDCIRequest) (*BuyDCIResponse, error) {
	if req == nil {
		return nil, ErrBuyDCIRequestNil
	}

	solutionRef := uuid.NewString()
	networkDesign := createDCIOrder(req, solutionRef)
	solutions := []solutionOrder{{
		Ref:                   solutionRef,
		Type:                  SolutionTypeDCI,
		Name:                  req.Name,
		ServiceLevelReference: req.ServiceLevelReference,
	}}

	responseBody, err := executeOrder(ctx, svc.Client, networkDesign, solutions)
	if err != nil {
		return nil, err
	}
	orderInfo := solutionOrderResponse{}
	if err = json.Unmarshal(*responseBody, &orderInfo); err != nil {
		return nil, err
	}

	toReturn := &BuyDCIResponse{}
	for _, s := range orderInfo.Solutions {
		if s.Ref == solutionRef {
			toReturn.SolutionUID = s.UID
			break
		}
	}
	if toReturn.SolutionUID == "" {
		return nil, ErrDCIOrderResponseMissingSolution
	}

	// wait until the DCI is provisioned before returning if requested by the user
	if req.WaitForProvision {
		what := fmt.Sprintf("DCI %s to provision", toReturn.SolutionUID)
		err := waitUntilReady(ctx, req.WaitForTime, what, func() (bool, error) {
			solution, err := svc.GetSolution(ctx, toReturn.SolutionUID)
			if err != nil {
				return false, err
			}

			// the status is rolled up from the members, so it is ready only once every member is
			return slices.Contains(SERVICE_STATE_READY, solution.ProvisioningStatus), nil
		})
		if err != nil {
			return toReturn, err
		}
	}

	// return the Solution UID right away if the user doesn't want to wait for provision
	return toReturn, nil
}

// createDCIOrder builds the network design of a DCI: the A-End Port carrying the VXC, then the B-End Port.
func createDCIOrder(req *BuyDCIRequest, solutionRef string) []dciPortOrder {
	// The Ports do not exist yet, so the VXC names its ends by temporary UIDs.
	aEndPortUID, bEndPortUID, vxcUID := uuid.NewString(), uuid.NewString(), uuid.NewString()

	port := func(uid string, locationID int, diversityZone string) dciPortOrder {
		return dciPortOrder{
			ProductUID:         uid,
			Name:               req.Name,
			ProductType:        "MEGAPORT",
			ProvisioningStatus: STATUS_DESIGN,
			LocationID:         locationID,
			PortSpeed:          req.PortSpeed,
			Term:               req.Term,
			AutoRenewTerm:      req.AutoRenewTerm,
			SolutionRef:        solutionRef,
			Config:             PortOrderConfig{DiversityZone: diversityZone},
			AssociatedVXCs:     []dciVXCOrder{},
		}
	}

	aEndPort := port(aEndPortUID, req.AEndLocationID, req.AEndDiversityZone)
	aEndPort.AssociatedVXCs = []dciVXCOrder{{
		ProductUID:         vxcUID,
		Name:               req.Name,
		ProductType:        "VXC",
		ProvisioningStatus: STATUS_DESIGN,
		RateLimit:          req.PortSpeed,
		Term:               req.Term,
		AutoRenewTerm:      req.AutoRenewTerm,
		ConnectType:        "DEFAULT",
		AEnd:               dciVXCOrderEnd{ProductUID: aEndPortUID},
		BEnd:               dciVXCOrderEnd{ProductUID: bEndPortUID},
		SolutionRef:        solutionRef,
	}}
	bEndPort := port(bEndPortUID, req.BEndLocationID, req.BEndDiversityZone)

	return []dciPortOrder{aEndPort, bEndPort}
}

// GetSolution gets a Solution and its member services using GET /v1/solutions/{uid}.
func (svc *SolutionServiceOp) GetSolution(ctx context.Context, solutionUID string) (*Solution, error) {
	if solutionUID == "" {
		return nil, ErrSolutionUIDRequired
	}

	req, err := svc.Client.NewRequest(ctx, http.MethodGet, svc.solutionURL(solutionUID), nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, req, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	resp := solutionAPIResponse{}
	if err = json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode GetSolution response: %w", err)
	}
	if resp.Data == nil || resp.Data.UID == "" {
		return nil, ErrSolutionResponseEmpty
	}

	return resp.Data, nil
}

// GetSolutionBillable gets a Solution's current term using GET /v1/solutions/{uid}/billable.
// Before billing starts, the term and auto-renew setting are as ordered and the term dates are nil.
func (svc *SolutionServiceOp) GetSolutionBillable(ctx context.Context, solutionUID string) (*SolutionBillable, error) {
	if solutionUID == "" {
		return nil, ErrSolutionUIDRequired
	}

	req, err := svc.Client.NewRequest(ctx, http.MethodGet, svc.solutionURL(solutionUID, "billable"), nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, req, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	resp := solutionBillableAPIResponse{}
	if err = json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode GetSolutionBillable response: %w", err)
	}
	if resp.Data == nil {
		return nil, ErrSolutionResponseEmpty
	}

	return resp.Data, nil
}

// UpdateSolution changes a Solution's name and service level reference using PUT /v1/solutions/{uid}.
// The API sets both on every update, so send the current value of the one that stays as it is.
// The API checks the name, and refuses the update while the Solution is locked. Read the updated Solution with GetSolution.
func (svc *SolutionServiceOp) UpdateSolution(ctx context.Context, req *UpdateSolutionRequest) error {
	if req == nil {
		return ErrUpdateSolutionRequestNil
	}
	if req.SolutionUID == "" {
		return ErrSolutionUIDRequired
	}

	clientReq, err := svc.Client.NewRequest(ctx, http.MethodPut, svc.solutionURL(req.SolutionUID), req)
	if err != nil {
		return err
	}

	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return nil
}

// solutionURL returns the URL of a Solution, or of one of its sub-resources.
func (svc *SolutionServiceOp) solutionURL(solutionUID string, elem ...string) string {
	path := append([]string{"/v1/solutions", url.PathEscape(solutionUID)}, elem...)
	return svc.Client.BaseURL.JoinPath(path...).String()
}
