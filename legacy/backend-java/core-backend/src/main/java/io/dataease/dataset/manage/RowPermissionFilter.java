package io.dataease.dataset.manage;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import io.dataease.auth.bo.TokenUserBO;
import io.dataease.system.dao.auto.mapper.DataPermRowMapper;
import io.dataease.system.dao.auto.mapper.DataPermColumnMapper;
import io.dataease.system.entity.DataPermRow;
import io.dataease.system.entity.DataPermColumn;
import io.dataease.engine.utils.SQLUtils;
import io.dataease.utils.AuthUtils;
import org.apache.commons.lang3.ObjectUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

@Component
public class RowPermissionFilter {

    private static final ObjectMapper objectMapper = new ObjectMapper();

    @Autowired
    private DataPermRowMapper dataPermRowMapper;

    @Autowired
    private DataPermColumnMapper dataPermColumnMapper;

    public String buildWhereFilter(Long datasetId) {
        return buildWhereFilter(datasetId, null);
    }

    public String buildSelectColumns(Long datasetId) {
        if (datasetId == null) {
            return "*";
        }

        TokenUserBO user = AuthUtils.getUser();
        if (user == null) {
            return "*";
        }

        if (AuthUtils.isSysAdmin(user.getUserId())) {
            return "*";
        }

        List<DataPermRow> rowPermissions = dataPermRowMapper.listByDatasetId(datasetId);
        List<DataPermColumn> columnPermissions = dataPermColumnMapper.listByDatasetId(datasetId);

        if (ObjectUtils.isEmpty(rowPermissions) && ObjectUtils.isEmpty(columnPermissions)) {
            return "*";
        }

        Set<String> excludedColumns = new HashSet<>();
        Map<String, String> maskedColumns = new HashMap<>();

        for (DataPermColumn colPerm : columnPermissions) {
            if (colPerm.getStatus() != 1) {
                continue;
            }

            String fieldName = colPerm.getFieldName();
            if ("disable".equalsIgnoreCase(colPerm.getPermType())) {
                excludedColumns.add(fieldName);
            } else if ("mask".equalsIgnoreCase(colPerm.getPermType())) {
                maskedColumns.put(fieldName, colPerm.getMaskRule());
            }
        }

        if (excludedColumns.isEmpty() && maskedColumns.isEmpty()) {
            return "*";
        }

        return buildSelectClause(datasetId, excludedColumns, maskedColumns);
    }

    public String buildWhereFilter(Long datasetId, Set<String> selectColumns) {
        if (datasetId == null) {
            return null;
        }

        TokenUserBO user = AuthUtils.getUser();
        if (user == null) {
            return null;
        }

        if (AuthUtils.isSysAdmin(user.getUserId())) {
            return null;
        }

        List<DataPermRow> permissions = dataPermRowMapper.listByDatasetId(datasetId);
        if (ObjectUtils.isEmpty(permissions)) {
            return null;
        }

        List<String> whereConditions = new ArrayList<>();
        for (DataPermRow perm : permissions) {
            if (perm.getStatus() != 1) {
                continue;
            }

            boolean matches = false;
            if ("user".equals(perm.getAuthTargetType())) {
                matches = perm.getAuthTargetId().equals(user.getUserId());
            }

            if (matches) {
                String condition = parseExpressionTree(perm.getExpressionTree(), selectColumns);
                if (ObjectUtils.isNotEmpty(condition)) {
                    whereConditions.add(condition);
                }
            }
        }

        if (whereConditions.isEmpty()) {
            return null;
        }

        return "(" + String.join(" OR ", whereConditions) + ")";
    }

    private String parseExpressionTree(String expressionTreeJson, Set<String> selectColumns) {
        try {
            JsonNode root = objectMapper.readTree(expressionTreeJson);
            String logic = root.has("logic") ? root.get("logic").asText() : "OR";
            JsonNode items = root.get("items");
            if (items == null || !items.isArray() || items.size() == 0) {
                return "";
            }

            List<String> conditions = new ArrayList<>();
            for (JsonNode item : items) {
                if (item.has("subTree")) {
                    String subCondition = parseSubTree(item.get("subTree"), selectColumns);
                    if (ObjectUtils.isNotEmpty(subCondition)) {
                        conditions.add("(" + subCondition + ")");
                    }
                } else {
                    String fieldId = item.has("fieldId") ? item.get("fieldId").asText() : "";
                    String filterType = item.has("filterType") ? item.get("filterType").asText() : "";
                    String term = item.has("term") ? item.get("term").asText() : "";
                    String value = item.has("value") ? item.get("value").asText() : "";

                    if (ObjectUtils.isNotEmpty(fieldId) && ObjectUtils.isNotEmpty(filterType) && ObjectUtils.isNotEmpty(value)) {
                        if (selectColumns != null && selectColumns.contains(fieldId)) {
                            String condition = buildFieldCondition(fieldId, filterType, term, value);
                            if (ObjectUtils.isNotEmpty(condition)) {
                                conditions.add(condition);
                            }
                        }
                    }
                }
            }

            if (conditions.isEmpty()) {
                return "";
            }

            String operator = "OR".equalsIgnoreCase(logic) ? " OR " : " AND ";
            return "(" + String.join(operator, conditions) + ")";
        } catch (Exception e) {
            return "";
        }
    }

