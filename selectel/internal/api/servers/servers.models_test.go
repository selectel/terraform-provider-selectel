package servers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServers_FindOneByName(t *testing.T) {
	svrs := Servers{
		&Server{Name: "server1"},
		&Server{Name: "server2"},
	}
	tests := []struct {
		name string
		arg  string
		want *Server
	}{
		{"FoundServer1", "server1", &Server{Name: "server1"}},
		{"NotFound", "server3", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare

			// Execute
			got := svrs.FindOneByName(tt.arg)

			// Analyse
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServer_IsLocationAvailable(t *testing.T) {
	server := Server{
		Available: []*ServerAvailable{
			{LocationID: "loc1"},
			{LocationID: "loc2"},
		},
	}
	tests := []struct {
		name       string
		locationID string
		want       bool
	}{
		{"Available", "loc1", true},
		{"NotAvailable", "loc3", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			got := server.IsLocationAvailable(tt.locationID)

			// Analyse
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServer_IsPricePlanAvailableForLocation(t *testing.T) {
	server := Server{
		PricePlanAvailable: []string{"plan1", "plan2"},
		Available: []*ServerAvailable{
			{
				LocationID: "loc1",
				PlanCount: []*ServerAvailablePricePlan{
					{Count: 1, PlanUUID: "plan1"},
					{Count: 0, PlanUUID: "plan2"},
				},
			},
		},
	}
	tests := []struct {
		name        string
		pricePlanID string
		locationID  string
		want        bool
	}{
		{"PlanAvailableInLocation", "plan1", "loc1", true},
		{"PlanNotInPricePlanAvailable", "plan3", "loc1", false},
		{"LocationNotAvailable", "plan1", "loc2", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			got := server.IsPricePlanAvailableForLocation(tt.pricePlanID, tt.locationID)

			// Analyse
			require.Equal(t, tt.want, got)
		})
	}
}

func TestResourceDetails_IsServerChip(t *testing.T) {
	tests := []struct {
		name        string
		serviceType string
		want        bool
	}{
		{"IsServerChip", "serverchip", true},
		{"IsNotServerChip", "server", false},
		{"EmptyServiceType", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare
			rd := ResourceDetails{ServiceType: tt.serviceType}

			// Execute
			got := rd.IsServerChip()

			// Analyse
			require.Equal(t, tt.want, got)
		})
	}
}

func TestResourceDetails_IsServer(t *testing.T) {
	tests := []struct {
		name        string
		serviceType string
		want        bool
	}{
		{"IsServer", "server", true},
		{"IsNotServer", "serverchip", false},
		{"EmptyServiceType", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare
			rd := ResourceDetails{ServiceType: tt.serviceType}

			// Execute
			got := rd.IsServer()

			// Analyse
			require.Equal(t, tt.want, got)
		})
	}
}

func TestServer_IsPrivateNetworkAvailable(t *testing.T) {
	tests := []struct {
		name         string
		isServerChip bool
		tags         []string
		want         bool
	}{
		{"PrivateNetworkAvailable", false, []string{"tag1", "tag2"}, true},
		{"PrivateNetworkNotAvailable_ServerChip", true, []string{"tag1", "tag2"}, false},
		{"PrivateNetworkNotAvailable_TagRestricted", false, []string{"10GE_Internet_MC-LAG"}, false},
		{"PrivateNetworkAvailable_EmptyTags", false, []string{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare
			server := Server{
				IsServerChip: tt.isServerChip,
				Tags:         tt.tags,
			}

			// Execute
			got := server.IsPrivateNetworkAvailable()

			// Analyse
			require.Equal(t, tt.want, got)
		})
	}
}
