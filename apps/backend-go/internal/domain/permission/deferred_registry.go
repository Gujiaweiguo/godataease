package permission

import "fmt"

// DeferredDimension represents a permission dimension that is explicitly deferred.
type DeferredDimension struct {
	Name      string
	ErrorCode string
	Message   string
}

// DeferredDimensionRegistry tracks all deferred permission dimensions.
type DeferredDimensionRegistry struct {
	dimensions map[string]DeferredDimension
}

// NewDeferredDimensionRegistry creates a registry with all known deferred dimensions.
func NewDeferredDimensionRegistry() *DeferredDimensionRegistry {
	return &DeferredDimensionRegistry{
		dimensions: map[string]DeferredDimension{
			"sysParams": {
				Name:      "sysParams",
				ErrorCode: "DEFERRED_DIMENSION_SYS_PARAMS",
				Message:   "system-variable permission assignment is not supported in the current permission center; use system variable management for variable definitions",
			},
			"whiteList": {
				Name:      "whiteList",
				ErrorCode: "DEFERRED_DIMENSION_WHITELIST",
				Message:   "whitelist-based row permission is not supported in the current permission center; this dimension is deferred",
			},
			"dept": {
				Name:      "dept",
				ErrorCode: "DEFERRED_DIMENSION_DEPT",
				Message:   "department-based permission assignment is not supported in the current permission center; this dimension is deferred",
			},
		},
	}
}

// IsDeferred returns true if the dimension name is registered as deferred.
func (r *DeferredDimensionRegistry) IsDeferred(dimName string) bool {
	_, ok := r.dimensions[dimName]
	return ok
}

// GetRejectionError returns a formatted error for the deferred dimension.
func (r *DeferredDimensionRegistry) GetRejectionError(dimName string) error {
	dim, ok := r.dimensions[dimName]
	if !ok {
		return fmt.Errorf("permission dimension %q is not supported", dimName)
	}
	return fmt.Errorf("[%s] %s", dim.ErrorCode, dim.Message)
}

// ListDeferred returns all registered deferred dimensions.
func (r *DeferredDimensionRegistry) ListDeferred() []DeferredDimension {
	result := make([]DeferredDimension, 0, len(r.dimensions))
	for _, dim := range r.dimensions {
		result = append(result, dim)
	}
	return result
}
