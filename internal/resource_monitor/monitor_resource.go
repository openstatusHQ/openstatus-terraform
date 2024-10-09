package resource_monitor

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/numberplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func MonitorResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "If the monitor is active",
				MarkdownDescription: "If the monitor is active",
				Default:             booldefault.StaticBool(false),
			},
			"assertions": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"compare": schema.StringAttribute{
							Required:            true,
							Description:         "The comparison to run",
							MarkdownDescription: "The comparison to run",
							Validators: []validator.String{
								stringvalidator.OneOf(
									"eq",
									"not_eq",
									"gt",
									"gte",
									"lt",
									"lte",
									"contains",
									"not_contains",
									"not_empty",
									"empty",
								),
							},
						},
						"target": schema.StringAttribute{
							Required:            true,
							Description:         "The target value",
							MarkdownDescription: "The target value",
						},
						"key": schema.StringAttribute{
							Optional:            true,
							Description:         "The key to check",
							MarkdownDescription: "The key to check",
						},
						"type": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									"status",
									"header",
									"textBody",
								),
							},
						},
					},
					CustomType: AssertionsType{
						ObjectType: types.ObjectType{
							AttrTypes: AssertionsValue{}.AttributeTypes(ctx),
						},
					},
				},
				Optional:            true,
				Computed:            true,
				Description:         "The assertions to run",
				MarkdownDescription: "The assertions to run",
			},
			"body": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The body",
				MarkdownDescription: "The body",
				Default:             stringdefault.StaticString(""),
			},
			"degraded_after": schema.NumberAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The time after the monitor is considered degraded",
				MarkdownDescription: "The time after the monitor is considered degraded",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The description of your monitor",
				MarkdownDescription: "The description of your monitor",
			},
			"headers": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required: true,
						},
						"value": schema.StringAttribute{
							Required: true,
						},
					},
					CustomType: HeadersType{
						ObjectType: types.ObjectType{
							AttrTypes: HeadersValue{}.AttributeTypes(ctx),
						},
					},
				},
				Optional:            true,
				Computed:            true,
				Description:         "The headers of your request",
				MarkdownDescription: "The headers of your request",
			},
			"id": schema.NumberAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The id of the monitor",
				MarkdownDescription: "The id of the monitor",
				PlanModifiers: []planmodifier.Number{
					numberplanmodifier.UseStateForUnknown(),
				},
			},
			"method": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"GET",
						"POST",
						"HEAD",
					),
				},
				Default: stringdefault.StaticString("GET"),
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the monitor",
				MarkdownDescription: "The name of the monitor",
			},
			"periodicity": schema.StringAttribute{
				Required:            true,
				Description:         "How often the monitor should run",
				MarkdownDescription: "How often the monitor should run",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"30s",
						"1m",
						"5m",
						"10m",
						"30m",
						"1h",
						"other",
					),
				},
			},
			"public": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "If the monitor is public",
				MarkdownDescription: "If the monitor is public",
				Default:             booldefault.StaticBool(false),
			},
			"regions": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "Where we should monitor it",
				MarkdownDescription: "Where we should monitor it",
			},
			"timeout": schema.NumberAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "The timeout of the request",
				MarkdownDescription: "The timeout of the request",
			},
			"url": schema.StringAttribute{
				Required:            true,
				Description:         "The url to monitor",
				MarkdownDescription: "The url to monitor",
			},
		},
	}
}

type MonitorModel struct {
	Active        types.Bool   `tfsdk:"active"`
	Assertions    types.List   `tfsdk:"assertions"`
	Body          types.String `tfsdk:"body"`
	DegradedAfter types.Number `tfsdk:"degraded_after"`
	Description   types.String `tfsdk:"description"`
	Headers       types.List   `tfsdk:"headers"`
	Id            types.Number `tfsdk:"id"`
	Method        types.String `tfsdk:"method"`
	Name          types.String `tfsdk:"name"`
	Periodicity   types.String `tfsdk:"periodicity"`
	Public        types.Bool   `tfsdk:"public"`
	Regions       types.List   `tfsdk:"regions"`
	Timeout       types.Number `tfsdk:"timeout"`
	Url           types.String `tfsdk:"url"`
}

var _ basetypes.ObjectTypable = AssertionsType{}

type AssertionsType struct {
	basetypes.ObjectType
}

