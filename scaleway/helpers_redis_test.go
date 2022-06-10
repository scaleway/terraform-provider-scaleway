package scaleway

import (
	"testing"

	redis "github.com/scaleway/scaleway-sdk-go/api/redis/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/stretchr/testify/assert"
)

func TestOrderPrivateNetworksInterfaceByFirstIP(t *testing.T) {
	log := logger{}
	pnList := []map[string]interface{}(nil)
	pnList = append(pnList, map[string]interface{}{
		"service_ips": []interface{}{"172.168.1.0/20", "10.10.1.0/12"},
		"id":          "4",
	})
	pnList = append(pnList, map[string]interface{}{
		"id":          "2",
		"service_ips": []interface{}{"10.12.1.1/20"},
	})
	pnList = append(pnList, map[string]interface{}{
		"service_ips": []interface{}{"192.168.1.0/20", "34.12.1.0/25"},
		"id":          "5",
	})
	pnList = append(pnList, map[string]interface{}{
		"id":          "3",
		"service_ips": []interface{}{"10.13.1.0/20", "6.10.12.0/20", "123.23.45.6/21"},
	})
	pnList = append(pnList, map[string]interface{}{
		"id":          "1",
		"service_ips": []interface{}{"10.12.1.0/20", "15.10.12.0/20", "1.23.45.6/21"},
	})
	pnOrdered := orderPrivateNetworksInterfaceByFirstIP(pnList)
	resp := pnOrdered.([]map[string]interface{})
	for _, pn := range resp {
		log.Debugf("%s: %s", pn["id"], pn["service_ips"])
	}
	assert.Equal(t, "1", resp[0]["id"])
	assert.Equal(t, "2", resp[1]["id"])
	assert.Equal(t, "3", resp[2]["id"])
	assert.Equal(t, "4", resp[3]["id"])
	assert.Equal(t, "5", resp[4]["id"])
}

func scwIPNetSliceFromStrings(rawIPs ...string) []scw.IPNet {
	ipSlice := []scw.IPNet(nil)
	for _, ip := range rawIPs {
		ipNet, _ := expandIPNet(ip)
		ipSlice = append(ipSlice, ipNet)
	}
	return ipSlice
}

func TestOrderPrivateNetworksSpecsByFirstIP(t *testing.T) {
	log := logger{}
	pnList := []*redis.EndpointSpec(nil)
	pnList = append(pnList, &redis.EndpointSpec{
		PrivateNetwork: &redis.EndpointSpecPrivateNetworkSpec{
			ID:         "4",
			ServiceIPs: scwIPNetSliceFromStrings("172.168.1.0/20", "10.10.1.0/12"),
		},
	})
	pnList = append(pnList, &redis.EndpointSpec{
		PrivateNetwork: &redis.EndpointSpecPrivateNetworkSpec{
			ID:         "2",
			ServiceIPs: scwIPNetSliceFromStrings("10.12.1.1/20"),
		},
	})
	pnList = append(pnList, &redis.EndpointSpec{
		PrivateNetwork: &redis.EndpointSpecPrivateNetworkSpec{
			ID:         "5",
			ServiceIPs: scwIPNetSliceFromStrings("192.168.1.0/20", "34.12.1.0/25"),
		},
	})
	pnList = append(pnList, &redis.EndpointSpec{
		PrivateNetwork: &redis.EndpointSpecPrivateNetworkSpec{
			ID:         "3",
			ServiceIPs: scwIPNetSliceFromStrings("10.13.1.0/20", "6.10.12.0/20", "123.23.45.6/21"),
		},
	})
	pnList = append(pnList, &redis.EndpointSpec{
		PrivateNetwork: &redis.EndpointSpecPrivateNetworkSpec{
			ID:         "1",
			ServiceIPs: scwIPNetSliceFromStrings("10.12.1.0/20", "15.10.12.0/20", "1.23.45.6/21"),
		},
	})
	pnOrdered := orderPrivateNetworksSpecsByFirstIP(pnList)
	for _, pn := range pnOrdered {
		log.Debugf("%s: %v", pn.PrivateNetwork.ID, pn.PrivateNetwork.ServiceIPs)
	}
	assert.Equal(t, "1", pnOrdered[0].PrivateNetwork.ID)
	assert.Equal(t, "2", pnOrdered[1].PrivateNetwork.ID)
	assert.Equal(t, "3", pnOrdered[2].PrivateNetwork.ID)
	assert.Equal(t, "4", pnOrdered[3].PrivateNetwork.ID)
	assert.Equal(t, "5", pnOrdered[4].PrivateNetwork.ID)
}
