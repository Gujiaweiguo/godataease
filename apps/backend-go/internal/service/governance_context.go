package service

import "errors"

var ErrInvalidOrgContext = errors.New("invalid org context")

func requireGovernedOrgContext(orgID int64) error {
	if orgID <= 0 {
		return ErrInvalidOrgContext
	}
	return nil
}