func (t AssertionsType) Equal(o attr.Type) bool {
	other, ok := o.(AssertionsType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t AssertionsType) String() string {
	return "AssertionsType"
}

func (t AssertionsType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	compareAttribute, ok := attributes["compare"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`compare is missing from object`)

		return nil, diags
	}

	compareVal, ok := compareAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`compare expected to be basetypes.StringValue, was: %T`, compareAttribute))
	}

	targetAttribute, ok := attributes["target"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`target is missing from object`)

		return nil, diags
	}

	targetVal, ok := targetAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`target expected to be basetypes.StringValue, was: %T`, targetAttribute))
	}

	typeAttribute, ok := attributes["type"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`type is missing from object`)

		return nil, diags
	}

	typeVal, ok := typeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`type expected to be basetypes.StringValue, was: %T`, typeAttribute))
	}

	keyAttribute, ok := attributes["key"]
	if !ok {
		diags.AddError(
			"Attribute Missing",
			`type is missing from object`)

		return nil, diags
	}

	keyVal, ok := keyAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`type expected to be basetypes.StringValue, was: %T`, keyAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return AssertionsValue{
		Compare:        compareVal,
		Target:         targetVal,
		AssertionsType: typeVal,
		Key:            keyVal,
		state:          attr.ValueStateKnown,
	}, diags
}

func NewAssertionsValueNull() AssertionsValue {
	return AssertionsValue{
		state: attr.ValueStateNull,
	}
}

func NewAssertionsValueUnknown() AssertionsValue {
	return AssertionsValue{
		state: attr.ValueStateUnknown,
	}
}

func NewAssertionsValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (AssertionsValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing AssertionsValue Attribute Value",
				"While creating a AssertionsValue value, a missing attribute value was detected. "+
					"A AssertionsValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("AssertionsValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid AssertionsValue Attribute Type",
				"While creating a AssertionsValue value, an invalid attribute value was detected. "+
					"A AssertionsValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("AssertionsValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("AssertionsValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra AssertionsValue Attribute Value",
				"While creating a AssertionsValue value, an extra attribute value was detected. "+
					"A AssertionsValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra AssertionsValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewAssertionsValueUnknown(), diags
	}

	compareAttribute, ok := attributes["compare"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`compare is missing from object`)

		return NewAssertionsValueUnknown(), diags
	}

	compareVal, ok := compareAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`compare expected to be basetypes.StringValue, was: %T`, compareAttribute))
	}

	targetAttribute, ok := attributes["target"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`target is missing from object`)

		return NewAssertionsValueUnknown(), diags
	}

	targetVal, ok := targetAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`target expected to be basetypes.StringValue, was: %T`, targetAttribute))
	}

	typeAttribute, ok := attributes["type"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`type is missing from object`)

		return NewAssertionsValueUnknown(), diags
	}

	typeVal, ok := typeAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`type expected to be basetypes.StringValue, was: %T`, typeAttribute))
	}

	keyAttribute, ok := attributes["key"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`type is missing from object`)

		return NewAssertionsValueUnknown(), diags
	}

	keyVal, ok := keyAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`type expected to be basetypes.StringValue, was: %T`, typeAttribute))
	}

	if diags.HasError() {
		return NewAssertionsValueUnknown(), diags
	}

	return AssertionsValue{
		Compare:        compareVal,
		Target:         targetVal,
		AssertionsType: typeVal,
		Key:            keyVal,
		state:          attr.ValueStateKnown,
	}, diags
}

func NewAssertionsValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) AssertionsValue {
	object, diags := NewAssertionsValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewAssertionsValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t AssertionsType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewAssertionsValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewAssertionsValueUnknown(), nil
	}

	if in.IsNull() {
		return NewAssertionsValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewAssertionsValueMust(AssertionsValue{}.AttributeTypes(ctx), attributes), nil
}

func (t AssertionsType) ValueType(ctx context.Context) attr.Value {
	return AssertionsValue{}
}

var _ basetypes.ObjectValuable = AssertionsValue{}

type AssertionsValue struct {
	Compare        basetypes.StringValue `tfsdk:"compare"`
	Target         basetypes.StringValue `tfsdk:"target"`
	Key            basetypes.StringValue `tfsdk:"key"`
	AssertionsType basetypes.StringValue `tfsdk:"type"`

	state attr.ValueState
}

