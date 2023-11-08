package models

import (
	typeid "go.jetpack.io/typeid/typed"

	"github.com/nikoksr/dbench/ent"
)

type (
	// Result represents the result of a benchmark run.
	Result = ent.Result

	resultGroupPrefix struct{}
	ResultGroupID     struct {
		typeid.TypeID[resultGroupPrefix]
	}
)

func (resultGroupPrefix) Type() string { return "resgrp" }
