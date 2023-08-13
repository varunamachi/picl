package sample

import (
	"encoding/json"

	_ "embed"

	"github.com/rs/zerolog/log"
	"github.com/varunamachi/libx/errx"
	"github.com/varunamachi/picl/qctl/param"
)

// var values = map[string]any{}
// var defaults = map[string]any{}
// var paramMap = map[string]*param.ControlItem{}

var groups = make([]*param.ControlGroup, 0, 100)
var operators = map[string]param.Operator{}

//go:embed params.json
var paramsData []byte

func init() {

	err := json.Unmarshal(paramsData, &groups)
	if err != nil {
		log.Fatal().Err(err).
			Msg("failed to read sample param json embedded file")
	}

	for _, g := range groups {
		for _, item := range g.Items {
			if _, found := operators[item.Id]; found {
				log.Warn().Str("itemId", item.Id).
					Msg("duplicate item ids found, latest one will be used")
				log.Warn().Msg("NOTE: only item ids are considered to " +
					"identify control parameters. Groups are only used for " +
					"organization")
			}

			var opr param.Operator

			switch item.Type {
			case param.PtConstant:
				opr = newSampleOpr(
					item, item.Props.ConstVal, item.Props.ConstVal)
			case param.PtBoolean:
				opr = newSampleOpr(item, false, false)
			case param.PtTristate:
				opr = newSampleOpr(
					item, param.None, param.None)
			case param.PtChoice:
				if len(item.Props.Options) == 0 {
					continue
				}
				opt := item.Props.Options[0].Value
				opr = newSampleOpr(item, opt, opt)
			case param.PtNumber:
				opr = newSampleOpr(
					item, item.Props.Range.Start, item.Props.Range.Start)
			case param.PtRange:
				opr = newSampleOpr(item, item.Props.Range, item.Props.Range)
			case param.PtDate:
				opr = newSampleOpr(
					item, item.Props.Range.Start, item.Props.Range.Start)
			case param.PtDateRange:
				dr := param.DateRnage{
					Start: item.Props.DateRnage.Start,
					End:   item.Props.DateRnage.Start.AddDate(0, 0, 1),
				}
				opr = newSampleOpr(item, dr, dr)
			}

			operators[item.Id] = opr
		}
	}

	// load json file
	// go through the control groups, items etc
	// populate values map with default values from the json

}

type sampleOperator[T param.Value] struct {
	item         *param.ControlItem
	defaultValue T
	value        T
}

func newSampleOpr[T param.Value](
	item *param.ControlItem, def, val T) *sampleOperator[T] {
	return &sampleOperator[T]{
		item:         item,
		defaultValue: def,
		value:        val,
	}
}

func (so sampleOperator[T]) Get() (any, error) {
	return so.value, nil
}

func (so sampleOperator[T]) Set(value any) error {
	val, ok := value.(T)
	if !ok {
		return errx.Errf(param.ErrInvalidParamValue,
			"invalid value given for param '%s': '%v'", so.item.Id, value)
	}
	so.value = val
	return nil
}

func (so sampleOperator[T]) Default() any {
	return so.defaultValue
}
