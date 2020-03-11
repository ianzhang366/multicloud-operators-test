package controller

import (
	"github.ibm.com/steve-kim-ibm/multicloud-operators-test/pkg/controller/apptest"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, apptest.Add)
}
