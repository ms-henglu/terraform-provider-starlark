// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math/big"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// Ensure the implementation satisfies the interface.
var _ function.Function = Eval{}

func NewEvalFunction() function.Function {
	return Eval{}
}

// Eval implements the "eval" function.
type Eval struct{}

func (f Eval) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "eval"
}

func (f Eval) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Execute a Starlark script",
		Description: "Executes the provided Starlark script with the given inputs.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "script",
				Description: "The Starlark source code to execute.",
			},
			function.DynamicParameter{
				Name:        "inputs",
				Description: "A map of variables to inject into the Starlark global scope.",
			},
		},
		Return: function.DynamicReturn{},
	}
}

func (f Eval) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var script string
	var inputs types.Dynamic

	// Read Terraform arguments
	resp.Error = req.Arguments.Get(ctx, &script, &inputs)
	if resp.Error != nil {
		return
	}

	// Create Starlark thread
	thread := &starlark.Thread{
		Name:  "terraform-provider-starlark-eval",
		Print: func(_ *starlark.Thread, msg string) { fmt.Println(msg) }, // Optional: wire up to TF logs?
	}

	// Convert inputs to Starlark types
	globals := starlark.StringDict{}

	if !inputs.IsNull() && !inputs.IsUnknown() {
		val, err := attrValueToStarlark(ctx, inputs)
		if err != nil {
			resp.Error = function.NewFuncError(fmt.Sprintf("failed to convert inputs: %s", err))
			return
		}

		dict, ok := val.(*starlark.Dict)
		if !ok {
			resp.Error = function.NewFuncError(fmt.Sprintf("inputs must be a map or object, got %s", val.Type()))
			return
		}

		for _, item := range dict.Items() {
			k, ok := item[0].(starlark.String)
			if !ok {
				// Should have been checked upstream or during conversion if we enforce string keys
				resp.Error = function.NewFuncError(fmt.Sprintf("input keys must be strings, got %s", item[0].Type()))
				return
			}
			globals[string(k)] = item[1]
		}
	}

	// Execute Starlark script
	// We use ExecFile to execute the script. The script should assign the result to a global variable named "result"
	// or simply define variables. To support the requirement "usually the value of the last expression",
	// implementing that with ExecFile is tricky because it executes statements.
	// However, if the user defines a "result" variable, we pick it up.
	// As a fallback/alternative, if the script is a single expression, Eval calls might be appropriate, but scripts are usually multiple lines.
	// We will look for a global variable named "result".

	scriptGlobals, err := starlark.ExecFileOptions(&syntax.FileOptions{Recursion: true, While: true}, thread, "script.star", script, globals)
	if err != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("starlark execution failed: %s", err))
		return
	}

	// Extract result
	resultVal, ok := scriptGlobals["result"]
	if !ok {
		// If no "result" variable, return null or error?
		// Let's assume return null is safer if they just did things for side effects (though side effects in a pure function are useless).
		// Re-reading usage: "Return: result ... usually the value of the last expression or a specific global variable like result"
		// Since we can't easily get the last expression value from ExecFile, checking for "result" is the most robust contract.
		resp.Error = resp.Result.Set(ctx, types.DynamicNull())
		return
	}

	// Convert Starlark result back to Terraform
	tfVal, err := starlarkToTFValue(ctx, resultVal)
	if err != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("failed to convert result: %s", err))
		return
	}

	tfVal = types.DynamicValue(tfVal)
	resp.Error = resp.Result.Set(ctx, tfVal)
}

