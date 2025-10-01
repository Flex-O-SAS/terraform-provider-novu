// Types copied from novu-go/models/components/
// Each struct here has a field that needed to be fixed. When its properties do not need to be fixed, the type is components.xx
// Changes :  Slug type has been changed to string
// NB : This is not futureproof and we lose type methods, but I cba making a working unmarshaler and data format that handles everything

package api_client

import (
	"encoding/json"
	"fmt"

	"github.com/novuhq/novu-go/models/components"
)

type WorkflowResponseDto struct {
	// Name of the workflow
	Name string `json:"name"`
	// Description of the workflow
	Description *string `json:"description,omitempty"`
	// Tags associated with the workflow
	Tags []string `json:"tags,omitempty"`
	// Whether the workflow is active
	Active *bool `default:"false" json:"active"`
	// Unique identifier of the workflow
	ID string `json:"_id"`
	// Workflow identifier
	WorkflowID string `json:"workflowId"`
	// Slug of the workflow
	Slug string `json:"slug"`
	// Last updated timestamp
	UpdatedAt string `json:"updatedAt"`
	// Creation timestamp
	CreatedAt string `json:"createdAt"`
	// Steps of the workflow
	Steps []WorkflowResponseDtoSteps `json:"steps"`
	// Origin of the workflow
	Origin components.WorkflowOriginEnum `json:"origin"`
	// Preferences for the workflow
	Preferences components.WorkflowPreferencesResponseDto `json:"preferences"`
	// Status of the workflow
	Status components.WorkflowStatusEnum `json:"status"`
	// Runtime issues for workflow creation and update
	Issues map[string]components.RuntimeIssueDto `json:"issues,omitempty"`
	// Timestamp of the last workflow trigger
	LastTriggeredAt *string `json:"lastTriggeredAt,omitempty"`
	// The payload JSON Schema for the workflow
	PayloadSchema map[string]any `json:"payloadSchema,omitempty"`
	// Generated payload example based on the payload schema
	PayloadExample map[string]any `json:"payloadExample,omitempty"`
	// Whether payload schema validation is enabled
	ValidatePayload *bool `json:"validatePayload,omitempty"`
}

type WorkflowResponseDtoSteps struct {
	InAppStepResponseDto  *InAppStepResponseDto  `queryParam:"inline"`
	EmailStepResponseDto  *EmailStepResponseDto  `queryParam:"inline"`
	SmsStepResponseDto    *SmsStepResponseDto    `queryParam:"inline"`
	PushStepResponseDto   *PushStepResponseDto   `queryParam:"inline"`
	ChatStepResponseDto   *ChatStepResponseDto   `queryParam:"inline"`
	DelayStepResponseDto  *DelayStepResponseDto  `queryParam:"inline"`
	DigestStepResponseDto *DigestStepResponseDto `queryParam:"inline"`
	CustomStepResponseDto *CustomStepResponseDto `queryParam:"inline"`

	Type components.WorkflowResponseDtoStepsType
}

type StepResponseDto[C, CV any] struct {
	// Controls metadata for the in-app step
	Controls C `json:"controls"`
	// Control values for the in-app step
	ControlValues *CV `json:"controlValues,omitempty"`
	// JSON Schema for variables, follows the JSON Schema standard
	Variables map[string]any `json:"variables"`
	// Unique identifier of the step
	StepID string `json:"stepId"`
	// Database identifier of the step
	ID string `json:"_id"`
	// Name of the step
	Name string `json:"name"`
	// Slug of the step
	Slug string `json:"slug"`
	// Type of the step
	Type components.StepTypeEnum `json:"type"`
	// Origin of the workflow
	Origin components.WorkflowOriginEnum `json:"origin"`
	// Workflow identifier
	WorkflowID string `json:"workflowId"`
	// Workflow database identifier
	WorkflowDatabaseID string `json:"workflowDatabaseId"`
	// Issues associated with the step
	Issues *components.StepIssuesDto `json:"issues,omitempty"`
}

type InAppStepResponseDto struct {
	StepResponseDto[components.InAppControlsMetadataResponseDto, components.InAppStepResponseDtoControlValues]
}

type EmailStepResponseDto struct {
	StepResponseDto[components.EmailControlsMetadataResponseDto, components.EmailStepResponseDtoControlValues]
}

type SmsStepResponseDto struct {
	StepResponseDto[components.SmsControlsMetadataResponseDto, components.SmsStepResponseDtoControlValues]
}

type PushStepResponseDto struct {
	StepResponseDto[components.PushControlsMetadataResponseDto, components.PushStepResponseDtoControlValues]
}

