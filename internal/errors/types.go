package errors

type ValidationError struct {
    message string
}

type AuthError struct {
    message string
}

type NotFoundError struct {
    message string
}

type PaymentError struct {
    message string
    code    string
}

// Implement error interface for all types
func (e *ValidationError) Error() string { return e.message }
func (e *AuthError) Error() string      { return e.message }
func (e *NotFoundError) Error() string  { return e.message }
func (e *PaymentError) Error() string   { return e.message }

// Constructor functions
func NewValidationError(msg string) *ValidationError {
    return &ValidationError{message: msg}
}

func NewAuthError(msg string) *AuthError {
    return &AuthError{message: msg}
}

func NewNotFoundError(msg string) *NotFoundError {
    return &NotFoundError{message: msg}
}

func NewPaymentError(msg, code string) *PaymentError {
    return &PaymentError{message: msg, code: code}
}