package service

import "fmt"

type PermissionMutationScope struct {
	ActorID  int64
	OrgID    int64
	Username string
}

func (s PermissionMutationScope) isZero() bool {
	return s.ActorID == 0 && s.OrgID == 0 && s.Username == ""
}

func resolvePermissionScope(scopes []PermissionMutationScope) PermissionMutationScope {
	if len(scopes) == 0 {
		return PermissionMutationScope{}
	}
	return scopes[0]
}

func requireOrgScope(scope PermissionMutationScope) error {
	if err := requireGovernedOrgContext(scope.OrgID); err != nil {
		return err
	}
	return nil
}

func requireDatasetOrgValidator() error {
	return fmt.Errorf("org-scoped dataset permission save is unsupported: current dataset model does not expose a safe org boundary")
}