func (v AssertionsValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 3)

	var val tftypes.Value
	var err error

	attrTypes["compare"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["target"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["type"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["key"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 4)

		val, err = v.Compare.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["compare"] = val

		val, err = v.Key.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["key"] = val

		val, err = v.Target.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["target"] = val

		val, err = v.AssertionsType.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["type"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v AssertionsValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v AssertionsValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v AssertionsValue) String() string {
	return "AssertionsValue"
}

func (v AssertionsValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"compare": basetypes.StringType{},
			"target":  basetypes.StringType{},
			"type":    basetypes.StringType{},
			"key":     basetypes.StringType{},
		},
		map[string]attr.Value{
			"compare": v.Compare,
			"target":  v.Target,
			"type":    v.AssertionsType,
			"key":     v.Key,
		})

	return objVal, diags
}

func (v AssertionsValue) Equal(o attr.Value) bool {
	other, ok := o.(AssertionsValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Compare.Equal(other.Compare) {
		return false
	}

	if !v.Target.Equal(other.Target) {
		return false
	}

	if !v.AssertionsType.Equal(other.AssertionsType) {
		return false
	}

	return true
}

func (v AssertionsValue) Type(ctx context.Context) attr.Type {
	return AssertionsType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v AssertionsValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"compare": basetypes.StringType{},
		"target":  basetypes.StringType{},
		"type":    basetypes.StringType{},
		"key":     basetypes.StringType{},
	}
}

var _ basetypes.ObjectTypable = HeadersType{}

type HeadersType struct {
	basetypes.ObjectType
}

func (t HeadersType) Equal(o attr.Type) bool {
	other, ok := o.(HeadersType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t HeadersType) String() string {
	return "HeadersType"
}

func (t HeadersType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	keyAttribute, ok := attributes["key"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`key is missing from object`)

		return nil, diags
	}

	keyVal, ok := keyAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`key expected to be basetypes.StringValue, was: %T`, keyAttribute))
	}

	valueAttribute, ok := attributes["value"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`value is missing from object`)

		return nil, diags
	}

	valueVal, ok := valueAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`value expected to be basetypes.StringValue, was: %T`, valueAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return HeadersValue{
		Key:   keyVal,
		Value: valueVal,
		state: attr.ValueStateKnown,
	}, diags
}

func NewHeadersValueNull() HeadersValue {
	return HeadersValue{
		state: attr.ValueStateNull,
	}
}

func NewHeadersValueUnknown() HeadersValue {
	return HeadersValue{
		state: attr.ValueStateUnknown,
	}
}

func NewHeadersValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (HeadersValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing HeadersValue Attribute Value",
				"While creating a HeadersValue value, a missing attribute value was detected. "+
					"A HeadersValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("HeadersValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid HeadersValue Attribute Type",
				"While creating a HeadersValue value, an invalid attribute value was detected. "+
					"A HeadersValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("HeadersValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("HeadersValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra HeadersValue Attribute Value",
				"While creating a HeadersValue value, an extra attribute value was detected. "+
					"A HeadersValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra HeadersValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewHeadersValueUnknown(), diags
	}

	keyAttribute, ok := attributes["key"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`key is missing from object`)

		return NewHeadersValueUnknown(), diags
	}

	keyVal, ok := keyAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`key expected to be basetypes.StringValue, was: %T`, keyAttribute))
	}

	valueAttribute, ok := attributes["value"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`value is missing from object`)

		return NewHeadersValueUnknown(), diags
	}

	valueVal, ok := valueAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`value expected to be basetypes.StringValue, was: %T`, valueAttribute))
	}

	if diags.HasError() {
		return NewHeadersValueUnknown(), diags
	}

	return HeadersValue{
		Key:   keyVal,
		Value: valueVal,
		state: attr.ValueStateKnown,
	}, diags
}

func NewHeadersValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) HeadersValue {
	object, diags := NewHeadersValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewHeadersValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t HeadersType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewHeadersValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewHeadersValueUnknown(), nil
	}

	if in.IsNull() {
		return NewHeadersValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewHeadersValueMust(HeadersValue{}.AttributeTypes(ctx), attributes), nil
}

func (t HeadersType) ValueType(ctx context.Context) attr.Value {
	return HeadersValue{}
}

var _ basetypes.ObjectValuable = HeadersValue{}

type HeadersValue struct {
	Key   basetypes.StringValue `tfsdk:"key"`
	Value basetypes.StringValue `tfsdk:"value"`
	state attr.ValueState
}

func (v HeadersValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 2)

	var val tftypes.Value
	var err error

	attrTypes["key"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["value"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 2)

		val, err = v.Key.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["key"] = val

		val, err = v.Value.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["value"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v HeadersValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v HeadersValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v HeadersValue) String() string {
	return "HeadersValue"
}

func (v HeadersValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	objVal, diags := types.ObjectValue(
		map[string]attr.Type{
			"key":   basetypes.StringType{},
			"value": basetypes.StringType{},
		},
		map[string]attr.Value{
			"key":   v.Key,
			"value": v.Value,
		})

	return objVal, diags
}

func (v HeadersValue) Equal(o attr.Value) bool {
	other, ok := o.(HeadersValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Key.Equal(other.Key) {
		return false
	}

	if !v.Value.Equal(other.Value) {
		return false
	}

	return true
}

func (v HeadersValue) Type(ctx context.Context) attr.Type {
	return HeadersType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v HeadersValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"key":   basetypes.StringType{},
		"value": basetypes.StringType{},
	}
}
