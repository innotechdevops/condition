# condition

Execute statement by condition for Golang.

## Install

```shell
go get github.com/innotechdevops/condition
```

## How to use

```go
value := 9.9
conditions := []Conditions{
	{ Operand: "value", Operator: ">=", Value: 1.5 },
	{ Operator: "&&" },
	{ Operand: "value", Operator: "<=", Value: 9.9 },
}
result := condition.Execute(conditions, map[string]any{"value": value})
```