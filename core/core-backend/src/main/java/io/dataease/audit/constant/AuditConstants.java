package io.dataease.audit.constant;

public class AuditConstants {

    public static final String ACTION_TYPE_USER_ACTION = "USER_ACTION";
    public static final String ACTION_TYPE_PERMISSION_CHANGE = "PERMISSION_CHANGE";
    public static final String ACTION_TYPE_DATA_ACCESS = "DATA_ACCESS";
    public static final String ACTION_TYPE_SYSTEM_CONFIG = "SYSTEM_CONFIG";

    public static final String OPERATION_CREATE = "CREATE";
    public static final String OPERATION_UPDATE = "UPDATE";
    public static final String OPERATION_DELETE = "DELETE";
    public static final String OPERATION_EXPORT = "EXPORT";
    public static final String OPERATION_LOGIN = "LOGIN";
    public static final String OPERATION_LOGOUT = "LOGOUT";

    public static final String RESOURCE_TYPE_USER = "USER";
    public static final String RESOURCE_TYPE_ORGANIZATION = "ORGANIZATION";
    public static final String RESOURCE_TYPE_ROLE = "ROLE";
    public static final String RESOURCE_TYPE_PERMISSION = "PERMISSION";
    public static final String RESOURCE_TYPE_DATASET = "DATASET";
    public static final String RESOURCE_TYPE_DASHBOARD = "DASHBOARD";

    public static final String STATUS_SUCCESS = "SUCCESS";
    public static final String STATUS_FAILED = "FAILED";

    public static final int DEFAULT_RETENTION_DAYS = 90;
}
