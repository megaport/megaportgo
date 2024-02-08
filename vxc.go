package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"
)

const PARTNER_AZURE string = "AZURE"
const PARTNER_GOOGLE string = "GOOGLE"
const PARTNER_AWS string = "AWS"
const PARTNER_OCI string = "ORACLE"

type VXCService interface {
	BuyVXC(ctx context.Context, req *BuyVXCRequest) (*BuyVXCResponse, error)
	GetVXC(ctx context.Context, id string) (*VXC, error)
	DeleteVXC(ctx context.Context, id string, req *DeleteVXCRequest) error
	UpdateVXC(ctx context.Context, id string, req *UpdateVXCRequest) (*VXC, error)
}

func NewVXCService(c *Client) *VXCServiceOp {
	return &VXCServiceOp{
		Client: c,
	}
}

var _ VXCService = &VXCServiceOp{}

type VXCServiceOp struct {
	Client *Client
}

type BuyVXCRequest struct {
	PortUID           string
	VXCName           string
	RateLimit         int
	AEndConfiguration VXCOrderAEndConfiguration
	BEndConfiguration VXCOrderBEndConfiguration

	WaitForProvision bool          // Wait until the VXC provisions before returning
	WaitForTime      time.Duration // How long to wait for the VXC to provision if WaitForProvision is true (default is 5 minutes)
}

type BuyVXCResponse struct {
	TechnicalServiceUID string
}

type DeleteVXCRequest struct {
	DeleteNow bool
}

type DeleteVXCResponse struct {
	IsDeleting bool
}

type UpdateVXCRequest struct {
	AEndVLAN  *int
	BEndVlan  *int
	RateLimit *int
	Name      *string

	WaitForUpdate bool          // Wait until the VXC updates before returning
	WaitForTime   time.Duration // How long to wait for the VXC to update if WaitForUpdate is true (default is 5 minutes)
}

type UpdateVXCResponse struct {
}

func (svc *VXCServiceOp) BuyVXC(ctx context.Context, req *BuyVXCRequest) (*BuyVXCResponse, error) {
	buyOrder := []VXCOrder{{
		PortID: req.PortUID,
		AssociatedVXCs: []VXCOrderConfiguration{
			{
				Name:      req.VXCName,
				RateLimit: req.RateLimit,
				AEnd:      req.AEndConfiguration,
				BEnd:      req.BEndConfiguration,
			},
		},
	}}

	responseBody, responseError := svc.Client.ProductService.ExecuteOrder(ctx, buyOrder)
	if responseError != nil {
		return nil, responseError
	}

	orderInfo := VXCOrderResponse{}
	if err := json.Unmarshal(*responseBody, &orderInfo); err != nil {
		return nil, err
	}
	serviceUID := orderInfo.Data[0].TechnicalServiceUID

	// wait until the VXC is provisioned before returning if reqested by the user
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
				return nil, fmt.Errorf("time expired waiting for VXC %s to provision", serviceUID)
			case <-ctx.Done():
				return nil, fmt.Errorf("context expired waiting for VXC %s to provision", serviceUID)
			case <-ticker.C:
				vxcDetails, err := svc.GetVXC(ctx, serviceUID)
				if err != nil {
					return nil, err
				}

				if slices.Contains(SERVICE_STATE_READY, vxcDetails.ProvisioningStatus) {
					return &BuyVXCResponse{
						TechnicalServiceUID: serviceUID,
					}, nil
				}

			}
		}
	} else {
		// return the service UID right away if the user doesn't want to wait for provision
		return &BuyVXCResponse{
			TechnicalServiceUID: serviceUID,
		}, nil
	}
}

func (svc *VXCServiceOp) GetVXC(ctx context.Context, id string) (*VXC, error) {
	path := "/v2/product/" + id
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

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	vxcDetails := VXCResponse{}
	if err = json.Unmarshal(body, &vxcDetails); err != nil {
		return nil, err
	}

	return &vxcDetails.Data, nil
}

func (svc *VXCServiceOp) DeleteVXC(ctx context.Context, id string, req *DeleteVXCRequest) error {
	_, err := svc.Client.ProductService.DeleteProduct(ctx, &DeleteProductRequest{
		ProductID: id,
		DeleteNow: req.DeleteNow,
	})
	if err != nil {
		return err
	}
	return nil
}

func (svc *VXCServiceOp) UpdateVXC(ctx context.Context, id string, req *UpdateVXCRequest) (*VXC, error) {
	path := fmt.Sprintf("/v2/product/%s/%s", PRODUCT_VXC, id)
	url := svc.Client.BaseURL.JoinPath(path).String()

	update := &VXCUpdate{
		Name:      req.Name,
		RateLimit: req.RateLimit,
		AEndVLAN:  req.AEndVLAN,
		BEndVLAN:  req.BEndVlan,
	}

	clientReq, err := svc.Client.NewRequest(ctx, http.MethodPut, url, update)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// TODO: add waiting mechanics here

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	vxcDetails := VXCResponse{}
	if err = json.Unmarshal(body, &vxcDetails); err != nil {
		return nil, err
	}

	return &vxcDetails.Data, nil
}