    private String parseSubTree(JsonNode subTree, Set<String> selectColumns) {
        String logic = subTree.has("logic") ? subTree.get("logic").asText() : "OR";
        JsonNode items = subTree.get("items");
        if (items == null || !items.isArray()) {
            return "";
        }

        List<String> conditions = new ArrayList<>();
        for (JsonNode item : items) {
            if (item.has("subTree")) {
                String subCondition = parseSubTree(item.get("subTree"), selectColumns);
                if (ObjectUtils.isNotEmpty(subCondition)) {
                    conditions.add("(" + subCondition + ")");
                }
            } else {
                String fieldId = item.has("fieldId") ? item.get("fieldId").asText() : "";
                String filterType = item.has("filterType") ? item.get("filterType").asText() : "";
                String term = item.has("term") ? item.get("term").asText() : "";
                String value = item.has("value") ? item.get("value").asText() : "";
                String enumValue = item.has("enumValue") ? item.get("enumValue").asText() : "";

                if (ObjectUtils.isNotEmpty(fieldId) && ObjectUtils.isNotEmpty(filterType)) {
                    String condition;
                    if ("enum".equals(filterType) && ObjectUtils.isNotEmpty(enumValue)) {
                        condition = buildEnumCondition(fieldId, enumValue);
                    } else {
                        condition = buildFieldCondition(fieldId, filterType, term, value);
                    }
                    if (ObjectUtils.isNotEmpty(condition)) {
                        if (selectColumns == null || selectColumns.contains(fieldId)) {
                            conditions.add(condition);
                        }
                    }
                }
            }
        }

        if (conditions.isEmpty()) {
            return "";
        }

        String operator = "OR".equalsIgnoreCase(logic) ? " OR " : " AND ";
        return "(" + String.join(operator, conditions) + ")";
    }

    private String buildFieldCondition(String fieldId, String filterType, String term, String value) {
        if (ObjectUtils.isEmpty(value)) {
            return "";
        }

        String field = "`" + fieldId + "`";
        switch (term) {
            case "eq":
                return field + " = '" + escapeSql(value) + "'";
            case "not eq":
                return field + " != '" + escapeSql(value) + "'";
            case "like":
                return field + " LIKE '%" + escapeSql(value) + "%'";
            case "not like":
                return field + " NOT LIKE '%" + escapeSql(value) + "%'";
            case "null":
                return field + " IS NULL";
            case "not null":
                return field + " IS NOT NULL";
            case "empty":
                return field + " = ''";
            case "not empty":
                return field + " != ''";
            case "gt":
                return field + " > " + escapeSql(value);
            case "lt":
                return field + " < " + escapeSql(value);
            case "ge":
                return field + " >= " + escapeSql(value);
            case "le":
                return field + " <= " + escapeSql(value);
            default:
                return field + " = '" + escapeSql(value) + "'";
        }
    }

    private String buildEnumCondition(String fieldId, String enumValue) {
        String[] values = enumValue.split(",");
        List<String> quotedValues = new ArrayList<>();
        for (String v : values) {
            v = v.trim();
            if (ObjectUtils.isNotEmpty(v)) {
                quotedValues.add("'" + escapeSql(v) + "'");
            }
        }
        return "`" + fieldId + "` IN (" + String.join(", ", quotedValues) + ")";
    }

    private String buildSelectClause(Long datasetId, Set<String> excludedColumns, Map<String, String> maskedColumns) {
        if (excludedColumns.isEmpty() && maskedColumns.isEmpty()) {
            return "*";
        }

        if (excludedColumns.isEmpty()) {
            return "*";
        }

        List<String> selectParts = new ArrayList<>();
        for (String maskedCol : maskedColumns.keySet()) {
            selectParts.add(buildMaskedColumnExpression(maskedCol));
        }

        if (!excludedColumns.isEmpty()) {
            List<String> allColumns = dataPermColumnMapper.listAllColumnNamesByDatasetId(datasetId);
            for (String col : allColumns) {
                if (!excludedColumns.contains(col) && !maskedColumns.containsKey(col)) {
                    selectParts.add("`" + col + "`");
                }
            }
        }

        if (selectParts.isEmpty()) {
            return "*";
        }

        return String.join(", ", selectParts);
    }

    private String buildMaskedColumnExpression(String columnName) {
        String column = "`" + columnName + "`";
        return "CASE " + column + " WHEN " + column + " IS NULL THEN NULL ELSE '******' END AS " + column;
    }

    private String escapeSql(String value) {
        return SQLUtils.transKeyword(value);
    }
}