type ChatStepResponseDto struct {
	StepResponseDto[components.ChatControlsMetadataResponseDto, components.ChatStepResponseDtoControlValues]
}

type DelayStepResponseDto struct {
	StepResponseDto[components.DelayControlsMetadataResponseDto, components.DelayStepResponseDtoControlValues]
}

type DigestStepResponseDto struct {
	StepResponseDto[components.DigestControlsMetadataResponseDto, components.DigestStepResponseDtoControlValues]
}

type CustomStepResponseDto struct {
	StepResponseDto[components.CustomControlsMetadataResponseDto, components.CustomStepResponseDtoControlValues]
}

// Copy of the unmarshaler from the novu-go package (without using the utils.UnmarshalJSON)
func (steps *WorkflowResponseDtoSteps) UnmarshalJSON(data []byte) error {
	type discriminator struct {
		Type string `json:"type"`
	}

	dis := new(discriminator)
	if err := json.Unmarshal(data, &dis); err != nil {
		return fmt.Errorf("could not unmarshal discriminator: %w", err)
	}

	switch dis.Type {
	case string(components.WorkflowResponseDtoStepsTypeInApp):
		if err := json.Unmarshal(data, &steps.InAppStepResponseDto); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == in_app) type InAppStepResponseDto within WorkflowResponseDtoSteps: %w", string(data), err)
		}
		steps.Type = components.WorkflowResponseDtoStepsTypeInApp
		return nil
	case string(components.WorkflowResponseDtoStepsTypeEmail):
		if err := json.Unmarshal(data, &steps.EmailStepResponseDto); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == email) type EmailStepResponseDto within WorkflowResponseDtoSteps: %w", string(data), err)
		}
		steps.Type = components.WorkflowResponseDtoStepsTypeEmail
		return nil
	case string(components.WorkflowResponseDtoStepsTypeSms):
		if err := json.Unmarshal(data, &steps.SmsStepResponseDto); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == sms) type SmsStepResponseDto within WorkflowResponseDtoSteps: %w", string(data), err)
		}
		steps.Type = components.WorkflowResponseDtoStepsTypeSms
		return nil
	case string(components.WorkflowResponseDtoStepsTypePush):
		if err := json.Unmarshal(data, &steps.PushStepResponseDto); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == push) type PushStepResponseDto within WorkflowResponseDtoSteps: %w", string(data), err)
		}
		steps.Type = components.WorkflowResponseDtoStepsTypePush
		return nil
	case string(components.WorkflowResponseDtoStepsTypeChat):
		if err := json.Unmarshal(data, &steps.ChatStepResponseDto); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == chat) type ChatStepResponseDto within WorkflowResponseDtoSteps: %w", string(data), err)
		}
		steps.Type = components.WorkflowResponseDtoStepsTypeChat
		return nil
	case string(components.WorkflowResponseDtoStepsTypeDelay):
		if err := json.Unmarshal(data, &steps.DelayStepResponseDto); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == delay) type DelayStepResponseDto within WorkflowResponseDtoSteps: %w", string(data), err)
		}
		steps.Type = components.WorkflowResponseDtoStepsTypeDelay
		return nil
	case string(components.WorkflowResponseDtoStepsTypeDigest):
		if err := json.Unmarshal(data, &steps.DigestStepResponseDto); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == digest) type DigestStepResponseDto within WorkflowResponseDtoSteps: %w", string(data), err)
		}
		steps.Type = components.WorkflowResponseDtoStepsTypeDigest
		return nil
	case string(components.WorkflowResponseDtoStepsTypeCustom):
		if err := json.Unmarshal(data, &steps.CustomStepResponseDto); err != nil {
			return fmt.Errorf("could not unmarshal `%s` into expected (Type == custom) type CustomStepResponseDto within WorkflowResponseDtoSteps: %w", string(data), err)
		}
		steps.Type = components.WorkflowResponseDtoStepsTypeCustom
		return nil
	}
	return fmt.Errorf("could not unmarshal `%s` into any supported union types for WorkflowResponseDtoSteps", string(data))
}

