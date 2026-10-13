package rdsrds

import (
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// defaultDatabasePort is used only when the instance's own port is not in state
// yet and no annotation overrides it. PostgreSQL's port, matching the engine the
// proxy is most often placed in front of.
const defaultDatabasePort int32 = 5432

func proxyInstanceLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/rds/dbProxy::aws.rds.dbProxy.<targetInstance>.port": {
			Name:  "aws.rds.dbProxy.<targetInstance>.port",
			Label: "Database Port",
			Type:  core.ScalarTypeInteger,
			Description: "The port the proxy connects to the DB instance on. Defaults to the " +
				"instance's own port, falling back to 5432 when that is not yet known.",
			Examples: []*core.ScalarValue{
				core.ScalarFromInt(5432),
				core.ScalarFromInt(3306),
			},
			Required: false,
		},
	}
}

// Resolves the port the security group rule should open, preferring an
// explicit annotation, then the instance's own port from state.
func databasePort(
	proxyInfo *provider.ResourceInfo,
	instanceInfo *provider.ResourceInfo,
) int32 {
	if instanceInfo != nil {
		port, has := pluginutils.GetIntAnnotation(
			proxyInfo,
			&pluginutils.AnnotationQuery[int]{
				Key: fmt.Sprintf(
					"aws.rds.dbProxy.%s.port",
					pluginutils.GetResourceName(instanceInfo),
				),
			},
		)
		if has && port > 0 {
			return int32(port)
		}
	}

	if instancePort, has := instanceStatePort(instanceInfo); has {
		return instancePort
	}

	return defaultDatabasePort
}

// Reads the port from the instance's current state, accepting either
// the top-level port or the one carried on its endpoint.
func instanceStatePort(instanceInfo *provider.ResourceInfo) (int32, bool) {
	if instanceInfo == nil {
		return 0, false
	}
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(instanceInfo)
	for _, path := range []string{"$.port", "$.endpoint.port"} {
		node, has := pluginutils.GetValueByPath(path, spec)
		if !has || node == nil {
			continue
		}
		if value := core.IntValue(node); value > 0 {
			return int32(value), true
		}
	}
	return 0, false
}
