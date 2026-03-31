package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/repository"

	"go.uber.org/zap"
)

type RowPermissionService struct {
	rowPermRepo          *repository.RowPermissionRepository
	columnPermRepo       *repository.ColumnPermissionRepository
	userRoleRepo         UserRoleRepositoryInterface
	adminChecker         AdminCheckerInterface
	datasetFieldResolver DatasetFieldResolver
}

type UserRoleRepositoryInterface interface {
	GetRoleIDsByUserID(userID int64) ([]int64, error)
}

type AdminCheckerInterface interface {
	IsAdmin(userID int64) bool
}

type DatasetFieldResolver interface {
	GetFieldByID(id int64) (*dataset.CoreDatasetTableField, error)
}

func NewRowPermissionService(
	rowPermRepo *repository.RowPermissionRepository,
	columnPermRepo *repository.ColumnPermissionRepository,
	userRoleRepo UserRoleRepositoryInterface,
	adminChecker AdminCheckerInterface,
) *RowPermissionService {
	return &RowPermissionService{
		rowPermRepo:    rowPermRepo,
		columnPermRepo: columnPermRepo,
		userRoleRepo:   userRoleRepo,
		adminChecker:   adminChecker,
	}
}

func (s *RowPermissionService) SetDatasetFieldResolver(resolver DatasetFieldResolver) {
	s.datasetFieldResolver = resolver
}

func (s *RowPermissionService) GetRowPermissionsTree(datasetID, userID int64) ([]*permission.DataPermRow, error) {
	if s.adminChecker != nil && s.adminChecker.IsAdmin(userID) {
		return nil, nil
	}

	userPerms, err := s.rowPermRepo.ListByDatasetIDAndUserID(datasetID, userID)
	if err != nil {
		logger.Error("Failed to get user row permissions", zap.Error(err))
		return nil, err
	}

	if s.userRoleRepo != nil {
		roleIDs, err := s.userRoleRepo.GetRoleIDsByUserID(userID)
		if err != nil {
			logger.Warn("Failed to get user role IDs", zap.Error(err))
		} else if len(roleIDs) > 0 {
			rolePerms, err := s.rowPermRepo.ListByDatasetIDAndRoleIDs(datasetID, roleIDs)
			if err != nil {
				logger.Warn("Failed to get role row permissions", zap.Error(err))
			} else {
				userPerms = append(userPerms, rolePerms...)
			}
		}
	}

	return userPerms, nil
}

func (s *RowPermissionService) BuildWhereClause(datasetID, userID int64) (*permission.WhereClauseResult, error) {
	if s.adminChecker != nil && s.adminChecker.IsAdmin(userID) {
		return nil, nil
	}

	perms, err := s.GetRowPermissionsTree(datasetID, userID)
	if err != nil {
		return nil, err
	}

	if len(perms) == 0 {
		return nil, nil
	}

	var conditions []string
	var args []interface{}

	for _, perm := range perms {
		if perm.Status != 1 || perm.ExpressionTree == "" {
			continue
		}

		condition, conditionArgs := s.parseExpressionTree(perm.ExpressionTree)
		if condition != "" {
			conditions = append(conditions, condition)
			args = append(args, conditionArgs...)
		}
	}

	if len(conditions) == 0 {
		return nil, nil
	}

	clause := "(" + strings.Join(conditions, " OR ") + ")"
	return &permission.WhereClauseResult{
		Clause: clause,
		Args:   args,
	}, nil
}

func (s *RowPermissionService) BuildSelectColumns(datasetID, userID int64) (string, error) {
	if s.adminChecker != nil && s.adminChecker.IsAdmin(userID) {
		return "*", nil
	}

	colPerms, err := s.columnPermRepo.ListByDatasetID(datasetID)
	if err != nil {
		return "*", nil
	}

	if len(colPerms) == 0 {
		return "*", nil
	}

	excludedColumns := make(map[string]bool)
	for _, perm := range colPerms {
		if perm.PermType == "disable" {
			excludedColumns[perm.FieldName] = true
		}
	}

	if len(excludedColumns) == 0 {
		return "*", nil
	}

	allColumns, err := s.columnPermRepo.ListAllColumnNamesByDatasetID(datasetID)
	if err != nil {
		return "*", nil
	}

	var selectParts []string
	for _, col := range allColumns {
		if !excludedColumns[col] {
			selectParts = append(selectParts, fmt.Sprintf("`%s`", col))
		}
	}

	if len(selectParts) == 0 {
		return "*", nil
	}

	return strings.Join(selectParts, ", "), nil
}

func (s *RowPermissionService) parseExpressionTree(exprTree string) (string, []interface{}) {
	var obj permission.DatasetRowPermissionsTreeObj
	if err := json.Unmarshal([]byte(exprTree), &obj); err != nil {
		logger.Warn("Failed to parse expression tree", zap.Error(err))
		return "", nil
	}

	return s.parseTreeObj(&obj)
}