func (steps *WorkflowResponseDtoSteps) MarshalJSON() ([]byte, error) {
	if steps.InAppStepResponseDto != nil {
		return json.Marshal(steps.InAppStepResponseDto)
	}
	if steps.EmailStepResponseDto != nil {
		return json.Marshal(steps.EmailStepResponseDto)
	}
	if steps.SmsStepResponseDto != nil {
		return json.Marshal(steps.SmsStepResponseDto)
	}
	if steps.PushStepResponseDto != nil {
		return json.Marshal(steps.PushStepResponseDto)
	}
	if steps.ChatStepResponseDto != nil {
		return json.Marshal(steps.ChatStepResponseDto)
	}
	if steps.DelayStepResponseDto != nil {
		return json.Marshal(steps.DelayStepResponseDto)
	}
	if steps.DigestStepResponseDto != nil {
		return json.Marshal(steps.DigestStepResponseDto)
	}
	if steps.CustomStepResponseDto != nil {
		return json.Marshal(steps.CustomStepResponseDto)
	}
	return nil, fmt.Errorf("could not marshal union type WorkflowResponseDtoSteps: all supported fields are null")
}

/*
QUESTION FOR REVIEW : Do we prefer the generic version or the verbose version ?

type InAppStepResponseDto struct {
	// Controls metadata for the in-app step
	Controls components.InAppControlsMetadataResponseDto `json:"controls"`
	// Control values for the in-app step
	ControlValues *components.InAppStepResponseDtoControlValues `json:"controlValues,omitempty"`
	// JSON Schema for variables, follows the JSON Schema standard
	Variables map[string]any `json:"variables"`
	// Unique identifier of the step
	StepID string `json:"stepId"`
	// Database identifier of the step
	ID string `json:"_id"`
	// Name of the step
	Name string `json:"name"`
	// Slug of the step
	Slug string `json:"slug"`
	// Type of the step
	Type components.StepTypeEnum `json:"type"`
	// Origin of the workflow
	Origin components.WorkflowOriginEnum `json:"origin"`
	// Workflow identifier
	WorkflowID string `json:"workflowId"`
	// Workflow database identifier
	WorkflowDatabaseID string `json:"workflowDatabaseId"`
	// Issues associated with the step
	Issues *components.StepIssuesDto `json:"issues,omitempty"`
}

type EmailStepResponseDto struct {
	// Controls metadata for the email step
	Controls components.EmailControlsMetadataResponseDto `json:"controls"`
	// Control values for the email step
	ControlValues *components.EmailStepResponseDtoControlValues `json:"controlValues,omitempty"`
	// JSON Schema for variables, follows the JSON Schema standard
	Variables map[string]any `json:"variables"`
	// Unique identifier of the step
	StepID string `json:"stepId"`
	// Database identifier of the step
	ID string `json:"_id"`
	// Name of the step
	Name string `json:"name"`
	// Slug of the step
	Slug string `json:"slug"`
	// Type of the step
	Type components.StepTypeEnum `json:"type"`
	// Origin of the workflow
	Origin components.WorkflowOriginEnum `json:"origin"`
	// Workflow identifier
	WorkflowID string `json:"workflowId"`
	// Workflow database identifier
	WorkflowDatabaseID string `json:"workflowDatabaseId"`
	// Issues associated with the step
	Issues *components.StepIssuesDto `json:"issues,omitempty"`
}

type SmsStepResponseDto struct {
	// Controls metadata for the SMS step
	Controls components.SmsControlsMetadataResponseDto `json:"controls"`
	// Control values for the SMS step
	ControlValues *components.SmsStepResponseDtoControlValues `json:"controlValues,omitempty"`
	// JSON Schema for variables, follows the JSON Schema standard
	Variables map[string]any `json:"variables"`
	// Unique identifier of the step
	StepID string `json:"stepId"`
	// Database identifier of the step
	ID string `json:"_id"`
	// Name of the step
	Name string `json:"name"`
	// Slug of the step
	Slug string `json:"slug"`
	// Type of the step
	Type components.StepTypeEnum `json:"type"`
	// Origin of the workflow
	Origin components.WorkflowOriginEnum `json:"origin"`
	// Workflow identifier
	WorkflowID string `json:"workflowId"`
	// Workflow database identifier
	WorkflowDatabaseID string `json:"workflowDatabaseId"`
	// Issues associated with the step
	Issues *components.StepIssuesDto `json:"issues,omitempty"`
}

type PushStepResponseDto struct {
	// Controls metadata for the push step
	Controls components.PushControlsMetadataResponseDto `json:"controls"`
	// Control values for the push step
	ControlValues *components.PushStepResponseDtoControlValues `json:"controlValues,omitempty"`
	// JSON Schema for variables, follows the JSON Schema standard
	Variables map[string]any `json:"variables"`
	// Unique identifier of the step
	StepID string `json:"stepId"`
	// Database identifier of the step
	ID string `json:"_id"`
	// Name of the step
	Name string `json:"name"`
	// Slug of the step
	Slug string `json:"slug"`
	// Type of the step
	Type components.StepTypeEnum `json:"type"`
	// Origin of the workflow
	Origin components.WorkflowOriginEnum `json:"origin"`
	// Workflow identifier
	WorkflowID string `json:"workflowId"`
	// Workflow database identifier
	WorkflowDatabaseID string `json:"workflowDatabaseId"`
	// Issues associated with the step
	Issues *components.StepIssuesDto `json:"issues,omitempty"`
}

type ChatStepResponseDto struct {
	// Controls metadata for the chat step
	Controls components.ChatControlsMetadataResponseDto `json:"controls"`
	// Control values for the chat step
	ControlValues *components.ChatStepResponseDtoControlValues `json:"controlValues,omitempty"`
	// JSON Schema for variables, follows the JSON Schema standard
	Variables map[string]any `json:"variables"`
	// Unique identifier of the step
	StepID string `json:"stepId"`
	// Database identifier of the step
	ID string `json:"_id"`
	// Name of the step
	Name string `json:"name"`
	// Slug of the step
	Slug string `json:"slug"`
	// Type of the step
	Type components.StepTypeEnum `json:"type"`
	// Origin of the workflow
	Origin components.WorkflowOriginEnum `json:"origin"`
	// Workflow identifier
	WorkflowID string `json:"workflowId"`
	// Workflow database identifier
	WorkflowDatabaseID string `json:"workflowDatabaseId"`
	// Issues associated with the step
	Issues *components.StepIssuesDto `json:"issues,omitempty"`
}

type DelayStepResponseDto struct {
	// Controls metadata for the delay step
	Controls components.DelayControlsMetadataResponseDto `json:"controls"`
	// Control values for the delay step
	ControlValues *components.DelayStepResponseDtoControlValues `json:"controlValues,omitempty"`
	// JSON Schema for variables, follows the JSON Schema standard
	Variables map[string]any `json:"variables"`
	// Unique identifier of the step
	StepID string `json:"stepId"`
	// Database identifier of the step
	ID string `json:"_id"`
	// Name of the step
	Name string `json:"name"`
	// Slug of the step
	Slug string `json:"slug"`
	// Type of the step
	Type components.StepTypeEnum `json:"type"`
	// Origin of the workflow
	Origin components.WorkflowOriginEnum `json:"origin"`
	// Workflow identifier
	WorkflowID string `json:"workflowId"`
	// Workflow database identifier
	WorkflowDatabaseID string `json:"workflowDatabaseId"`
	// Issues associated with the step
	Issues *components.StepIssuesDto `json:"issues,omitempty"`
}

type DigestStepResponseDto struct {
	// Controls metadata for the digest step
	Controls components.DigestControlsMetadataResponseDto `json:"controls"`
	// Control values for the digest step
	ControlValues *components.DigestStepResponseDtoControlValues `json:"controlValues,omitempty"`
	// JSON Schema for variables, follows the JSON Schema standard
	Variables map[string]any `json:"variables"`
	// Unique identifier of the step
	StepID string `json:"stepId"`
	// Database identifier of the step
	ID string `json:"_id"`
	// Name of the step
	Name string `json:"name"`
	// Slug of the step
	Slug string `json:"slug"`
	// Type of the step
	Type components.StepTypeEnum `json:"type"`
	// Origin of the workflow
	Origin components.WorkflowOriginEnum `json:"origin"`
	// Workflow identifier
	WorkflowID string `json:"workflowId"`
	// Workflow database identifier
	WorkflowDatabaseID string `json:"workflowDatabaseId"`
	// Issues associated with the step
	Issues *components.StepIssuesDto `json:"issues,omitempty"`
}

type CustomStepResponseDto struct {
	// Controls metadata for the custom step
	Controls components.CustomControlsMetadataResponseDto `json:"controls"`
	// Control values for the custom step
	ControlValues *components.CustomStepResponseDtoControlValues `json:"controlValues,omitempty"`
	// JSON Schema for variables, follows the JSON Schema standard
	Variables map[string]any `json:"variables"`
	// Unique identifier of the step
	StepID string `json:"stepId"`
	// Database identifier of the step
	ID string `json:"_id"`
	// Name of the step
	Name string `json:"name"`
	// Slug of the step
	Slug string `json:"slug"`
	// Type of the step
	Type components.StepTypeEnum `json:"type"`
	// Origin of the workflow
	Origin components.WorkflowOriginEnum `json:"origin"`
	// Workflow identifier
	WorkflowID string `json:"workflowId"`
	// Workflow database identifier
	WorkflowDatabaseID string `json:"workflowDatabaseId"`
	// Issues associated with the step
	Issues *components.StepIssuesDto `json:"issues,omitempty"`
}
*/
