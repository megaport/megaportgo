package megaport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"
)

// PortService is an interface for interfacing with the Port endpoints
// of the Megaport API.

type PortService interface {
	BuyPort(ctx context.Context, req *BuyPortRequest) (*BuyPortResponse, error)
	BuySinglePort(ctx context.Context, req *BuySinglePortRequest) (*BuyPortResponse, error)
	BuyLAGPort(ctx context.Context, req *BuyLAGPortRequest) (*BuyPortResponse, error)
	ListPorts(ctx context.Context) ([]*Port, error)
	GetPort(ctx context.Context, portId string) (*Port, error)
	ModifyPort(ctx context.Context, req *ModifyPortRequest) (*ModifyPortResponse, error)
	DeletePort(ctx context.Context, req *DeletePortRequest) (*DeletePortResponse, error)
	RestorePort(ctx context.Context, portId string) (*RestorePortResponse, error)
	LockPort(ctx context.Context, portId string) (*LockPortResponse, error)
	UnlockPort(ctx context.Context, portId string) (*UnlockPortResponse, error)
}

func NewPortService(c *Client) *PortServiceOp {
	return &PortServiceOp{
		Client: c,
	}
}

// PortServiceOp handles communication with Port methods of the Megaport API.
type PortServiceOp struct {
	Client *Client
}

type BuyPortRequest struct {
	Name          string `json:"name"`
	Term          int    `json:"term"`
	PortSpeed     int    `json:"portSpeed"`
	LocationId    int    `json:"locationId"`
	Market        string `json:"market"`
	IsLag         bool   `json:"isLag"`
	LagCount      int    `json:"lagCount"`
	IsPrivate     bool   `json:"isPrivate"`
	DiversityZone string `json:"diversityZone"`

	WaitForProvision bool          // Wait until the VXC provisions before returning
	WaitForTime      time.Duration // How long to wait for the VXC to provision if WaitForProvision is true (default is 5 minutes)
}

type BuySinglePortRequest struct {
	Name          string
	Term          int
	PortSpeed     int
	LocationId    int
	Market        string
	IsPrivate     bool
	DiversityZone string

	WaitForProvision bool          // Wait until the VXC provisions before returning
	WaitForTime      time.Duration // How long to wait for the VXC to provision if WaitForProvision is true (default is 5 minutes)
}

type BuyLAGPortRequest struct {
	Name          string
	Term          int
	PortSpeed     int
	LocationId    int
	Market        string
	LagCount      int
	IsPrivate     bool
	DiversityZone string

	WaitForProvision bool          // Wait until the Port provisions before returning
	WaitForTime      time.Duration // How long to wait for the Port to provision if WaitForProvision is true (default is 5 minutes)
}

type BuyPortResponse struct {
	TechnicalServiceUID string
}

type GetPortRequest struct {
	PortID string
}

type ModifyPortRequest struct {
	PortID                string
	Name                  string
	MarketplaceVisibility bool
	CostCentre            string

	WaitForUpdate bool          // Wait until the Port updates before returning
	WaitForTime   time.Duration // How long to wait for the Port to update if WaitForUpdate is true (default is 5 minutes)
}

type ModifyPortResponse struct {
	IsUpdated bool
}

type DeletePortRequest struct {
	PortID    string
	DeleteNow bool
}

type DeletePortResponse struct {
	IsDeleting bool
}

type RestorePortRequest struct {
	PortID string
}

type RestorePortResponse struct {
	IsRestored bool
}

type LockPortRequest struct {
	PortID string
}

type LockPortResponse struct {
	IsLocking bool
}

type UnlockPortRequest struct {
	PortID string
}

type UnlockPortResponse struct {
	IsUnlocking bool
}

