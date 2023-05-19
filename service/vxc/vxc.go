// Copyright 2020 Megaport Pty Ltd
//
// Licensed under the Mozilla Public License, Version 2.0 (the
// "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//       https://mozilla.org/MPL/2.0/
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vxc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/service/product"
	"github.com/megaport/megaportgo/types"
)

type VXC struct {
	*config.Config
	product *product.Product
}

func New(cfg *config.Config) *VXC {
	return &VXC{
		Config:  cfg,
		product: product.New(cfg),
	}
}

func (v *VXC) BuyVXC(
	portUID string,
	vxcName string,
	rateLimit int,
	aEndConfiguration types.VXCOrderAEndConfiguration,
	bEndConfiguration types.VXCOrderBEndConfiguration,
) (string, error) {

	buyOrder := []types.VXCOrder{
		{
			PortID: portUID,
			AssociatedVXCs: []types.VXCOrderConfiguration{
				{
					Name:      vxcName,
					RateLimit: rateLimit,
					AEnd:      aEndConfiguration,
					BEnd:      bEndConfiguration,
				},
			},
		},
	}

	requestBody, _ := json.Marshal(buyOrder)

	responseBody, responseErr := v.product.ExecuteOrder(&requestBody)

	if responseErr != nil {
		return "", responseErr
	}

	orderInfo := types.VXCOrderResponse{}
	err := json.Unmarshal(*responseBody, &orderInfo)

	if err != nil {
		return "", err
	}

	return orderInfo.Data[0].TechnicalServiceUID, nil
}

// GetVXCDetails gets the details of a VXC.
func (v *VXC) GetVXCDetails(id string) (types.VXC, error) {
	url := "/v2/product/" + id
	response, err := v.Config.MakeAPICall("GET", url, nil)

	if err != nil {
		return types.VXC{}, err
	}

	defer ioutil.NopCloser(response.Body)

	body, fileErr := ioutil.ReadAll(response.Body)

	if fileErr != nil {
		return types.VXC{}, fileErr
	}

	vxcDetails := types.VXCResponse{}
	unmarshalErr := json.Unmarshal(body, &vxcDetails)

	if unmarshalErr != nil {
		return types.VXC{}, unmarshalErr
	}

	return vxcDetails.Data, nil
}

// DeleteVXC deletes a VXC.
func (v *VXC) DeleteVXC(id string, deleteNow bool) (bool, error) {
	return v.product.DeleteProduct(id, deleteNow)
}

// UpdateVXC updates a VXC
func (v *VXC) UpdateVXC(id string, name string, rateLimit int, aEndVLAN int, bEndVLAN int) (bool, error) {
	url := fmt.Sprintf("/v2/product/%s/%s", types.PRODUCT_VXC, id)
	var update interface{}

	if bEndVLAN == 0 {
		update = types.PartnerVXCUpdate{
			Name:      name,
			RateLimit: rateLimit,
			AEndVLAN:  aEndVLAN,
		}
	} else {
		update = types.VXCUpdate{
			Name:      name,
			RateLimit: rateLimit,
			AEndVLAN:  aEndVLAN,
			BEndVLAN:  &bEndVLAN,
		}
	}

	body, marshalErr := json.Marshal(update)

	if marshalErr != nil {
		return false, marshalErr
	}

	updateResponse, err := v.Config.MakeAPICall("PUT", url, []byte(body))
	isResErr, compiledResErr := v.Config.IsErrorResponse(updateResponse, &err, 200)

	if isResErr {
		return false, compiledResErr
	} else {
		return true, nil
	}
}

// WaitForVXCProvisioning waits up to 15 minutes for the VXC to reach the "LIVE" status.
// See WaitForVXCProvisioningCtx
func (v *VXC) WaitForVXCProvisioning(vxcId string) (bool, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancelFunc()
	return v.WaitForVXCProvisioningCtx(ctx, 30*time.Second, vxcId)
}

