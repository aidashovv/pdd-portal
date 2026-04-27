package users

import (
	"fmt"
	"strings"

	coreerrors "pdd-service/internal/core/errors"
)

type Role int16

const (
	RoleUser Role = iota
	RoleModerator
	RoleAdmin
)

func (r Role) String() string {
	switch r {
	case RoleUser:
		return "USER"
	case RoleModerator:
		return "MODERATOR"
	case RoleAdmin:
		return "ADMIN"
	default:
		return "UNKNOWN"
	}
}

func (r Role) IsValid() bool {
	return r == RoleUser || r == RoleModerator || r == RoleAdmin
}

func ParseRole(raw string) (Role, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "USER":
		return RoleUser, nil
	case "MODERATOR":
		return RoleModerator, nil
	case "ADMIN":
		return RoleAdmin, nil
	default:
		return RoleUser, fmt.Errorf("parse role %q: %w", raw, coreerrors.ErrInvalidDomainValue)
	}
}

func (r Role) IsAdmin() bool {
	return r == RoleAdmin
}

func (r Role) IsModerator() bool {
	return r == RoleModerator
}

func (r Role) CanModerateReports() bool {
	return r == RoleModerator || r == RoleAdmin
}

func (r Role) CanManageViolationTypes() bool {
	return r == RoleAdmin
}

func (r Role) CanCreateReport() bool {
	return r == RoleUser
}
