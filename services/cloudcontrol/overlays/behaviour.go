package overlays

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// SpecTransform mutates a spec mapping node in place. Used for the bespoke field
// handling some resources need that the generic engine cannot infer.
type SpecTransform func(spec *core.MappingNode) error

// NameGeneration auto-populates a name field with a unique provenance name when the
// user leaves it empty, reusing the shared name generators.
type NameGeneration struct {
	// Field is the camelCase spec field to populate (e.g. "queueName").
	Field string
	// Generate produces the unique name; it may read the resolved spec from the
	// input (e.g. to vary behaviour for FIFO queues).
	Generate utils.UniqueNameGenerator
}

// Behaviour carries per-type, hand-written behaviour the generic Cloud Control engine
// cannot derive from a schema: custom validation, unique-name generation, and
// transforms applied to desired/external state. It is the functionality-level equivalent to
// the schema and example overlays.
type Behaviour struct {
	CustomValidate func(
		context.Context,
		*provider.ResourceValidateInput,
	) (*provider.ResourceValidateOutput, error)
	Name                   *NameGeneration
	BeforeCreate           SpecTransform
	AfterReadExternalState SpecTransform
	// ValidateResolvedSpec runs against the fully resolved spec before both create
	// and update, for constraints on values that are only known once references are
	// resolved at deploy time (e.g. lists wired from another resource's computed
	// fields). Returning an error aborts the deployment with that message, giving
	// an actionable diagnostic in place of an opaque downstream AWS error.
	ValidateResolvedSpec SpecTransform
	// SettledAfterLinkUpdate answers whether the resource has settled following a write
	// made by a link rather than by the engine.
	//
	// The generic Cloud Control check cannot answer this as it polls the request status of
	// the operation the engine started, and a link writing through a service API starts no
	// such operation, so the check reports settled immediately. For a type whose service
	// API accepts a write and applies it asynchronously, that is the difference between
	// the next link succeeding and it being rejected.
	//
	// Nil means the type has nothing to wait for after a link's write, the common case.
	SettledAfterLinkUpdate func(
		ctx context.Context,
		input *provider.ResourceHasStabilisedInput,
		awsConfig *aws.Config,
	) (bool, error)
}

var behaviours = map[string]*Behaviour{}

// RegisterBehaviour associates a behaviour overlay with a Bluelink resource type,
// from a per-type init function. It panics on a duplicate registration.
func RegisterBehaviour(blueprintType string, behaviour *Behaviour) {
	if _, exists := behaviours[blueprintType]; exists {
		panic("duplicate behaviour overlay registered for " + blueprintType)
	}
	behaviours[blueprintType] = behaviour
}

// BehaviourFor returns the behaviour overlay for a type, or nil if none is registered.
func BehaviourFor(blueprintType string) *Behaviour {
	return behaviours[blueprintType]
}