// WaitForVXCProvisioningCtx waits for the VXC to reach the "LIVE" status, retrying every [pollFrequency]
// seconds until the "LIVE" status is reached or the specified context expires or is canceled.
func (v *VXC) WaitForVXCProvisioningCtx(ctx context.Context, pollFrequency time.Duration, vxcId string) (bool, error) {
	vxcInfo, _ := v.GetVXCDetails(vxcId)
	if strings.Compare(vxcInfo.ProvisioningStatus, "LIVE") != 0 {
		return true, nil
	}

	// Go-Live
	v.Log.Info("Waiting for VXC status transition.")

	wait := 0
	timer := time.NewTicker(pollFrequency)
	defer timer.Stop()
	for {
		wait++

		select {

		case <-ctx.Done():
			vxcInfo, _ = v.GetVXCDetails(vxcId)
			if strings.Compare(vxcInfo.ProvisioningStatus, "LIVE") != 0 {
				return true, nil
			}

			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return false, errors.New(mega_err.ERR_VXC_PROVISION_TIMEOUT_EXCEED)
			} else {
				return false, ctx.Err()
			}

		case t, ok := <-timer.C:
			vxcInfo, _ = v.GetVXCDetails(vxcId)
			if strings.Compare(vxcInfo.ProvisioningStatus, "LIVE") != 0 {
				return true, nil
			}

			deadline, hasDeadline := ctx.Deadline()
			if !ok || (hasDeadline && t.After(deadline)) {
				v.Log.Debugln("VXC waiting cycle complete. Status: ", vxcInfo.ProvisioningStatus)
				return false, errors.New(mega_err.ERR_VXC_PROVISION_TIMEOUT_EXCEED)
			}

			if wait%5 == 0 {
				v.Log.Infoln("VXC is currently being provisioned. Status: ", vxcInfo.ProvisioningStatus)
			}
		}
	}
}

// WaitForVXCUpdated waits up to 15 minutes for the VXC update to get applied.
// See WaitForVXCUpdatedCtx
func (v *VXC) WaitForVXCUpdated(id string, name string, rateLimit int, aEndVLAN int, bEndVLAN int) (bool, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancelFunc()
	return v.WaitForVXCUpdatedCtx(ctx, 30*time.Second, id, name, rateLimit, aEndVLAN, bEndVLAN)
}

// WaitForVXCUpdatedCtx waits for the VXC update to be applied successfully, retrying every  [pollFrequency]
// seconds until the update completes or the specified context expires or is canceled.
func (v *VXC) WaitForVXCUpdatedCtx(
	ctx context.Context, pollFrequency time.Duration, id string, name string, rateLimit int, aEndVLAN int, bEndVLAN int,
) (bool, error) {
	wait := 0

	timer := time.NewTicker(pollFrequency)
	defer timer.Stop()
	for {
		wait++
		select {
		case <-ctx.Done():
			vxcDetails, _ := v.GetVXCDetails(id)
			if isUpdated(name, rateLimit, aEndVLAN, bEndVLAN, vxcDetails) {
				return true, nil
			}

			v.logUpdateStatus("VXC wait cycle complete", name, rateLimit, aEndVLAN, bEndVLAN, vxcDetails)
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return false, errors.New(mega_err.ERR_VXC_UPDATE_TIMEOUT_EXCEED)
			} else {
				return false, ctx.Err()
			}

		case t, ok := <-timer.C:
			vxcDetails, _ := v.GetVXCDetails(id)
			if isUpdated(name, rateLimit, aEndVLAN, bEndVLAN, vxcDetails) {
				return true, nil
			}

			deadline, hasDeadline := ctx.Deadline()
			if !ok || (hasDeadline && t.After(deadline)) {
				v.logUpdateStatus("VXC wait cycle complete", name, rateLimit, aEndVLAN, bEndVLAN, vxcDetails)
				return false, errors.New(mega_err.ERR_VXC_UPDATE_TIMEOUT_EXCEED)
			}

			if wait%5 == 0 {
				v.logUpdateStatus("VXC Update in progress", name, rateLimit, aEndVLAN, bEndVLAN, vxcDetails)
			}
		}
	}
}

