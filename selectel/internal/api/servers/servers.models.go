package servers

import (
	"net"
	"slices"
)

type (
	Server struct {
		ID                 string             `json:"uuid"`
		Name               string             `json:"name"`
		Available          []*ServerAvailable `json:"available"`
		PricePlanAvailable []string           `json:"price_plan_available"`
		Tags               []string           `json:"tags"`

		IsServerChip bool
	}

	ServerAvailable struct {
		LocationID string                      `json:"location"`
		PlanCount  []*ServerAvailablePricePlan `json:"plan_count"`
	}

	ServerAvailablePricePlan struct {
		Count    int    `json:"count"`
		PlanUUID string `json:"plan_uuid"`
	}

	Servers []*Server
)

func (s Server) IsLocationAvailable(locationID string) bool {
	for _, available := range s.Available {
		if available.LocationID == locationID {
			return true
		}
	}

	return false
}

func (s Server) IsPrivateNetworkAvailable() bool {
	return !s.IsServerChip &&
		!slices.ContainsFunc(s.Tags, func(s string) bool {
			return s == "10GE_Internet_MC-LAG" || s == "10GE_Internet" || s == "25GE_Local_MC-LAG" || s == "10GE_Local" ||
				s == "10GE_Local_MC-LAG"
		})
}

func (s Server) IsPricePlanAvailableForLocation(pricePlanID, locationID string) bool {
	var foundAvailable bool
	for _, psAvailable := range s.PricePlanAvailable {
		if psAvailable == pricePlanID {
			foundAvailable = true
			break
		}
	}

	if !foundAvailable {
		return false
	}

	for _, avByLoc := range s.Available {
		if avByLoc.LocationID == locationID {
			for _, planCount := range avByLoc.PlanCount {
				if planCount.PlanUUID == pricePlanID && planCount.Count >= 1 {
					return true
				}
			}
		}
	}

	return false
}

func (s Servers) FindOneByName(name string) *Server {
	for _, server := range s {
		if server.Name == name {
			return server
		}
	}

	return nil
}

type ServiceBilling struct {
	Currency         string     `json:"currency"`
	CurrentPricePlan *PricePlan `json:"current_price_plan,omitempty"`
	HasEnoughBalance bool       `json:"has_enough_balance"`
}

const (
	ServiceBillingPayCurrencyBonus = "bonus"
	ServiceBillingPayCurrencyMain  = "main"
)

type (
	ServerBillingPostPayload struct {
		ServiceUUID      string           `json:"service_uuid"`
		PricePlanUUID    string           `json:"price_plan_uuid"`
		PayCurrency      string           `json:"pay_currency"`
		LocationUUID     string           `json:"location_uuid"`
		Quantity         int              `json:"quantity,omitempty"`
		IPList           []net.IP         `json:"ip_list,omitempty"`
		LocalSubnetUUID  string           `json:"local_subnet_uuid,omitempty"`
		LocalIPList      []net.IP         `json:"local_ip_list,omitempty"`
		ProjectUUID      string           `json:"project_uuid"`
		PartitionsConfig PartitionsConfig `json:"partitions_config,omitempty"`
		OSVersion        string           `json:"version"`
		OSTemplate       string           `json:"os_template"`
		OSArch           string           `json:"arch"`
		UserSSHKey       string           `json:"user_ssh_key,omitempty"`
		UserHostname     string           `json:"userhostname"`
		UserDesc         string           `json:"user_desc"`
		Password         string           `json:"password,omitempty"`
		UserData         string           `json:"cloud_init_user_data,omitempty"`
	}

	ServerBillingPostResult struct {
		UUID string `json:"uuid"`
	}
)

type ResourceDetails struct {
	UUID         string          `json:"uuid"`
	State        string          `json:"state"`
	LocationUUID string          `json:"location_uuid"`
	ServiceUUID  string          `json:"service_uuid"`
	Billing      *ServiceBilling `json:"billing"`
	ServiceType  string          `json:"service_type"`
}

func (rd ResourceDetails) IsServerChip() bool {
	return rd.ServiceType == "serverchip"
}

func (rd ResourceDetails) IsServer() bool {
	return rd.ServiceType == "server"
}

const (
	ResourceDetailsStateActive     = "active"
	ResourceDetailsStatePending    = "pending"
	ResourceDetailsStateProcessing = "processing"
	ResourceDetailsStatePaid       = "paid"
	ResourceDetailsStateDeploy     = "deploy"
	ResourceDetailsStateExpiring   = "expiring"
	ResourceDetailsStateEnding     = "ending"
)
