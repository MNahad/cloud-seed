package orchestration

type step[T call | condition | ret] map[string]T

type steps[T call | condition | ret] struct {
	Steps []step[T]
}

type jump struct {
	Next string `json:"next"`
}

type call struct {
	Call string `json:"call"`
	Args struct {
		Url  string    `json:"url"`
		Auth *callAuth `json:"auth"`
		Body *string   `json:"body"`
	} `json:"args"`
	Result *string `json:"result"`
	jump
}

type callAuth struct {
	Type *string `json:"type"`
}

type condition struct {
	Switch conditionSwitch `json:"switch"`
}

type conditionSwitch []struct {
	Condition string `json:"condition"`
	jump
}

type ret struct {
	Return string `json:"return"`
}

type stepsCollection struct {
	StepsCall      steps[call]
	StepsCondition steps[condition]
	StepsReturn    steps[ret]
}

type main struct {
	Params []string `json:"params"`
	Steps  []any    `json:"steps"`
}

type workflow struct {
	Main main `json:"main"`
}
