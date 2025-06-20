package lambdalinks

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type FunctionCodeSigningConfigLinkUpdateSuite struct {
	suite.Suite
}

// type linkUpdateTestCase struct{}

func TestFunctionCodeSigningConfigLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionCodeSigningConfigLinkUpdateSuite))
}