// attrValueToStarlark converts a Terraform attr.Value (including Dynamic) to a Starlark value.
func attrValueToStarlark(ctx context.Context, val attr.Value) (starlark.Value, error) {
	if val.IsNull() || val.IsUnknown() {
		return starlark.None, nil
	}

	switch v := val.(type) {
	case types.String:
		return starlark.String(v.ValueString()), nil
	case types.Bool:
		return starlark.Bool(v.ValueBool()), nil
	case types.Int64:
		return starlark.MakeInt64(v.ValueInt64()), nil
	case types.Float64:
		return starlark.Float(v.ValueFloat64()), nil
	case types.Number:
		// types.Number is big.Float
		f, _ := v.ValueBigFloat().Float64()
		return starlark.Float(f), nil
	case types.List:
		return listToStarlarkList(ctx, v.Elements())
	case types.Tuple:
		return listToStarlarkList(ctx, v.Elements())
	case types.Map:
		return mapToStarlarkDict(ctx, v.Elements())
	case types.Object:
		return mapToStarlarkDict(ctx, v.Attributes())
	case types.Dynamic:
		return attrValueToStarlark(ctx, v.UnderlyingValue())
	default:
		return nil, fmt.Errorf("unsupported attribute type: %T", v)
	}
}

func listToStarlarkList(ctx context.Context, elements []attr.Value) (*starlark.List, error) {
	var elems []starlark.Value
	for _, elem := range elements {
		conv, err := attrValueToStarlark(ctx, elem)
		if err != nil {
			return nil, err
		}
		elems = append(elems, conv)
	}
	return starlark.NewList(elems), nil
}

func mapToStarlarkDict(ctx context.Context, elements map[string]attr.Value) (*starlark.Dict, error) {
	dict := starlark.NewDict(len(elements))
	keys := make([]string, 0, len(elements))
	for k := range elements {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		elem := elements[k]
		conv, err := attrValueToStarlark(ctx, elem)
		if err != nil {
			return nil, err
		}
		if err := dict.SetKey(starlark.String(k), conv); err != nil {
			return nil, err
		}
	}
	return dict, nil
}

// starlarkToTFValue converts a Starlark value to a Terraform Dynamic value.
func starlarkToTFValue(ctx context.Context, val starlark.Value) (attr.Value, error) {
	switch v := val.(type) {
	case starlark.NoneType:
		return types.DynamicNull(), nil
	case starlark.String:
		return types.StringValue(string(v)), nil
	case starlark.Bool:
		return types.BoolValue(bool(v)), nil
	case starlark.Int:
		// Try Int64
		if i, ok := v.Int64(); ok {
			return types.Int64Value(i), nil
		}
		// Try BigInt -> Number? Or Float.
		// Use Float for safety if it assumes Number
		return types.NumberValue(new(big.Float).SetInt(v.BigInt())), nil
	case starlark.Float:
		return types.Float64Value(float64(v)), nil
	case *starlark.List:
		// Convert list to TupleValue for flexibility with varied types
		n := v.Len()
		elemTypes := make([]attr.Type, 0, n)
		elemValues := make([]attr.Value, 0, n)

		for i := 0; i < n; i++ {
			elemVal := v.Index(i)
			tfVal, err := starlarkToTFValue(ctx, elemVal)
			if err != nil {
				return nil, err
			}
			elemTypes = append(elemTypes, tfVal.Type(ctx))
			elemValues = append(elemValues, tfVal)
		}
		tupVal, diags := types.TupleValue(elemTypes, elemValues)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to create tuple: %s", diags)
		}
		return tupVal, nil

	case *starlark.Dict:
		// Convert to ObjectValue for flexibility
		attrTypes := make(map[string]attr.Type)
		attrValues := make(map[string]attr.Value)

		for _, k := range v.Keys() {
			ks, ok := k.(starlark.String)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings, got %s", k.Type())
			}
			keyStr := string(ks)

			val, _, _ := v.Get(k)
			tfVal, err := starlarkToTFValue(ctx, val)
			if err != nil {
				return nil, err
			}

			attrTypes[keyStr] = tfVal.Type(ctx)
			attrValues[keyStr] = tfVal
		}

		objVal, diags := types.ObjectValue(attrTypes, attrValues)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to create object: %s", diags)
		}
		return objVal, nil

	default:
		return nil, fmt.Errorf("unsupported starlark return type: %s", v.Type())
	}
}
