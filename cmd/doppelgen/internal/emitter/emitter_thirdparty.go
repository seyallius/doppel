// Package emitter. emitter_thirdparty - shows the targeted additions and changes
// to emitter.go for third-party type detection and convention-stub emission.
//
// These changes integrate with the new TypeCategory values (CatThirdPartyStruct,
// CatThirdPartyPtrStruct) and FieldInfo flags (IsThirdParty, ElemIsThirdParty,
// ValueIsThirdParty) introduced in types.go.
//
// Merge instructions for emitter.go:
//  1. Replace collectImports with the version below.
//  2. Replace emitDeepField with the version below.
//  3. Add emitThirdPartyField as a new private helper.
//  4. Add emitThirdPartySliceElem and emitThirdPartyMapValue helpers.
package emitter

import (
	"fmt"

	"github.com/seyallius/doppel/cmd/doppelgen/internal/types"
)

// emitThirdPartyField emits a convention-function call for a struct field whose
// type originates from an external (third-party) Go module. The generated code:
//
//  1. Includes a "todo" comment identifying the required function signature.
//  2. Calls "Clone<StructName><FieldName>(x.<FieldName>)".
//  3. Wraps errors via core.WrapError.
//
// The convention function must be implemented by the library user in the same
// package as the generated file. doppelgen does not generate it.
func (e *Emitter) emitThirdPartyField(field types.FieldInfo, structName, varName string) error {
	cloneFn := fmt.Sprintf("Clone%s%s", structName, field.Name)

	e.emitLine(fmt.Sprintf("// Field: %s (third-party: %s) — auto-detected external type; convention function required.",
		field.Name, field.Type))
	e.emitLine(fmt.Sprintf("//todo(%s): implement a function with signature:", cloneFn))
	e.emitLine(fmt.Sprintf("//   func %s(src %s) (%s, error)", cloneFn, field.Type, field.Type))

	e.emitRaw(fmt.Sprintf("%s, err := %s(x.%s)", varName, cloneFn, field.Name))
	e.emitRaw("if err != nil {")
	e.indent++
	e.emitRaw(fmt.Sprintf("return nil, core.WrapError(%q, err)", structName+"."+field.Name))
	e.indent--
	e.emitRaw("}")

	return nil
}

// emitThirdPartySliceElem emits a CloneSlice call for a slice whose element type
// is a third-party struct. The element clone function must be provided by the user.
//
// Generated pattern:
//
//	// todo(Clone<Struct><Field>Elem): implement Clone for external element type <ElemType>
//	//   func Clone<Struct><Field>Elem(src <ElemType>) (<ElemType>, error)
//	cloned<Field>, err := manual.CloneSlice(x.<Field>, Clone<Struct><Field>Elem)
func (e *Emitter) emitThirdPartySliceElem(field types.FieldInfo, structName, varName string) error {
	cloneFn := fmt.Sprintf("Clone%s%sElem", structName, field.Name)

	e.emitLine(fmt.Sprintf("// Field: %s (tag: deep) — slice of third-party type %s.", field.Name, field.ElemType))
	e.emitLine(fmt.Sprintf("//todo(%s): implement a function with signature:", cloneFn))
	e.emitLine(fmt.Sprintf("//   func %s(src %s) (%s, error)", cloneFn, field.ElemType, field.ElemType))

	e.emitRaw(fmt.Sprintf("%s, err := manual.CloneSlice(x.%s, %s)", varName, field.Name, cloneFn))
	e.emitRaw("if err != nil {")
	e.indent++
	e.emitRaw(fmt.Sprintf("return nil, core.WrapError(%q, err)", structName+"."+field.Name))
	e.indent--
	e.emitRaw("}")

	return nil
}

// emitThirdPartyMapValue emits a CloneMap call for a map whose value type is a
// third-party struct. The value clone function must be provided by the user.
//
// Generated pattern:
//
//	// todo(Clone<Struct><Field>Val): implement Clone for external value type <ValueType>
//	//   func Clone<Struct><Field>Val(src <ValueType>) (<ValueType>, error)
//	cloned<Field>, err := manual.CloneMap(x.<Field>, Clone<Struct><Field>Val)
func (e *Emitter) emitThirdPartyMapValue(field types.FieldInfo, structName, varName string) error {
	cloneFn := fmt.Sprintf("Clone%s%sVal", structName, field.Name)

	e.emitLine(fmt.Sprintf("// Field: %s (tag: deep) — map with third-party value type %s.", field.Name, field.ValueType))
	e.emitLine(fmt.Sprintf("//todo(%s): implement a function with signature:", cloneFn))
	e.emitLine(fmt.Sprintf("//   func %s(src %s) (%s, error)", cloneFn, field.ValueType, field.ValueType))

	e.emitRaw(fmt.Sprintf("%s, err := manual.CloneMap(x.%s, %s)", varName, field.Name, cloneFn))
	e.emitRaw("if err != nil {")
	e.indent++
	e.emitRaw(fmt.Sprintf("return nil, core.WrapError(%q, err)", structName+"."+field.Name))
	e.indent--
	e.emitRaw("}")

	return nil
}
