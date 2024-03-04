package condition

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"reflect"
	"strings"
)

var Conditions = []string{">", ">=", "==", "<", "<=", "&&", "||", "offline"}
var IgnoreExecutes = []string{"offline"}

type Condition struct {
	Operand  string  `json:"operand,omitempty"`
	Operator string  `json:"operator"`
	Value    float64 `json:"value"`
}

func Match(find string, source ...string) bool {
	found := 0
	for _, c := range source {
		if find == c {
			found++
		}
	}
	return found > 0
}

func Supported(condition string) bool {
	return Match(condition, Conditions...)
}

func SupportedList(condition []Condition) bool {
	for _, c := range condition {
		if !Supported(c.Operator) {
			return false
		}
	}
	return true
}

func ToJSON(conditions []Condition) string {
	if conByte, err := json.Marshal(conditions); err == nil {
		return string(conByte)
	}
	return "[]"
}

func ToConditions(conditionJson string) []Condition {
	con := []Condition{}
	_ = json.Unmarshal([]byte(conditionJson), &con)
	return con
}

func IsIgnoreExecutes(expression string) bool {
	for _, ig := range IgnoreExecutes {
		if strings.Contains(expression, ig) {
			return true
		}
	}
	return false
}

func Execute(conditions []Condition, value map[string]any) bool {
	expression := Expression(conditions)

	if IsIgnoreExecutes(expression) {
		return false
	}

	return Eval(expression, value)
}

func Offline(conditions []Condition) bool {
	for _, c := range conditions {
		if c.Operator == "offline" {
			return true
		}
	}
	return false
}

func ExecuteJson(conditionsJson string, value map[string]any) bool {
	return Eval(ToExpression(conditionsJson), value)
}

func ToExpression(conditionsJSON string) string {
	cond, err := ParseConditionsJSON(conditionsJSON)
	if err != nil {
		return ""
	}
	return Expression(cond)
}

// ParseConditionsJSON parses the JSON representation of conditions into a slice of Condition structs.
func ParseConditionsJSON(conditionsJSON string) ([]Condition, error) {
	var conditions []Condition
	if err := json.Unmarshal([]byte(conditionsJSON), &conditions); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}
	return conditions, nil
}

// Expression converts an array of conditions to a CEL expression
func Expression(conditions []Condition) string {
	var parts []string

	for _, cond := range conditions {
		// If the condition is a logical operator, append it directly
		if cond.Operand == "" && cond.Operator != "" && cond.Value == 0 {
			parts = append(parts, cond.Operator)
		} else {
			// Otherwise, convert the individual condition to CEL
			parts = append(parts, cond.ToCEL())
		}
	}

	return strings.Join(parts, " ")
}

// ToCEL converts an individual condition to a CEL expression
func (c Condition) ToCEL() string {
	var builder strings.Builder

	// Append variable
	builder.WriteString(c.Operand)

	// If the operator is present, append it
	if c.Operator != "" {
		builder.WriteString(" ")
		builder.WriteString(c.Operator)
		builder.WriteString(" ")

		// If the value is present, append it
		if c.Operator != "&&" && c.Operator != "||" && c.Operator != "offline" {
			builder.WriteString(fmt.Sprintf("%f", c.Value))
		}
	}

	return builder.String()
}

func Eval(expression string, value map[string]any) bool {
	declare := []*exprpb.Decl{}
	for key, val := range value {
		t := decls.String
		switch reflect.TypeOf(val).Kind() {
		case reflect.Float32, reflect.Float64:
			t = decls.Double
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			t = decls.Int
		default:
		}
		declare = append(declare, decls.NewVar(key, t))
	}

	env, err := cel.NewEnv(
		cel.Declarations(declare...),
	)
	if err != nil {
		fmt.Println("new env error:", err)
	}

	// Parse the expression
	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		fmt.Println("type-check error:", issues.Err())
		return false
	}

	prg, err := env.Program(ast)
	if err != nil {
		fmt.Println("program error:", err)
		return false
	}

	out, dtl, err := prg.Eval(value)
	if err != nil {
		fmt.Println("eval error:", err, dtl)
		return false
	}

	return out.Value() == true
}
