package errors

import "fmt"

// Error types for better error handling
type (
	// NotFoundError is returned when a resource is not found
	NotFoundError struct {
		ID   string
		Path string
	}

	// ConflictError is returned when a resource already exists
	ConflictError struct {
		ID string
	}

	// InvalidRequestError is returned when the request is invalid
	InvalidRequestError struct {
		Reason string
	}

	// StorageError is returned when there's an error in storage operations
	StorageError struct {
		Operation string
		Err       error
	}
)

func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("resource with ID %s not found", e.ID)
	}
	return fmt.Sprintf("resource at path %s not found", e.Path)
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("resource with ID %s already exists", e.ID)
}

func (e *InvalidRequestError) Error() string {
	return fmt.Sprintf("invalid request: %s", e.Reason)
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("storage error during %s: %v", e.Operation, e.Err)
}

// NewNotFoundError creates a new NotFoundError with the given ID and path.
func NewNotFoundError(id, path string) error {
	return &NotFoundError{ID: id, Path: path}
}

// NewConflictError creates a new ConflictError with the given ID.
func NewConflictError(id string) error {
	return &ConflictError{ID: id}
}

// NewInvalidRequestError creates a new InvalidRequestError with the given reason.
func NewInvalidRequestError(reason string) error {
	return &InvalidRequestError{Reason: reason}
}

// NewStorageError creates a new StorageError with the given operation and error.
func NewStorageError(operation string, err error) error {
	return &StorageError{Operation: operation, Err: err}
}
