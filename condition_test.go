package condition_test

import (
	"github.com/innotechdevops/condition"
	"testing"
)

func TestToExpression(t *testing.T) {
	// Given
	jsonString := `[{"operand": "x", "operator": ">", "value": 10}, {"operator": "&&"}, {"operand": "x", "operator": "<", "value": 20}]`

	// When
	actual := condition.ToExpression(jsonString)

	// Then
	if actual != "x > 10.000000 && x < 20.000000" {
		t.Error("Error not match", actual)
	}
}

func TestEval(t *testing.T) {
	// Given
	expression := "x > 10 && x < 20"
	value := map[string]any{
		"x": 15,
	}

	// When
	actual := condition.Eval(expression, value)

	// Then
	if !actual {
		t.Error("Error", actual)
	}
}