func (svc *PortServiceOp) BuyPort(ctx context.Context, req *BuyPortRequest) (*BuyPortResponse, error) {
	var buyOrder []PortOrder
	if req.Term != 1 && req.Term != 12 && req.Term != 24 && req.Term != 36 {
		return nil, errors.New(ERR_TERM_NOT_VALID)
	}
	if req.IsLag {
		buyOrder = []PortOrder{
			{
				Name:                  req.Name,
				Term:                  req.Term,
				ProductType:           "MEGAPORT",
				PortSpeed:             req.PortSpeed,
				LocationID:            req.LocationId,
				DiversityZone: 	       req.DiversityZone,
				CreateDate:            GetCurrentTimestamp(),
				Virtual:               false,
				Market:                req.Market,
				LagPortCount:          req.LagCount,
				MarketplaceVisibility: !req.IsPrivate,
			},
		}
	} else {
		buyOrder = []PortOrder{
			{
				Name:                  req.Name,
				Term:                  req.Term,
				ProductType:           "MEGAPORT",
				PortSpeed:             req.PortSpeed,
				LocationID:            req.LocationId,
				DiversityZone:         req.DiversityZone,
				CreateDate:            GetCurrentTimestamp(),
				Virtual:               false,
				Market:                req.Market,
				MarketplaceVisibility: !req.IsPrivate,
			},
		}
	}

	responseBody, responseError := svc.Client.ProductService.ExecuteOrder(ctx, buyOrder)
	if responseError != nil {
		return nil, responseError
	}
	orderInfo := PortOrderResponse{}
	unmarshalErr := json.Unmarshal(*responseBody, &orderInfo)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	toReturn := &BuyPortResponse{
		TechnicalServiceUID: orderInfo.Data[0].TechnicalServiceUID,	
	}

	// wait until the Port is provisioned before returning if requested by the user
	if req.WaitForProvision {
		toWait := req.WaitForTime
		if toWait == 0 {
			toWait = 5 * time.Minute
		}

		ticker := time.NewTicker(30 * time.Second) // check on the provision status every 30 seconds
		timer := time.NewTimer(toWait)
		defer ticker.Stop()
		defer timer.Stop()

		for {
			select {
			case <-timer.C:
				return nil, fmt.Errorf("time expired waiting for Port %s to provision", toReturn.TechnicalServiceUID)
			case <-ctx.Done():
				return nil, fmt.Errorf("context expired waiting for Port %s to provision", toReturn.TechnicalServiceUID)
			case <-ticker.C:
				portDetails, err := svc.GetPort(ctx, toReturn.TechnicalServiceUID)
				if err != nil {
					return nil, err
				}

				if slices.Contains(SERVICE_STATE_READY, portDetails.ProvisioningStatus) {
					return toReturn, nil
				}

			}
		}
	} else {
		// return the service UID right away if the user doesn't want to wait for provision
		return toReturn, nil
	}
}

func (svc *PortServiceOp) BuySinglePort(ctx context.Context, req *BuySinglePortRequest) (*BuyPortResponse, error) {
	return svc.BuyPort(ctx, &BuyPortRequest{
		Name:             req.Name,
		Term:             req.Term,
		PortSpeed:        req.PortSpeed,
		LocationId:       req.LocationId,
		Market:           req.Market,
		IsLag:            false,
		LagCount:         0,
		IsPrivate:        req.IsPrivate,
		DiversityZone:    req.DiversityZone,
		WaitForProvision: req.WaitForProvision,
		WaitForTime:      req.WaitForTime,
	})
}

func (svc *PortServiceOp) BuyLAGPort(ctx context.Context, req *BuyLAGPortRequest) (*BuyPortResponse, error) {
	return svc.BuyPort(ctx, &BuyPortRequest{
		Name:             req.Name,
		Term:             req.Term,
		PortSpeed:        req.PortSpeed,
		LocationId:       req.LocationId,
		Market:           req.Market,
		IsLag:            true,
		LagCount:         req.LagCount,
		IsPrivate:        req.IsPrivate,
		DiversityZone:    req.DiversityZone,
		WaitForProvision: req.WaitForProvision,
		WaitForTime:      req.WaitForTime,
	})
}

func (svc *PortServiceOp) ListPorts(ctx context.Context) ([]*Port, error) {
	path := "/v2/products"
	url := svc.Client.BaseURL.JoinPath(path).String()
	req, err := svc.Client.NewRequest(ctx, http.MethodGet, url, nil)
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

	parsed := ParsedProductsResponse{}

	unmarshalErr := json.Unmarshal(body, &parsed)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	ports := []*Port{}

	for _, unmarshaledData := range parsed.Data {
		// The products query response will likely contain non-port objects.  As a result
		// we need to initially Unmarshal as ParsedProductsResponse so that we may iterate
		// over the entries in Data then re-Marshal those entries so that we may Unmarshal
		// them as Port (and `continue` where that doesn't work).  We could write a custom
		// deserializer to avoid this but that is a lot of work for a performance
		// optimization which is likely irrelevant in practice.
		// Unfortunately I know of no better (maintainable) method of making this work.
		remarshaled, err := json.Marshal(unmarshaledData)
		if err != nil {
			svc.Client.Logger.WarnContext(ctx, fmt.Sprintf("Could not remarshal %v as port.", err.Error()))
			continue
		}
		port := Port{}
		unmarshalErr = json.Unmarshal(remarshaled, &port)
		if unmarshalErr != nil {
			svc.Client.Logger.WarnContext(ctx, fmt.Sprintf("Could not unmarshal %v as port.", unmarshalErr.Error()))
			continue
		}
		ports = append(ports, &port)
	}
	return ports, nil
}

