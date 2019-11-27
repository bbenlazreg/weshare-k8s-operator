package controller

import (
	"github.com/kube_op/minicd-operator/pkg/controller/appset"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, appset.Add)
}
