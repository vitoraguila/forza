package forza

type functionProps struct {
	Description string
	Required    bool
}

type functionShape map[string]functionProps

func WithProperty(name, description string, required bool) func(functionShape) {
	return func(shape functionShape) {
		shape[name] = functionProps{
			Description: description,
			Required:    required,
		}
	}
}

func NewFunction(properties ...func(functionShape)) functionShape {
	shape := make(functionShape)
	for _, prop := range properties {
		prop(shape)
	}
	return shape
}