func (s *RowPermissionService) parseTreeObj(obj *permission.DatasetRowPermissionsTreeObj) (string, []interface{}) {
	if obj == nil || len(obj.Items) == 0 {
		return "", nil
	}

	logic := strings.ToUpper(obj.Logic)
	if logic == "" {
		logic = "OR"
	}

	var conditions []string
	var args []interface{}

	for _, item := range obj.Items {
		if item.SubTree != nil {
			subCondition, subArgs := s.parseTreeObj(item.SubTree)
			if subCondition != "" {
				conditions = append(conditions, "("+subCondition+")")
				args = append(args, subArgs...)
			}
		} else {
			condition, conditionArgs := s.buildItemCondition(&item)
			if condition != "" {
				conditions = append(conditions, condition)
				args = append(args, conditionArgs...)
			}
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}

	operator := " OR "
	if logic == "AND" {
		operator = " AND "
	}

	return "(" + strings.Join(conditions, operator) + ")", args
}

func (s *RowPermissionService) buildItemCondition(item *permission.DatasetRowPermissionsTreeItem) (string, []interface{}) {
	if item.FieldID == 0 {
		return "", nil
	}

	field := s.resolveFieldReference(item.FieldID)

	if item.FilterType == "enum" && len(item.EnumValue) > 0 {
		return s.buildEnumCondition(field, item.EnumValue)
	}

	return s.buildLogicCondition(field, item.Term, item.Value)
}

func (s *RowPermissionService) resolveFieldReference(fieldID int64) string {
	if s.datasetFieldResolver == nil {
		return fmt.Sprintf("`%d`", fieldID)
	}
	field, err := s.datasetFieldResolver.GetFieldByID(fieldID)
	if err != nil || field == nil {
		return fmt.Sprintf("`%d`", fieldID)
	}
	if field.OriginName != nil && strings.TrimSpace(*field.OriginName) != "" {
		return fmt.Sprintf("`%s`", strings.TrimSpace(*field.OriginName))
	}
	if field.Name != nil && strings.TrimSpace(*field.Name) != "" {
		return fmt.Sprintf("`%s`", strings.TrimSpace(*field.Name))
	}
	return fmt.Sprintf("`%d`", fieldID)
}

//nolint:gocyclo
func (s *RowPermissionService) buildLogicCondition(field, term, value string) (string, []interface{}) {
	if value == "" && term != permission.OperatorNull && term != permission.OperatorNotNull &&
		term != permission.OperatorEmpty && term != permission.OperatorNotEmpty {
		return "", nil
	}

	escapedValue := s.escapeSQL(value)

	switch term {
	case permission.OperatorEq:
		return fmt.Sprintf("%s = ?", field), []interface{}{escapedValue}
	case "not_eq", "not eq":
		return fmt.Sprintf("%s != ?", field), []interface{}{escapedValue}
	case permission.OperatorLike:
		return fmt.Sprintf("%s LIKE ?", field), []interface{}{"%" + escapedValue + "%"}
	case permission.OperatorNotLike:
		return fmt.Sprintf("%s NOT LIKE ?", field), []interface{}{"%" + escapedValue + "%"}
	case permission.OperatorNull:
		return fmt.Sprintf("%s IS NULL", field), nil
	case permission.OperatorNotNull:
		return fmt.Sprintf("%s IS NOT NULL", field), nil
	case permission.OperatorEmpty:
		return fmt.Sprintf("%s = ''", field), nil
	case permission.OperatorNotEmpty:
		return fmt.Sprintf("%s != ''", field), nil
	case permission.OperatorGt:
		return fmt.Sprintf("%s > ?", field), []interface{}{escapedValue}
	case permission.OperatorLt:
		return fmt.Sprintf("%s < ?", field), []interface{}{escapedValue}
	case permission.OperatorGe:
		return fmt.Sprintf("%s >= ?", field), []interface{}{escapedValue}
	case permission.OperatorLe:
		return fmt.Sprintf("%s <= ?", field), []interface{}{escapedValue}
	case permission.OperatorIn:
		return "", nil
	case permission.OperatorNotIn:
		return "", nil
	default:
		return fmt.Sprintf("%s = ?", field), []interface{}{escapedValue}
	}
}

func (s *RowPermissionService) buildEnumCondition(field string, values []string) (string, []interface{}) {
	if len(values) == 0 {
		return "", nil
	}

	placeholders := make([]string, len(values))
	args := make([]interface{}, len(values))
	for i, v := range values {
		placeholders[i] = "?"
		args[i] = s.escapeSQL(v)
	}

	return fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ", ")), args
}

func (s *RowPermissionService) escapeSQL(value string) string {
	replacer := strings.NewReplacer(
		"'", "''",
		"\\", "\\\\",
		"\x00", "",
		"\n", "",
		"\r", "",
		"\x1a", "",
	)
	return replacer.Replace(value)
}

func (s *RowPermissionService) IsAdmin(userID int64) bool {
	if s.adminChecker == nil {
		return false
	}
	return s.adminChecker.IsAdmin(userID)
}

func (s *RowPermissionService) GetUserRoleIDs(userID int64) ([]int64, error) {
	if s.userRoleRepo == nil {
		return nil, nil
	}
	return s.userRoleRepo.GetRoleIDsByUserID(userID)
}