func (svc *PortServiceOp) GetPort(ctx context.Context, portId string) (*Port, error) {
	path := "/v2/product/" + portId
	url := svc.Client.BaseURL.JoinPath(path).String()

	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, fileErr := io.ReadAll(response.Body)

	if fileErr != nil {
		return nil, fileErr
	}

	portDetails := PortResponse{}
	unmarshalErr := json.Unmarshal(body, &portDetails)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return &portDetails.Data, nil
}

func (svc *PortServiceOp) ModifyPort(ctx context.Context, req *ModifyPortRequest) (*ModifyPortResponse, error) {
	modifyRes, err := svc.Client.ProductService.ModifyProduct(ctx, &ModifyProductRequest{
		ProductID:             req.PortID,
		ProductType:           PRODUCT_MEGAPORT,
		Name:                  req.Name,
		CostCentre:            req.CostCentre,
		MarketplaceVisibility: req.MarketplaceVisibility,
	})
	if err != nil {
		return nil, err
	}
	toReturn := &ModifyPortResponse{
		IsUpdated: modifyRes.IsUpdated,
	}

	// wait until the Port is updated before returning if requested by the user
	if req.WaitForUpdate {
		toWait := req.WaitForTime
		if toWait == 0 {
			toWait = 5 * time.Minute
		}

		ticker := time.NewTicker(30 * time.Second) // check on the update status every 30 seconds
		timer := time.NewTimer(toWait)
		defer ticker.Stop()
		defer timer.Stop()

		for {
			select {
			case <-timer.C:
				return nil, fmt.Errorf("time expired waiting for Port %s to update", req.PortID)
			case <-ctx.Done():
				return nil, fmt.Errorf("context expired waiting for Port %s to update",req.PortID)
			case <-ticker.C:
				portDetails, err := svc.GetPort(ctx, req.PortID)
				if err != nil {
					return nil, err
				}
				if portDetails.Name == req.Name && portDetails.CostCentre == req.CostCentre && portDetails.MarketplaceVisibility == req.MarketplaceVisibility && portDetails.ProvisioningStatus == "LIVE" {
					return toReturn, nil
				}
			}
		}
	} else {
		// return the response right away if the user doesn't want to wait for update
		return toReturn, nil
	}
}

func (svc *PortServiceOp) DeletePort(ctx context.Context, req *DeletePortRequest) (*DeletePortResponse, error) {
	_, err := svc.Client.ProductService.DeleteProduct(ctx, &DeleteProductRequest{
		ProductID: req.PortID,
		DeleteNow: req.DeleteNow,
	})
	if err != nil {
		return nil, err
	}
	return &DeletePortResponse{
		IsDeleting: true,
	}, nil
}

func (svc *PortServiceOp) RestorePort(ctx context.Context, portId string) (*RestorePortResponse, error) {
	_, err := svc.Client.ProductService.RestoreProduct(ctx, portId)
	if err != nil {
		return nil, err
	}
	return &RestorePortResponse{
		IsRestored: true,
	}, nil
}

func (svc *PortServiceOp) LockPort(ctx context.Context, portId string) (*LockPortResponse, error) {
	port, err := svc.GetPort(ctx, portId)
	if err != nil {
		return nil, err
	}
	if !port.Locked {
		_, err = svc.Client.ProductService.ManageProductLock(ctx, &ManageProductLockRequest{
			ProductID:  portId,
			ShouldLock: true,
		})
		if err != nil {
			return nil, err
		}
		return &LockPortResponse{IsLocking: true}, nil
	} else {
		return nil, errors.New(ERR_PORT_ALREADY_LOCKED)
	}
}

func (svc *PortServiceOp) UnlockPort(ctx context.Context, portId string) (*UnlockPortResponse, error) {
	port, err := svc.GetPort(ctx, portId)
	if err != nil {
		return nil, err
	}
	if port.Locked {
		_, err = svc.Client.ProductService.ManageProductLock(ctx, &ManageProductLockRequest{
			ProductID:  portId,
			ShouldLock: false,
		})
		if err != nil {
			return nil, err
		}
		return &UnlockPortResponse{IsUnlocking: true}, nil
	} else {
		return nil, errors.New(ERR_PORT_NOT_LOCKED)
	}
}