func (v *VXC) logUpdateStatus(
	prefix string,
	name string,
	rateLimit int,
	aEndVLAN int,
	bEndVLAN int,
	vxcDetails types.VXC,
) {
	if aEndVLAN == 0 {
		aEndVLAN = vxcDetails.AEndConfiguration.VLAN
	}

	if bEndVLAN == 0 {
		bEndVLAN = vxcDetails.BEndConfiguration.VLAN
	}
	v.Log.Debugf("%s: Name %t; RateLimit %t; AEndVLAN %t; BEndVLAN %t\n",
		prefix,
		vxcDetails.Name == name,
		vxcDetails.RateLimit == rateLimit,
		vxcDetails.AEndConfiguration.VLAN == aEndVLAN,
		vxcDetails.BEndConfiguration.VLAN == bEndVLAN)
}

func isUpdated(
	name string,
	rateLimit int,
	aEndVLAN int,
	bEndVLAN int,
	vxcDetails types.VXC,
) bool {
	if aEndVLAN == 0 {
		aEndVLAN = vxcDetails.AEndConfiguration.VLAN
	}

	if bEndVLAN == 0 {
		bEndVLAN = vxcDetails.BEndConfiguration.VLAN
	}

	if vxcDetails.Name == name &&
		vxcDetails.RateLimit == rateLimit &&
		vxcDetails.AEndConfiguration.VLAN == aEndVLAN &&
		vxcDetails.BEndConfiguration.VLAN == bEndVLAN {
		return true
	}
	return false
}

func (v *VXC) UnmarshallMcrAEndConfig(vxcDetails types.VXC) (interface{}, error) {

	v.Log.Warn("Unmarshall")

	cspConnection := v.GetCspConnection("resource_name", "a_csp_connection", vxcDetails)

	if partner_interfaces, ok := cspConnection["interfaces"].([]interface{}); ok {

		v.Log.Warn("Interfaces")
		// handle more than one interface
		if len(partner_interfaces) != 1 {
			v.Log.Warn("More than one interface present in MCR A end Resource")
			return nil, errors.New("More than one interface present in MCR A end Resource")
		}

		for _, partner_interface := range partner_interfaces {

			v.Log.Warn("...processing")
			partner_configuration := map[string]interface{}{}

			partner_interface_map, pi_ok := partner_interface.(map[string]interface{})
			if !pi_ok {
				v.Log.Warn("Error casting partner_interface_map")
			}
			v.Log.Info(partner_interface_map)

			// add ip addresses to configuration
			if ip_slice, ip_ok := partner_interface_map["ipAddresses"].([]interface{}); ip_ok {
				if len(ip_slice) > 0 {
					v.Log.Info(" - ipAddresses field present")
					partner_configuration["ip_addresses"] = ip_slice
				} else {
					v.Log.Info(" - ipAddresses is empty")
				}
			} else {
				v.Log.Info(" - ipAddresses field not present")
			}

			// extract ip routes configurations
			ip_routes_list := []interface{}{}
			if ip_routes, ipr_ok := partner_interface_map["ipRoutes"].([]interface{}); ipr_ok {

				v.Log.Info(" - ip routes present")
				for _, ipRoute := range ip_routes {

					ip_route_map, iprm_ok := ipRoute.(map[string]interface{})
					if iprm_ok {

						new_ip_route := map[string]interface{}{
							"prefix":      ip_route_map["prefix"],
							"description": ip_route_map["description"],
							"next_hop":    ip_route_map["nextHop"],
						}

						ip_routes_list = append(ip_routes_list, new_ip_route)
					}

				} // end ip routes loop

				// add ip routes to configuration
				partner_configuration["ip_routes"] = ip_routes_list

			} else {
				v.Log.Info(" - ipRoutes field not present")
			} // end ip routes inspection

			// add nat ip addresses to configuration
			if nat_slice, nat_ok := partner_interface_map["natIpAddresses"].([]interface{}); nat_ok {
				if len(nat_slice) > 0 {
					v.Log.Info(" - natIpAddresses field present")
					partner_configuration["nat_ip_addresses"] = nat_slice
				} else {
					v.Log.Info(" - natIpAddresses is empty")
				}
			} else {
				v.Log.Info(" - natIpAddresses field not present")
			}

			// extract bfd settings
			bfd_map, bfd_ok := partner_interface_map["bfd"].(map[string]interface{})
			if bfd_ok {

				v.Log.Info(" - bfd field present")
				// add bfd to configuration
				partner_configuration["bfd_configuration"] = []interface{}{map[string]interface{}{
					"tx_interval": bfd_map["txInterval"],
					"rx_interval": bfd_map["rxInterval"],
					"multiplier":  bfd_map["multiplier"],
				}}

			} else {
				v.Log.Info(" - bfd field not present")
			}

			// extract bgp configurations
			bgp_connection_list := []interface{}{}
			if bgpConnections, bgp_ok := partner_interface_map["bgpConnections"].([]interface{}); bgp_ok {

				v.Log.Info(" - bgpConnections field present")
				for _, bgpConnection := range bgpConnections {

					bgp_connection_map, bgpm_ok := bgpConnection.(map[string]interface{})
					if bgpm_ok {

						new_bgp_connection := map[string]interface{}{
							"peer_asn":         bgp_connection_map["peerAsn"],
							"local_ip_address": bgp_connection_map["localIpAddress"],
							"peer_ip_address":  bgp_connection_map["peerIpAddress"],
							"password":         bgp_connection_map["password"],
							"shutdown":         bgp_connection_map["shutdown"],
							"description":      bgp_connection_map["description"],
							"med_in":           bgp_connection_map["medIn"],
							"med_out":          bgp_connection_map["medOut"],
							"bfd_enabled":      bgp_connection_map["bfdEnabled"],
							"export_policy":    bgp_connection_map["exportPolicy"],
							"permit_export_to": bgp_connection_map["permitExportTo"],
							"deny_export_to":   bgp_connection_map["denyExportTo"],
							"import_whitelist": bgp_connection_map["importWhitelist"],
							"import_blacklist": bgp_connection_map["importBlacklist"],
							"export_whitelist": bgp_connection_map["exportWhitelist"],
							"export_blacklist": bgp_connection_map["exportBlacklist"],
						}

						bgp_connection_list = append(bgp_connection_list, new_bgp_connection)
					}

				} // end bgp connections loop

				// add bgp to configuration
				partner_configuration["bgp_connection"] = bgp_connection_list

			} else {
				v.Log.Info(" - bgpConnections field not present")
			} // end bgp connection inspection

			if len(partner_configuration) > 0 {
				v.Log.Info("Package for return")
				wrapped_partner_configuration := append([]interface{}{}, partner_configuration)

				// Return here
				return wrapped_partner_configuration, nil
			}

		} // end interface loop

	}

	v.Log.Info("Nothing of value was found...")
	return nil, nil
}

