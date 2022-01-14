package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/policy-server/cc_client"
	. "github.com/onsi/gomega"
)

type ConfigurableMockCCServer struct {
	server *httptest.Server

	apps           map[string]struct{}
	spaces         map[string]struct{}
	securityGroups map[string]cc_client.SecurityGroupResource
}

type resource struct {
	Guid string `json:"guid"`
}

func NewConfigurableMockCCServer() *ConfigurableMockCCServer {
	c := &ConfigurableMockCCServer{
		apps:           make(map[string]struct{}),
		spaces:         make(map[string]struct{}),
		securityGroups: make(map[string]cc_client.SecurityGroupResource),
	}
	c.server = httptest.NewUnstartedServer(c)

	return c
}

func (c *ConfigurableMockCCServer) Start() {
	c.server.Start()
}

func (c *ConfigurableMockCCServer) Close() {
	c.server.Close()
}

func (c *ConfigurableMockCCServer) URL() string {
	return c.server.URL
}

func (c *ConfigurableMockCCServer) AddApp(guid string) {
	c.apps[guid] = struct{}{}
}

func (c *ConfigurableMockCCServer) AddSpace(guid string) {
	c.spaces[guid] = struct{}{}
}

func (c *ConfigurableMockCCServer) AddSecurityGroup(securityGroup cc_client.SecurityGroupResource) {
	c.securityGroups[securityGroup.GUID] = securityGroup
}

func (c *ConfigurableMockCCServer) DeleteApp(guid string) {
	delete(c.apps, guid)
}

func (c *ConfigurableMockCCServer) DeleteSpace(guid string) {
	delete(c.spaces, guid)
}

func (c *ConfigurableMockCCServer) DeleteSecurityGroup(guid string) {
	delete(c.securityGroups, guid)
}

func (c *ConfigurableMockCCServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header["Authorization"][0] != "bearer valid-token" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if r.URL.Path == "/v3/apps" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(buildCCGuidsResponse(c.apps)))
		return
	}

	if r.URL.Path == "/v3/spaces" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(buildCCGuidsResponse(c.spaces)))
		return
	}

	if r.URL.Path == "/v3/security_groups" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(buildCCSecurityGroupsResponse(c.securityGroups)))
		return
	}

	w.WriteHeader(http.StatusTeapot)
	return
}

func buildCCGuidsResponse(guids map[string]struct{}) string {
	var resources []interface{}

	for guid, _ := range guids {
		resources = append(resources, resource{Guid: guid})
	}

	return buildCCResponse(resources)
}

func buildCCSecurityGroupsResponse(securityGroups map[string]cc_client.SecurityGroupResource) string {
	var resources []interface{}

	for _, securityGroup := range securityGroups {
		resources = append(resources, securityGroup)
	}

	return buildCCResponse(resources)
}

func buildCCResponse(resources []interface{}) string {
	resourceJSON, err := json.Marshal(resources)
	Expect(err).NotTo(HaveOccurred())

	return fmt.Sprintf(`{
		"pagination": {
			"total_results": %d,
			"total_pages": 1
		},
		"resources": %s
	}`, len(resources), string(resourceJSON))
}
