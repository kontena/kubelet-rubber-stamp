package controller

import (
	"github.com/jnummelin/csr-approver/pkg/controller/certificatesigningrequest"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, certificatesigningrequest.Add)
}
