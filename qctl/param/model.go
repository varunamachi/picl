package param

import (
	"errors"
	"time"
)

type Type string

const (
	PtConstant  Type = "Constant"
	PtBoolean   Type = "Boolean"
	PtTristate  Type = "Tristate"
	PtChoice    Type = "Choice"
	PtNumber    Type = "Number"
	PtRange     Type = "Range"
	PtDate      Type = "Date"
	PtDateRange Type = "DateRange"
)

type Tristate string

const (
	On   Tristate = "On"
	Off  Tristate = "Off"
	None Tristate = "None"
)

type Range struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Divs  int     `json:"divs"`
}

type DateRnage struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type BoolOpts struct {
	TrueLabel  string `json:"trueLabel"`
	FalseLabel string `json:"falseLabel"`
}

type TristateOpts struct {
	TrueLabel  string `json:"trueLabel"`
	FalseLabel string `json:"falseLabel"`
	NoneLabel  string `json:"noneLabel"`
}

type Option struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ControlProps struct {
	BoolOpts     BoolOpts     `json:"boolOpts"`
	TristateOpts TristateOpts `json:"tristateOpts"`
	Options      []Option     `json:"options"`
	Range        Range        `json:"range"`
	DateRnage    DateRnage    `json:"dateRnage"`
	ConstVal     string       `json:"constVal"`
}

type ControlItem struct {
	Id    string       `json:"id"`
	Name  string       `json:"name"`
	Desc  string       `json:"desc"`
	Type  Type         `json:"type"`
	Props ControlProps `json:"props"`
}

type ControlGroup struct {
	Id    string         `json:"id"`
	Name  string         `json:"name"`
	Desc  string         `json:"desc"`
	Items []*ControlItem `json:"items"`
}

var (
	ErrParamNotFound     = errors.New("qctl.paramNotFound")
	ErrParamGetFailed    = errors.New("qctl.paramGetFailed")
	ErrParamSetFailed    = errors.New("qctl.paramSetFailed")
	ErrInvalidParamValue = errors.New("qctl.invalidParamValue")
)

type Operator interface {
	Get() (any, error)
	Set(value any) error
	Default() any
}

type Value interface {
	bool | string | Tristate | float64 | Range | DateRnage | time.Time
}
