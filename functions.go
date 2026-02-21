package forza

// FunctionProps defines properties for a function parameter.
type FunctionProps struct {
	Description string
	Required    bool
}

// FunctionShape maps parameter names to their properties.
type FunctionShape map[string]FunctionProps

// WithProperty returns an option that adds a parameter definition to a FunctionShape.
func WithProperty(name, description string, required bool) func(FunctionShape) {
	return func(shape FunctionShape) {
		shape[name] = FunctionProps{
			Description: description,
			Required:    required,
		}
	}
}

// NewFunction creates a new FunctionShape with the given property options.
func NewFunction(properties ...func(FunctionShape)) FunctionShape {
	shape := make(FunctionShape)
	for _, prop := range properties {
		prop(shape)
	}
	return shape
}
