package studyview

// Button represents a button displayed on each card face
type Button string

// The buttons displayed on each card
const (
	ButtonLeft        Button = "button-l"
	ButtonCenterLeft  Button = "button-cl"
	ButtonCenterRight Button = "button-cr"
	ButtonRight       Button = "button-r"
)

func (b Button) String() string {
	switch b {
	case ButtonLeft:
		return "Left"
	case ButtonCenterLeft:
		return "Center Left"
	case ButtonCenterRight:
		return "Center Right"
	case ButtonRight:
		return "Right"
	}
	return "Unknown"
}

// ButtonMap is the state of the three answer buttons
type ButtonMap map[Button]ButtonState

// ButtonState is one of the four answer buttons.
type ButtonState struct {
	Name    string
	Enabled bool
}
