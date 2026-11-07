package klarerrs

const (
	_ Code = ImplementationErrorPrefix + iota

	ErrMissingImpl       // Missing implementation for some targets
	ErrUnsupportedTarget // Object isn't supported for specific targets
	ErrImplPathNotFound  // File path used as implementation doesn't exist
	ErrMultipleImpls     // Implementation already declared for the same target

	// JavaScript target

	ErrReservedJSKeyword // Public object name can't be a JS keyword
	ErrConstructorName   // Field or enum item can't be named 'constructor'
	ErrRedeclaredJSName  // Object redeclared via @name attribute
	ErrInvalidJSName     // Name provided in @name attribute is an invalid JS identifier
)

func (e *Error) handleImplementationError() string {
	switch e.Code {
	default:
		e.noMessage()
		return ""

	case ErrReservedJSKeyword:
		return Quote(e.Name) + " is a reserved keyword in JavaScript and can't be used as a name"
	case ErrConstructorName:
		return Capitalize(WithA(e.Name)) + " can't be named 'constructor' in JavaScript"
	case ErrInvalidJSName:
		e.Hint("An identifier in JavaScript can be any Klar identifer or keyword that isn't reserved in JavaScript. They can also contain '$'.")
		return Quote(e.Name) + " isn't a valid JavaScript identifier"
	}
}