func (v *VXC) GetCspConnection(cspIdentifier string, cspIdentifierValue string, vxcDetails types.VXC) map[string]interface{} {

	v.Log.Info("searching for  csp where " + cspIdentifier + "=" + cspIdentifierValue)
	cspConnectionList := []map[string]interface{}{}

	if cspConnectionListInner, ok := vxcDetails.Resources.CspConnection.([]interface{}); ok {
		for _, conn := range cspConnectionListInner {
			v.Log.Info("searchCspConnections - adding connection")
			cspConnection := conn.(map[string]interface{})
			cspConnectionList = append(cspConnectionList, cspConnection)
		}
	} else if cspConnection, ok := vxcDetails.Resources.CspConnection.(map[string]interface{}); ok {
		v.Log.Info("searchCspConnections - adding connection")
		cspConnectionList = append(cspConnectionList, cspConnection)
	}

	for _, conn := range cspConnectionList {
		v.Log.Info("inspecting - " + conn[cspIdentifier].(string))
		if cspIdentifierValue == conn[cspIdentifier].(string) {
			v.Log.Info("searchCspConnections - found")
			v.Log.Info(conn)
			return conn
		}
	}

	return nil
}

// GetPrefixFilterLists returns all Prefix Filter Lists on an MCR.
func (v *VXC) GetPrefixFilterLists(id string) ([]types.PrefixFilterList, error) {
	prefix, prefixErr := v.product.GetMCRPrefixFilterLists(id)
	return prefix, prefixErr
}
